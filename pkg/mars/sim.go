package mars

import "fmt"

type SimulatorMode uint8

const (
	ICWS88 SimulatorMode = iota
	NOP94
	ICWS94
)

type SimulatorState uint8

const (
	Initialized SimulatorState = iota
	Running
	Complete
)

type Simulator interface {
	CoreSize() Address
	CycleCount() int
	AddWarrior(data *WarriorData) (Warrior, error)
	GetWarrior(wi int) Warrior
	SpawnWarrior(wi int, startOffset Address) error
	Run() []bool

	// RunCycle runs a full cyle of the living warriors, starting at s,warriorIndex.
	//
	// If s.cycleCount > s.maxCycles or there are no living warriors, no warrior
	// will be run, s.cycleCount will not be incremented, and the return value will
	// be 0. Otherwise s.cycleCount is incremented and
	RunCycle() int
	GetMem(a Address) Instruction
	Reset()

	WarriorLivingCount() int
	WarriorCount() int
}

type ReportingSimulator interface {
	Simulator
	AddReporter(r Reporter)
}

type reportSim struct {
	m          Address
	maxProcs   Address
	maxCycles  Address
	readLimit  Address
	writeLimit Address
	mem        []Instruction
	legacy     bool

	warriors           []*warrior
	reporters          []Reporter
	warriorIndex       int
	warriorCount       int
	warriorLivingCount int

	cycleCount Address
}

func NewSimulator(config SimulatorConfig) (Simulator, error) {
	return newReportSim(config)
}

func NewReportingSimulator(config SimulatorConfig) (ReportingSimulator, error) {
	return newReportSim(config)
}

func newReportSim(config SimulatorConfig) (*reportSim, error) {
	err := config.Validate()
	if err != nil {
		return nil, err
	}

	sim := &reportSim{
		m:          Address(config.CoreSize),
		maxProcs:   Address(config.Processes),
		maxCycles:  Address(config.Cycles),
		readLimit:  Address(config.ReadLimit),
		writeLimit: Address(config.WriteLimit),
		legacy:     config.Mode == ICWS88,
	}

	sim.mem = make([]Instruction, sim.m)

	return sim, nil
}

func (s *reportSim) CoreSize() Address {
	return s.m
}

func (s *reportSim) CycleCount() int {
	return int(s.cycleCount)
}

func (s *reportSim) AddReporter(r Reporter) {
	s.reporters = append(s.reporters, r)
}

func (s *reportSim) Report(report Report) {
	for _, r := range s.reporters {
		r.Report(report)
	}
}

func (s *reportSim) addressSigned(a Address) int {
	if a > (s.m / 2) {
		return -(int(s.m) - int(a))
	}
	return int(a)
}

func (s *reportSim) GetWarrior(i int) Warrior {
	if i > s.warriorCount {
		return nil
	}
	return s.warriors[i]
}

func (s *reportSim) AddWarrior(data *WarriorData) (Warrior, error) {
	return s.addWarrior(data)
}

func (s *reportSim) addWarrior(data *WarriorData) (*warrior, error) {
	w := &warrior{
		data: data.Copy(),
		sim:  s,
	}
	w.index = len(s.warriors)
	s.warriors = append(s.warriors, w)
	s.warriorCount += 1
	w.state = WarriorAdded

	return w, nil
}

func (s *reportSim) SpawnWarrior(wi int, startOffset Address) error {
	return s.spawnWarrior(wi, startOffset)
}

func (s *reportSim) spawnWarrior(wi int, startOffset Address) error {
	if wi > s.warriorCount {
		return fmt.Errorf("warrior index out of bounds")
	}
	w := s.warriors[wi]
	if w.state == WarriorAlive {
		return fmt.Errorf("warrior already spawned")
	}

	for i := Address(0); i < Address(len(w.data.Code)); i++ {
		s.mem[(startOffset+i)%s.m] = w.data.Code[i]
	}

	w.pq = newProcessQueue(s.maxProcs)
	w.pq.Push(startOffset + Address(w.data.Start))
	w.state = WarriorAlive
	s.warriorLivingCount += 1

	s.Report(Report{Type: WarriorSpawn, WarriorIndex: w.index, Address: startOffset})

	return nil
}

func (s *reportSim) WarriorCount() int {
	return s.warriorCount
}

func (s *reportSim) WarriorLivingCount() int {
	return s.warriorLivingCount
}

func (s *reportSim) RunCycle() int {
	if s.cycleCount >= s.maxCycles || s.warriorLivingCount < 1 {
		return 0
	}

	if s.warriorIndex == 0 {
		s.Report(Report{Type: CycleStart, Cycle: int(s.cycleCount)})
	}

	// step through s.warriors from s.warriorIndex to the end
	for i := s.warriorIndex; i < s.warriorCount; i++ {
		if s.warriors[i].state == WarriorAlive {
			// I don't like this, and this should never happen, but we will
			// silently reap any zombie warriors here that are 'alive' without
			// a process queue so we can continue and check the next ones.
			var err error
			pc, err := s.warriors[i].pq.Pop()
			if err != nil {
				s.warriors[i].state = WarriorDead
				continue
			}

			s.Report(Report{Type: WarriorTaskPop, Cycle: int(s.cycleCount), WarriorIndex: i, Address: pc})

			s.exec(pc, s.warriors[i])
			if s.warriors[i].pq.Len() == 0 {
				s.Report(Report{Type: WarriorTerminate, Cycle: int(s.cycleCount), WarriorIndex: i, Address: pc})
				s.warriors[i].state = WarriorDead
				s.warriorLivingCount--
			}
		}
	}

	s.Report(Report{Type: CycleEnd, Cycle: int(s.cycleCount)})

	s.warriorIndex = 0
	s.cycleCount++

	return s.warriorLivingCount
}

func (s *reportSim) readFold(pointer Address) Address {
	res := pointer % s.readLimit
	if res > (s.readLimit / 2) {
		res += (s.m - s.readLimit)
	}
	return res
}

func (s *reportSim) writeFold(pointer Address) Address {
	res := pointer % s.writeLimit
	if res > (s.writeLimit / 2) {
		res += (s.m - s.writeLimit)
	}
	return res
}

func (s *reportSim) exec(PC Address, w *warrior) {
	IR := s.mem[PC]

	// read and write limit folded pointers for A, B
	var RPA, WPA, RPB, WPB Address

	// instructions referenced by A, B
	var IRA, IRB Instruction

	// pointer to increment after IRA, IRB
	var PIP Address

	// prepare A indirect references and decrement or save increment pointer
	if IR.AMode != IMMEDIATE {
		RPA = s.readFold(IR.A)
		WPA = s.writeFold(IR.A)

		if IR.AMode == A_INDIRECT || IR.AMode == A_DECREMENT || IR.AMode == A_INCREMENT {
			if IR.AMode == A_DECREMENT {
				dptr := (PC + WPA) % s.m
				s.mem[dptr].A = (s.mem[dptr].A + s.m - 1) % s.m
				s.Report(Report{Type: WarriorDecrement, WarriorIndex: w.index, Address: dptr})
			}

			if IR.AMode == A_INCREMENT {
				PIP = (PC + WPA) % s.m
			}

			RPA = s.readFold(RPA + s.mem[(PC+RPA)%s.m].A)
			// not used, but should be updated here if it were to be
			// WPA = s.writeFold(WPA + s.mem[(PC+WPA)%s.m].A)
		}

		if IR.AMode == B_INDIRECT || IR.AMode == B_DECREMENT || IR.AMode == B_INCREMENT {
			if IR.AMode == B_DECREMENT {
				dptr := (PC + WPA) % s.m
				s.mem[dptr].B = (s.mem[dptr].B + s.m - 1) % s.m
				s.Report(Report{Type: WarriorDecrement, WarriorIndex: w.index, Address: dptr})
			}

			if IR.AMode == B_INCREMENT {
				PIP = (PC + WPA) % s.m
			}

			RPA = s.readFold(RPA + s.mem[(PC+RPA)%s.m].B)
			// not used, but should be updated here if it were to be
			// WPA = s.writeFold(WPA + s.mem[(PC+WPA)%s.m].B)
		}

	}

	// assign referenced value to IRA
	IRA = s.mem[(PC+RPA)%s.m]

	// do post-increments, if needed, after IRA has been assigned
	if IR.AMode == A_INCREMENT {
		s.mem[PIP].A = (s.mem[PIP].A + 1) % s.m
	}
	if IR.AMode == B_INCREMENT {
		s.mem[PIP].B = (s.mem[PIP].B + 1) % s.m
	}

	// prepare B indirect references and decrement or save increment pointer
	if IR.BMode != IMMEDIATE {
		RPB = s.readFold(IR.B)
		WPB = s.writeFold(IR.B)

		if IR.BMode == A_INDIRECT || IR.BMode == A_DECREMENT || IR.BMode == A_INCREMENT {
			if IR.BMode == A_DECREMENT {
				dptr := (PC + WPB) % s.m
				s.mem[dptr].A = (s.mem[dptr].A + s.m - 1) % s.m
			}

			if IR.BMode == A_INCREMENT {
				PIP = (PC + WPB) % s.m
			}

			RPB = s.readFold(RPB + s.mem[(PC+RPB)%s.m].A)
			WPB = s.writeFold(WPB + s.mem[(PC+WPB)%s.m].A)
		}

		if IR.BMode == B_INDIRECT || IR.BMode == B_DECREMENT || IR.BMode == B_INCREMENT {
			if IR.BMode == B_DECREMENT {
				dptr := (PC + WPB) % s.m
				s.mem[dptr].B = (s.mem[dptr].B + s.m - 1) % s.m
			}

			if IR.BMode == B_INCREMENT {
				PIP = (PC + WPB) % s.m
			}

			RPB = s.readFold(RPB + s.mem[(PC+RPB)%s.m].B)
			WPB = s.writeFold(WPB + s.mem[(PC+WPB)%s.m].B)
		}

	}

	// assign referenced value to IRB
	IRB = s.mem[(PC+RPB)%s.m]

	// do post-increments, if needed, after IRB has been assigned
	if IR.BMode == A_INCREMENT {
		s.mem[PIP].A = (s.mem[PIP].A + 1) % s.m
		s.Report(Report{Type: WarriorIncrement, WarriorIndex: w.index, Address: PIP})
	} else if IR.BMode == B_INCREMENT {
		s.mem[PIP].B = (s.mem[PIP].B + 1) % s.m
		s.Report(Report{Type: WarriorIncrement, WarriorIndex: w.index, Address: PIP})
	}

	WAB := (PC + WPB) % s.m
	RAB := (PC + RPA) % s.m

	switch IR.Op {
	case DAT:
		s.Report(Report{Type: WarriorTaskTerminate, WarriorIndex: w.index, Address: PC})
		return
	case MOV:
		s.mov(IR, IRA, WAB, PC, w)
		s.Report(Report{Type: WarriorWrite, WarriorIndex: w.index, Address: WAB})
	case ADD:
		s.add(IR, IRA, IRB, WAB, PC, w)
		s.Report(Report{Type: WarriorWrite, WarriorIndex: w.index, Address: WAB})
	case SUB:
		s.sub(IR, IRA, IRB, WAB, PC, w)
		s.Report(Report{Type: WarriorWrite, WarriorIndex: w.index, Address: WAB})
	case MUL:
		s.mul(IR, IRA, IRB, WAB, PC, w)
		s.Report(Report{Type: WarriorWrite, WarriorIndex: w.index, Address: WAB})
	case DIV:
		s.div(IR, IRA, IRB, WAB, PC, w)
		s.Report(Report{Type: WarriorWrite, WarriorIndex: w.index, Address: WAB})
	case MOD:
		s.mod(IR, IRA, IRB, WAB, PC, w)
		s.Report(Report{Type: WarriorWrite, WarriorIndex: w.index, Address: WAB})
	case JMP:
		w.pq.Push(RAB)
	case JMZ:
		s.jmz(IR, IRB, RAB, PC, w)
	case JMN:
		s.jmn(IR, IRB, RAB, PC, w)
	case DJN:
		s.djn(IR, IRB, RAB, WAB, PC, w)
		s.Report(Report{Type: WarriorDecrement, WarriorIndex: w.index, Address: WAB})
	case CMP:
		fallthrough
	case SEQ:
		s.cmp(IR, IRA, IRB, PC, w)
		s.Report(Report{Type: WarriorRead, WarriorIndex: w.index, Address: (PC + RPA) % s.m})
		s.Report(Report{Type: WarriorRead, WarriorIndex: w.index, Address: (PC + RPB) % s.m})
	case SLT:
		s.slt(IR, IRA, IRB, PC, w)
		s.Report(Report{Type: WarriorRead, WarriorIndex: w.index, Address: (PC + RPA) % s.m})
		s.Report(Report{Type: WarriorRead, WarriorIndex: w.index, Address: (PC + RPB) % s.m})
	case SNE:
		s.sne(IR, IRA, IRB, PC, w)
		s.Report(Report{Type: WarriorRead, WarriorIndex: w.index, Address: (PC + RPA) % s.m})
		s.Report(Report{Type: WarriorRead, WarriorIndex: w.index, Address: (PC + RPB) % s.m})
	case SPL:
		w.pq.Push((PC + 1) % s.m)
		w.pq.Push(RAB)
	case NOP:
		w.pq.Push((PC + 1) % s.m)
	}
}

// Run runs the simulator until the max cycles are reached, one warrior
// remains in a battle with more than one warrior, or the only warrior
// dies in a single warrior battle
func (s *reportSim) Run() []bool {
	nWarriors := len(s.warriors)

	// if no warriors are loaded, return nil
	if nWarriors == 0 {
		return nil
	}

	// run until simulation
	for s.cycleCount < s.maxCycles {
		aliveCount := s.RunCycle()

		if nWarriors == 1 && aliveCount == 0 {
			break
		} else if nWarriors > 1 && aliveCount == 1 {
			break
		}
	}

	// collect and return results
	result := make([]bool, nWarriors)
	for i, warrior := range s.warriors {
		result[i] = warrior.Alive()
	}
	return result
}

func (s *reportSim) GetMem(a Address) Instruction {
	return s.mem[a%s.m]
}

func (s *reportSim) Reset() {
	s.Report(Report{Type: SimReset})

	for _, warrior := range s.warriors {
		warrior.state = WarriorAdded
	}
	s.mem = make([]Instruction, s.m)
	s.cycleCount = 0
	s.warriorLivingCount = 0
}

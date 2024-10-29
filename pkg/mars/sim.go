package mars

type Simulator struct {
	m          Address
	maxProcs   Address
	maxCycles  Address
	readLimit  Address
	writeLimit Address
	mem        []Instruction
	legacy     bool

	warriors  []*Warrior
	reporters []Reporter
	// state    WarriorState

	cycleCount Address
}

type SimulatorConfig struct {
	Mode       SimulatorMode
	CoreSize   Address
	Processes  Address
	Cycles     Address
	ReadLimit  Address
	WriteLimit Address
	Length     Address
	Distance   Address
}

func StandardConfig() SimulatorConfig {
	return SimulatorConfig{
		Mode:       ICWS88,
		CoreSize:   8000,
		Processes:  8000,
		Cycles:     80000,
		ReadLimit:  8000,
		WriteLimit: 8000,
		Length:     100,
		Distance:   100,
	}
}

func BasicConfig(mode SimulatorMode, coreSize, processes, cycles, length Address) SimulatorConfig {
	out := SimulatorConfig{
		Mode:       mode,
		CoreSize:   coreSize,
		Processes:  processes,
		Cycles:     cycles,
		ReadLimit:  coreSize,
		WriteLimit: coreSize,
		Length:     length,
		Distance:   length,
	}
	return out
}

// func NewSimulator(coreSize, maxProcs, maxCycles, readLimit, writeLimit Address, legacy bool) *Simulator {
func NewSimulator(config SimulatorConfig) *Simulator {
	sim := &Simulator{
		m:          Address(config.CoreSize),
		maxProcs:   Address(config.Processes),
		maxCycles:  Address(config.Cycles),
		readLimit:  Address(config.ReadLimit),
		writeLimit: Address(config.WriteLimit),
		legacy:     config.Mode == ICWS88,
	}
	sim.mem = make([]Instruction, sim.m)
	return sim
}

func (s *Simulator) CycleCount() Address {
	return s.cycleCount
}

func (s *Simulator) addressSigned(a Address) int {
	if a > (s.m / 2) {
		return -(int(s.m) - int(a))
	}
	return int(a)
}

func (s *Simulator) SpawnWarrior(data *WarriorData, startOffset Address) (*Warrior, error) {

	w := &Warrior{
		data: data.Copy(),
		sim:  s,
	}

	for i := Address(0); i < Address(len(w.data.Code)); i++ {
		s.mem[(startOffset+i)%s.m] = w.data.Code[i]
	}

	s.warriors = append(s.warriors, w)
	w.index = len(s.warriors)
	w.pq = NewProcessQueue(s.maxProcs)
	w.pq.Push(startOffset + Address(data.Start))
	w.state = ALIVE

	for _, r := range s.reporters {
		r.WarriorSpawn(len(s.warriors), startOffset, startOffset+Address(w.data.Start))
	}

	return w, nil
}

// RunTurn runs a cycle, executing each living warrior and returns
// the number of living warriors at the end of the cycle
func (s *Simulator) RunCycle() int {
	nAlive := 0

	for _, warrior := range s.warriors {
		if warrior.state != ALIVE {
			continue
		}

		pc, err := warrior.pq.Pop()
		if err != nil {
			warrior.state = DEAD
			continue
		}

		s.exec(pc, warrior.pq)
		if warrior.pq.Len() > 0 {
			nAlive++
		} else {
			warrior.state = DEAD
		}
	}

	s.cycleCount++

	return nAlive
}

func (s *Simulator) readFold(pointer Address) Address {
	res := pointer % s.readLimit
	if res < (s.readLimit / 2) {
		res += (s.m - s.readLimit)
	}
	return res
}

func (s *Simulator) writeFold(pointer Address) Address {
	res := pointer % s.writeLimit
	if res < (s.writeLimit / 2) {
		res += (s.m - s.writeLimit)
	}
	return res
}

func (s *Simulator) exec(PC Address, pq *processQueue) {
	IR := s.mem[PC]

	// read and write limit folded pointers for A, B
	var RPA, WPA, RPB, WPB Address

	// instructions referenced by A, B
	var IRA, IRB Instruction

	if IR.AMode == IMMEDIATE {
		IRA = IR
	} else {
		RPA = s.readFold(IR.A)
		WPA = s.writeFold(IR.A)

		if IR.AMode == B_INDIRECT || IR.AMode == B_DECREMENT {
			if IR.AMode == B_DECREMENT {
				dptr := (PC + WPA) % s.m
				s.mem[dptr].B = (s.mem[dptr].B + s.m - 1) % s.m
			}
			RPA = s.readFold(RPA + s.mem[(PC+RPA)%s.m].B)
			// WPA = s.writeFold(WPA + s.mem[(PC+WPA)%s.m].B)
		}
		IRA = s.mem[(PC+RPA)%s.m]
	}

	if IR.BMode == IMMEDIATE {
		IRB = IR

	} else {
		RPB = s.readFold(IR.B)
		WPB = s.writeFold(IR.B)

		if IR.BMode == B_INDIRECT || IR.BMode == B_DECREMENT {
			if IR.BMode == B_DECREMENT {
				dptr := (PC + WPB) % s.m
				s.mem[dptr].B = (s.mem[dptr].B + s.m - 1) % s.m
			}
			RPB = s.readFold(RPB + s.mem[(PC+RPB)%s.m].B)
			WPB = s.writeFold(WPB + s.mem[(PC+WPB)%s.m].B)
		}
		IRB = s.mem[(PC+RPB)%s.m]
	}

	WAB := (PC + WPB) % s.m
	RAB := (PC + RPA) % s.m

	switch IR.Op {
	case DAT:
		return
	case MOV:
		s.mov(IR, IRA, WAB, PC, pq)
	case ADD:
		s.add(IR, IRA, IRB, WAB, PC, pq)
	case SUB:
		s.sub(IR, IRA, IRB, WAB, PC, pq)
	case MUL:
		s.mul(IR, IRA, IRB, WAB, PC, pq)
	case DIV:
		s.div(IR, IRA, IRB, WAB, PC, pq)
	case MOD:
		s.mod(IR, IRA, IRB, WAB, PC, pq)
	case JMP:
		pq.Push(RAB)
	case JMZ:
		s.jmz(IR, IRB, RAB, PC, pq)
	case JMN:
		s.jmn(IR, IRB, RAB, PC, pq)
	case DJN:
		s.djn(IR, IRB, RAB, WAB, PC, pq)
	case CMP:
		fallthrough
	case SEQ:
		s.cmp(IR, IRA, IRB, PC, pq)
	case SLT:
		s.slt(IR, IRA, IRB, PC, pq)
	case SPL:
		pq.Push((PC + 1) % s.m)
		pq.Push(RAB)
	case NOP:
		pq.Push((PC + 1) % s.m)
	}
}

// Run runs the simulator until the max cycles are reached, one warrior
// remains in a battle with more than one warrior, or the only warrior
// dies in a single warrior battle
func (s *Simulator) Run() []bool {
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

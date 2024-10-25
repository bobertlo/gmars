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
}

func NewSimulator(coreSize, maxProcs, maxCycles, readLimit, writeLimit Address, legacy bool) *Simulator {
	sim := &Simulator{
		m:          coreSize,
		maxProcs:   maxProcs,
		maxCycles:  maxCycles,
		readLimit:  readLimit,
		writeLimit: writeLimit,
		legacy:     legacy,
	}
	sim.mem = make([]Instruction, coreSize)
	return sim
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

func (s *Simulator) run_turn() error {
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
	}
	return nil
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
			WPA = s.writeFold(WPA + s.mem[(PC+WPA)%s.m].B)
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

	switch IR.Op {
	case DAT:
		return
	case MOV:
		s.mov(IR, IRA, (WPB+PC)%s.m, PC, pq)
	case ADD:
		s.add(IR, IRA, IRB, (WPB+PC)%s.m, PC, pq)
	case JMP:
		pq.Push((PC + RPA) % s.m)
	}
}

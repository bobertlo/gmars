package mars

import "fmt"

type MARS struct {
	coreSize   Address
	maxProcs   Address
	maxCycles  Address
	readLimit  Address
	writeLimit Address
	mem        []Instruction
	legacy     bool

	warriors []*Warrior
	// state    WarriorState
}

func NewMARS(coreSize, maxProcs, maxCycles, readLimit, writeLimit Address, legacy bool) *MARS {
	sim := &MARS{
		coreSize:   coreSize,
		maxProcs:   maxProcs,
		maxCycles:  maxCycles,
		readLimit:  readLimit,
		writeLimit: writeLimit,
		legacy:     legacy,
	}
	sim.mem = make([]Instruction, coreSize)
	return sim
}

func (s *MARS) addressSigned(a Address) int {
	if a > (s.coreSize / 2) {
		return -(int(s.coreSize) - int(a))
	}
	return int(a)
}

func (s *MARS) AddWarrior(data *WarriorData, startOffset Address) (*Warrior, error) {
	w := &Warrior{
		data: data.Copy(),
		sim:  s,
	}

	for i := Address(0); i < Address(len(w.data.Code)); i++ {
		s.mem[(startOffset+i)%s.coreSize] = w.data.Code[i]
	}

	s.warriors = append(s.warriors, w)
	w.index = len(s.warriors)

	return w, nil
}

func (s *MARS) step() {
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

}

func (s *MARS) readFold(pointer Address) Address {
	res := pointer % s.readLimit
	if res < (s.readLimit / 2) {
		res += (s.coreSize - s.readLimit)
	}
	return res
}

func (s *MARS) writeFold(pointer Address) Address {
	res := pointer % s.writeLimit
	if res < (s.writeLimit / 2) {
		res += (s.coreSize - s.writeLimit)
	}
	return res
}

func (s *MARS) exec(PC Address, pq *processQueue) {
	IR := s.mem[PC]

	// read and write limit folded pointers for A, B
	var RPA, WPA, RPB, WPB Address

	// instructions referenced by A, B
	var IRA, IRB Instruction

	if IR.AMode != IMMEDIATE {
		RPA = s.readFold(IR.A)
		WPA = s.writeFold(IR.A)

		if IR.AMode == DIRECT {
			RPA = s.readFold(RPA + s.mem[(PC+RPA)%s.coreSize].A)
			WPA = s.writeFold(WPA + s.mem[(PC+WPA)%s.coreSize].A)
		}
		if IR.AMode == B_INDIRECT || IR.AMode == B_DECREMENT {
			if IR.AMode == B_DECREMENT {
				dptr := (PC + WPA) % s.coreSize
				s.mem[dptr].B = (s.mem[dptr].B + s.coreSize - 1) % s.coreSize
			}
			RPA = s.readFold(RPA + s.mem[(PC+RPA)%s.coreSize].B)
			WPA = s.writeFold(WPA + s.mem[(PC+WPA)%s.coreSize].B)
		}
	}
	IRA = s.mem[(PC+RPA)%s.coreSize]

	if IR.BMode != IMMEDIATE {
		RPB = s.readFold(IR.B)
		WPB = s.writeFold(IR.B)

		if IR.BMode == DIRECT {
			RPB = s.readFold(RPB + s.mem[(PC+RPB)%s.coreSize].A)
			WPB = s.writeFold(WPB + s.mem[(PC+WPB)%s.coreSize].A)
		}
		if IR.BMode == B_INDIRECT || IR.BMode == B_DECREMENT {
			if IR.BMode == B_DECREMENT {
				dptr := (PC + WPB) % s.coreSize
				s.mem[dptr].B = (s.mem[dptr].B + s.coreSize - 1) % s.coreSize
			}
			RPB = s.readFold(RPB + s.mem[(PC+RPB)%s.coreSize].B)
			WPB = s.writeFold(WPB + s.mem[(PC+WPB)%s.coreSize].B)
		}

	}
	IRB = s.mem[(PC+RPB)%s.coreSize]

	if IR.BMode != IMMEDIATE {
		RPB = s.readFold(IR.B)
		WPB = s.writeFold(IR.B)
	}

	switch IR.Op {
	case DAT:
		return
	case MOV:
		s.mov(IR, IRA, (WPB+PC)%s.coreSize, PC, pq)
	case ADD:
		s.add(IR, IRA, IRB, (WPB+PC)%s.coreSize, PC, pq)
	case JMP:
		pq.Push(RPA)
	}

	fmt.Println(IRB)
}

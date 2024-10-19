package mars

func (s *MARS) mov(IR, IRA Instruction, dest, PC Address, pq *processQueue) {
	switch IR.OpMode {
	case AB:
		s.mem[dest].B = IR.A
	case I:
		s.mem[dest] = IRA
	}
	pq.Push((PC + 1) % s.coreSize)
}

func (s *MARS) add(IR Instruction, RPA, dest, PC Address, pq *processQueue) {
	switch IR.OpMode {
	case A:
		s.mem[dest].B = IR.A
	case B:
		s.mem[dest].B = IR.B
	case F:
		s.mem[dest] = s.mem[RPA]
	}
	pq.Push((PC + 1) % s.coreSize)
}

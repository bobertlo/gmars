package mars

func (s *MARS) mov(IR, IRA Instruction, WAB, PC Address, pq *processQueue) {
	switch IR.OpMode {
	case A:
		s.mem[WAB].A = IRA.A
	case B:
		s.mem[WAB].B = IRA.B
	case AB:
		s.mem[WAB].B = IRA.A
	case BA:
		s.mem[WAB].A = IRA.B
	case F:
		s.mem[WAB].A = IRA.A
		s.mem[WAB].B = IRA.B
	case X:
		s.mem[WAB].B = IRA.A
		s.mem[WAB].A = IRA.B
	case I:
		s.mem[WAB] = IRA
	}
	pq.Push((PC + 1) % s.coreSize)
}

func (s *MARS) add(IR, IRA, IRB Instruction, WAB, PC Address, pq *processQueue) {
	switch IR.OpMode {
	case A:
		s.mem[WAB].A = (IRB.A + IRA.A) % s.coreSize
	case B:
		s.mem[WAB].B = (IRB.B + IRA.B) % s.coreSize
	case AB:
		s.mem[WAB].B = (IRB.A + IRA.B) % s.coreSize
	case BA:
		s.mem[WAB].A = (IRB.B + IRA.A) % s.coreSize
	case I:
		fallthrough
	case F:
		s.mem[WAB].A = (IRB.A + IRA.A) % s.coreSize
		s.mem[WAB].B = (IRB.B + IRA.B) % s.coreSize
	case X:
		s.mem[WAB].A = (IRB.B + IRA.A) % s.coreSize
		s.mem[WAB].B = (IRB.A + IRA.B) % s.coreSize
	}
	pq.Push((PC + 1) % s.coreSize)
}

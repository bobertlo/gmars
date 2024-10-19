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
	pq.Push((PC + 1) % s.m)
}

func (s *MARS) add(IR, IRA, IRB Instruction, WAB, PC Address, pq *processQueue) {
	switch IR.OpMode {
	case A:
		s.mem[WAB].A = (IRB.A + IRA.A) % s.m
	case B:
		s.mem[WAB].B = (IRB.B + IRA.B) % s.m
	case AB:
		s.mem[WAB].B = (IRB.A + IRA.B) % s.m
	case BA:
		s.mem[WAB].A = (IRB.B + IRA.A) % s.m
	case I:
		fallthrough
	case F:
		s.mem[WAB].A = (IRB.A + IRA.A) % s.m
		s.mem[WAB].B = (IRB.B + IRA.B) % s.m
	case X:
		s.mem[WAB].A = (IRB.B + IRA.A) % s.m
		s.mem[WAB].B = (IRB.A + IRA.B) % s.m
	}
	pq.Push((PC + 1) % s.m)
}

func (s *MARS) sub(IR, IRA, IRB Instruction, WAB, PC Address, pq *processQueue) {
	switch IR.OpMode {
	case A:
		s.mem[WAB].A = (IRB.A + (s.m - IRA.A)) % s.m
	case B:
		s.mem[WAB].B = (IRB.B + (s.m - IRA.B)) % s.m
	case AB:
		s.mem[WAB].B = (IRB.A + (s.m - IRA.B)) % s.m
	case BA:
		s.mem[WAB].A = (IRB.B + (s.m - IRA.A)) % s.m
	case I:
		fallthrough
	case F:
		s.mem[WAB].A = (IRB.A + (s.m - IRA.A)) % s.m
		s.mem[WAB].B = (IRB.B + (s.m - IRA.B)) % s.m
	case X:
		s.mem[WAB].A = (IRB.B + (s.m - IRA.A)) % s.m
		s.mem[WAB].B = (IRB.A + (s.m - IRA.B)) % s.m
	}
	pq.Push((PC + 1) % s.m)
}

func (s *MARS) mul(IR, IRA, IRB Instruction, WAB, PC Address, pq *processQueue) {
	switch IR.OpMode {
	case A:
		s.mem[WAB].A = (IRB.A * IRA.A) % s.m
	case B:
		s.mem[WAB].B = (IRB.B * IRA.B) % s.m
	case AB:
		s.mem[WAB].B = (IRB.A * IRA.B) % s.m
	case BA:
		s.mem[WAB].A = (IRB.B * IRA.A) % s.m
	case I:
		fallthrough
	case F:
		s.mem[WAB].A = (IRB.A * IRA.A) % s.m
		s.mem[WAB].B = (IRB.B * IRA.B) % s.m
	case X:
		s.mem[WAB].A = (IRB.B * IRA.A) % s.m
		s.mem[WAB].B = (IRB.A * IRA.B) % s.m
	}
	pq.Push((PC + 1) % s.m)
}

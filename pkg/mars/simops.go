package mars

func (s *Simulator) mov(IR, IRA Instruction, WAB, PC Address, pq *processQueue) {
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

func (s *Simulator) add(IR, IRA, IRB Instruction, WAB, PC Address, pq *processQueue) {
	switch IR.OpMode {
	case A:
		s.mem[WAB].A = (IRB.A + IRA.A) % s.m
	case B:
		s.mem[WAB].B = (IRB.B + IRA.B) % s.m
	case AB:
		s.mem[WAB].B = (IRB.B + IRA.A) % s.m
	case BA:
		s.mem[WAB].A = (IRB.A + IRA.B) % s.m
	case I:
		fallthrough
	case F:
		s.mem[WAB].A = (IRB.A + IRA.A) % s.m
		s.mem[WAB].B = (IRB.B + IRA.B) % s.m
	case X:
		s.mem[WAB].A = (IRB.A + IRA.B) % s.m
		s.mem[WAB].B = (IRB.B + IRA.A) % s.m
	}
	pq.Push((PC + 1) % s.m)
}

func (s *Simulator) sub(IR, IRA, IRB Instruction, WAB, PC Address, pq *processQueue) {
	switch IR.OpMode {
	case A:
		s.mem[WAB].A = (IRB.A + (s.m - IRA.A)) % s.m
	case B:
		s.mem[WAB].B = (IRB.B + (s.m - IRA.B)) % s.m
	case AB:
		s.mem[WAB].B = (IRB.B + (s.m - IRA.A)) % s.m
	case BA:
		s.mem[WAB].A = (IRB.A + (s.m - IRA.B)) % s.m
	case I:
		fallthrough
	case F:
		s.mem[WAB].A = (IRB.A + (s.m - IRA.A)) % s.m
		s.mem[WAB].B = (IRB.B + (s.m - IRA.B)) % s.m
	case X:
		s.mem[WAB].A = (IRB.A + (s.m - IRA.B)) % s.m
		s.mem[WAB].B = (IRB.B + (s.m - IRA.A)) % s.m
	}
	pq.Push((PC + 1) % s.m)
}

func (s *Simulator) mul(IR, IRA, IRB Instruction, WAB, PC Address, pq *processQueue) {
	switch IR.OpMode {
	case A:
		s.mem[WAB].A = (IRB.A * IRA.A) % s.m
	case B:
		s.mem[WAB].B = (IRB.B * IRA.B) % s.m
	case AB:
		s.mem[WAB].B = (IRB.B * IRA.A) % s.m
	case BA:
		s.mem[WAB].A = (IRB.A * IRA.B) % s.m
	case I:
		fallthrough
	case F:
		s.mem[WAB].A = (IRB.A * IRA.A) % s.m
		s.mem[WAB].B = (IRB.B * IRA.B) % s.m
	case X:
		s.mem[WAB].A = (IRB.A * IRA.B) % s.m
		s.mem[WAB].B = (IRB.B * IRA.A) % s.m
	}
	pq.Push((PC + 1) % s.m)
}

func (s *Simulator) div(IR, IRA, IRB Instruction, WAB, PC Address, pq *processQueue) {
	switch IR.OpMode {
	case A:
		if IRA.A != 0 {
			s.mem[WAB].A = IRB.A / IRA.A
		} else {
			return
		}
	case B:
		if IRA.B != 0 {
			s.mem[WAB].B = IRB.B / IRA.B
		} else {
			return
		}
	case AB:
		if IRA.A != 0 {
			s.mem[WAB].B = IRB.B / IRA.A
		} else {
			return
		}
	case BA:
		if IRA.B != 0 {
			s.mem[WAB].A = IRB.A / IRA.B
		} else {
			return
		}
	case F:
		fallthrough
	case I:
		if IRA.A != 0 {
			s.mem[WAB].A = IRB.A / IRA.A
		}
		if IRA.B != 0 {
			s.mem[WAB].B = IRB.B / IRA.B
		}
		if IRA.A == 0 || IRA.B == 0 {
			return
		}
	case X:
		if IRA.A != 0 {
			s.mem[WAB].B = IRB.B / IRA.A
		}
		if IRA.B != 0 {
			s.mem[WAB].A = IRB.A / IRA.B
		}
		if IRA.A == 0 || IRA.B == 0 {
			return
		}
	}
	pq.Push((PC + 1) % s.m)
}

func (s *Simulator) jmz(IR, IRB Instruction, RAB, PC Address, pq *processQueue) {
	switch IR.OpMode {
	case A:
		fallthrough
	case BA:
		if IRB.A == 0 {
			pq.Push(RAB)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	case B:
		fallthrough
	case AB:
		if IRB.B == 0 {
			pq.Push(RAB)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	case F:
		fallthrough
	case X:
		fallthrough
	case I:
		if IRB.A == 0 && IRB.B == 0 {
			pq.Push(RAB % s.m)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	}
}

func (s *Simulator) jmn(IR, IRB Instruction, RAB, PC Address, pq *processQueue) {
	switch IR.OpMode {
	case A:
		fallthrough
	case BA:
		if IRB.A != 0 {
			pq.Push(RAB)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	case B:
		fallthrough
	case AB:
		if IRB.B != 0 {
			pq.Push(RAB)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	case F:
		fallthrough
	case X:
		fallthrough
	case I:
		if IRB.A != 0 || IRB.B != 0 {
			pq.Push(RAB)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	}
}

func (s *Simulator) djn(IR, IRB Instruction, RAB, WAB, PC Address, pq *processQueue) {
	switch IR.OpMode {
	case A:
		fallthrough
	case BA:
		s.mem[WAB].A = (s.mem[WAB].A + s.m - 1) % s.m
		IRB.A -= 1
		if IRB.A != 0 {
			pq.Push(RAB)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	case B:
		fallthrough
	case AB:
		s.mem[WAB].B = (s.mem[WAB].B + s.m - 1) % s.m
		IRB.B -= 1
		if IRB.B != 0 {
			pq.Push(RAB)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	case F:
		fallthrough
	case X:
		fallthrough
	case I:
		s.mem[WAB].A = (s.mem[WAB].A + s.m - 1) % s.m
		IRB.A -= 1
		s.mem[WAB].B = (s.mem[WAB].B + s.m - 1) % s.m
		IRB.B -= 1
		if IRB.B != 0 || IRB.A != 0 {
			pq.Push(RAB)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	}
}

func (s *Simulator) cmp(IR, IRA, IRB Instruction, PC Address, pq *processQueue) {
	switch IR.OpMode {
	case A:
		if IRA.A == IRB.A {
			pq.Push((PC + 2) % s.m)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	case B:
		if IRA.B == IRB.B {
			pq.Push((PC + 2) % s.m)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	case AB:
		if IRA.A == IRB.B {
			pq.Push((PC + 2) % s.m)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	case BA:
		if IRA.B == IRB.A {
			pq.Push((PC + 2) % s.m)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	case F:
		if IRA.A == IRB.A && IRA.B == IRB.B {
			pq.Push((PC + 2) % s.m)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	case X:
		if IRA.A == IRB.B && IRA.B == IRB.A {
			pq.Push((PC + 2) % s.m)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	case I:
		if IRA.Op == IRB.Op && IRA.OpMode == IRB.OpMode &&
			IRA.AMode == IRB.AMode && IRA.A == IRB.A &&
			IRA.BMode == IRB.BMode && IRA.B == IRB.B {
			pq.Push((PC + 2) % s.m)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	}
}

func (s *Simulator) slt(IR, IRA, IRB Instruction, PC Address, pq *processQueue) {
	switch IR.OpMode {
	case A:
		if IRA.A < IRB.A {
			pq.Push((PC + 2) % s.m)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	case B:
		if IRA.B < IRB.B {
			pq.Push((PC + 2) % s.m)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	case AB:
		if IRA.A < IRB.B {
			pq.Push((PC + 2) % s.m)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	case BA:
		if IRA.B < IRB.A {
			pq.Push((PC + 2) % s.m)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	case F:
		fallthrough
	case I:
		if IRA.A < IRB.A && IRA.B < IRB.B {
			pq.Push((PC + 2) % s.m)
		} else {
			pq.Push((PC + 1) % s.m)
		}

	case X:
		if IRA.A < IRB.B && IRA.B < IRB.A {
			pq.Push((PC + 2) % s.m)
		} else {
			pq.Push((PC + 1) % s.m)
		}
	}
}

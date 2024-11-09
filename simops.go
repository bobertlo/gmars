package gmars

func (s *reportSim) mov(IR, IRA Instruction, WAB, PC Address, w *warrior) {
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
	nextPC := (PC + 1) % s.m
	s.Report(Report{Type: WarriorTaskPush, WarriorIndex: w.index, Address: nextPC})
	w.pq.Push(nextPC)
}

func (s *reportSim) add(IR, IRA, IRB Instruction, WAB, PC Address, w *warrior) {
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
	nextPC := (PC + 1) % s.m
	s.Report(Report{Type: WarriorTaskPush, WarriorIndex: w.index, Address: nextPC})
	w.pq.Push(nextPC)
}

func (s *reportSim) sub(IR, IRA, IRB Instruction, WAB, PC Address, w *warrior) {
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
	nextPC := (PC + 1) % s.m
	s.Report(Report{Type: WarriorTaskPush, WarriorIndex: w.index, Address: nextPC})
	w.pq.Push(nextPC)
}

func (s *reportSim) mul(IR, IRA, IRB Instruction, WAB, PC Address, w *warrior) {
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
	nextPC := (PC + 1) % s.m
	s.Report(Report{Type: WarriorTaskPush, WarriorIndex: w.index, Address: nextPC})
	w.pq.Push(nextPC)
}

func (s *reportSim) div(IR, IRA, IRB Instruction, WAB, PC Address, w *warrior) {
	switch IR.OpMode {
	case A:
		if IRA.A != 0 {
			s.mem[WAB].A = IRB.A / IRA.A
		} else {
			s.Report(Report{Type: WarriorTaskTerminate, WarriorIndex: w.index, Address: PC})
			return
		}
	case B:
		if IRA.B != 0 {
			s.mem[WAB].B = IRB.B / IRA.B
		} else {
			s.Report(Report{Type: WarriorTaskTerminate, WarriorIndex: w.index, Address: PC})
			return
		}
	case AB:
		if IRA.A != 0 {
			s.mem[WAB].B = IRB.B / IRA.A
		} else {
			s.Report(Report{Type: WarriorTaskTerminate, WarriorIndex: w.index, Address: PC})
			return
		}
	case BA:
		if IRA.B != 0 {
			s.mem[WAB].A = IRB.A / IRA.B
		} else {
			s.Report(Report{Type: WarriorTaskTerminate, WarriorIndex: w.index, Address: PC})
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
			s.Report(Report{Type: WarriorTaskTerminate, WarriorIndex: w.index, Address: PC})
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
			s.Report(Report{Type: WarriorTaskTerminate, WarriorIndex: w.index, Address: PC})
			return
		}
	}
	nextPC := (PC + 1) % s.m
	s.Report(Report{Type: WarriorTaskPush, WarriorIndex: w.index, Address: nextPC})
	w.pq.Push(nextPC)
}

func (s *reportSim) mod(IR, IRA, IRB Instruction, WAB, PC Address, w *warrior) {
	switch IR.OpMode {
	case A:
		if IRA.A != 0 {
			s.mem[WAB].A = IRB.A % IRA.A
		} else {
			s.Report(Report{Type: WarriorTaskTerminate, WarriorIndex: w.index, Address: PC})
			return
		}
	case B:
		if IRA.B != 0 {
			s.mem[WAB].B = IRB.B % IRA.B
		} else {
			s.Report(Report{Type: WarriorTaskTerminate, WarriorIndex: w.index, Address: PC})
			return
		}
	case AB:
		if IRA.A != 0 {
			s.mem[WAB].B = IRB.B % IRA.A
		} else {
			s.Report(Report{Type: WarriorTaskTerminate, WarriorIndex: w.index, Address: PC})
			return
		}
	case BA:
		if IRA.B != 0 {
			s.mem[WAB].A = IRB.A % IRA.B
		} else {
			s.Report(Report{Type: WarriorTaskTerminate, WarriorIndex: w.index, Address: PC})
			return
		}
	case F:
		fallthrough
	case I:
		if IRA.A != 0 {
			s.mem[WAB].A = IRB.A % IRA.A
		}
		if IRA.B != 0 {
			s.mem[WAB].B = IRB.B % IRA.B
		}
		if IRA.A == 0 || IRA.B == 0 {
			s.Report(Report{Type: WarriorTaskTerminate, WarriorIndex: w.index, Address: PC})
			return
		}
	case X:
		if IRA.A != 0 {
			s.mem[WAB].B = IRB.B % IRA.A
		}
		if IRA.B != 0 {
			s.mem[WAB].A = IRB.A % IRA.B
		}
		if IRA.A == 0 || IRA.B == 0 {
			s.Report(Report{Type: WarriorTaskTerminate, WarriorIndex: w.index, Address: PC})
			return
		}
	}
	nextPC := (PC + 1) % s.m
	s.Report(Report{Type: WarriorTaskPush, WarriorIndex: w.index, Address: nextPC})
	w.pq.Push(nextPC)
}

func (s *reportSim) jmz(IR, IRB Instruction, RAB, PC Address, w *warrior) {
	switch IR.OpMode {
	case A:
		fallthrough
	case BA:
		if IRB.A == 0 {
			w.pq.Push(RAB)
		} else {
			w.pq.Push((PC + 1) % s.m)
		}
	case B:
		fallthrough
	case AB:
		if IRB.B == 0 {
			w.pq.Push(RAB)
		} else {
			w.pq.Push((PC + 1) % s.m)
		}
	case F:
		fallthrough
	case X:
		fallthrough
	case I:
		if IRB.A == 0 && IRB.B == 0 {
			w.pq.Push(RAB)
		} else {
			w.pq.Push((PC + 1) % s.m)
		}
	}
}

func (s *reportSim) jmn(IR, IRB Instruction, RAB, PC Address, w *warrior) {
	nextPC := (PC + 1) % s.m
	switch IR.OpMode {
	case A:
		fallthrough
	case BA:
		if IRB.A != 0 {
			nextPC = RAB
		}
	case B:
		fallthrough
	case AB:
		if IRB.B != 0 {
			nextPC = RAB
		}
	case F:
		fallthrough
	case X:
		fallthrough
	case I:
		if IRB.A != 0 || IRB.B != 0 {
			nextPC = RAB
		}
	default:
		return
	}

	s.Report(Report{Type: WarriorTaskPush, WarriorIndex: w.index, Address: nextPC})
	w.pq.Push(nextPC)
}

func (s *reportSim) djn(IR, IRB Instruction, RAB, WAB, PC Address, w *warrior) {
	nextPC := (PC + 1) % s.m

	switch IR.OpMode {
	case A:
		fallthrough
	case BA:
		s.mem[WAB].A = (s.mem[WAB].A + s.m - 1) % s.m
		IRB.A -= 1
		if IRB.A != 0 {
			nextPC = RAB
		}
	case B:
		fallthrough
	case AB:
		s.mem[WAB].B = (s.mem[WAB].B + s.m - 1) % s.m
		IRB.B -= 1
		if IRB.B != 0 {
			nextPC = RAB
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
			nextPC = RAB
		}
	}
	s.Report(Report{Type: WarriorTaskPush, WarriorIndex: w.index, Address: nextPC})
	w.pq.Push(nextPC)
}

func (s *reportSim) cmp(IR, IRA, IRB Instruction, PC Address, w *warrior) {
	nextPC := (PC + 1) % s.m
	switch IR.OpMode {
	case A:
		if IRA.A == IRB.A {
			nextPC = (PC + 2) % s.m
		}
	case B:
		if IRA.B == IRB.B {
			nextPC = (PC + 2) % s.m
		}
	case AB:
		if IRA.A == IRB.B {
			nextPC = (PC + 2) % s.m
		}
	case BA:
		if IRA.B == IRB.A {
			nextPC = (PC + 2) % s.m
		}
	case F:
		if IRA.A == IRB.A && IRA.B == IRB.B {
			nextPC = (PC + 2) % s.m
		}
	case X:
		if IRA.A == IRB.B && IRA.B == IRB.A {
			nextPC = (PC + 2) % s.m
		}
	case I:
		if IRA.Op == IRB.Op && IRA.OpMode == IRB.OpMode &&
			IRA.AMode == IRB.AMode && IRA.A == IRB.A &&
			IRA.BMode == IRB.BMode && IRA.B == IRB.B {
			nextPC = (PC + 2) % s.m
		}
	}
	s.Report(Report{Type: WarriorTaskPush, WarriorIndex: w.index, Address: nextPC})
	w.pq.Push(nextPC)
}

func (s *reportSim) sne(IR, IRA, IRB Instruction, PC Address, w *warrior) {

	nextPC := (PC + 1) % s.m
	switch IR.OpMode {
	case A:
		if IRA.A != IRB.A {
			nextPC = (PC + 2) % s.m
		}
	case B:
		if IRA.B != IRB.B {
			nextPC = (PC + 2) % s.m
		}
	case AB:
		if IRA.A != IRB.B {
			nextPC = (PC + 2) % s.m
		}
	case BA:
		if IRA.B != IRB.A {
			nextPC = (PC + 2) % s.m
		}
	case F:
		if IRA.A != IRB.A || IRA.B != IRB.B {
			nextPC = (PC + 2) % s.m
		}
	case X:
		if IRA.A != IRB.B || IRA.B != IRB.A {
			nextPC = (PC + 2) % s.m
		}
	case I:
		if IRA.Op != IRB.Op || IRA.OpMode != IRB.OpMode ||
			IRA.AMode != IRB.AMode || IRA.A != IRB.A ||
			IRA.BMode != IRB.BMode || IRA.B != IRB.B {
			nextPC = (PC + 2) % s.m
		}
	}
	s.Report(Report{Type: WarriorTaskPush, WarriorIndex: w.index, Address: nextPC})
	w.pq.Push(nextPC)
}

func (s *reportSim) slt(IR, IRA, IRB Instruction, PC Address, w *warrior) {
	nextPC := (PC + 1) % s.m
	switch IR.OpMode {
	case A:
		if IRA.A < IRB.A {
			nextPC = (PC + 2) % s.m
		}
	case B:
		if IRA.B < IRB.B {
			nextPC = (PC + 2) % s.m
		}
	case AB:
		if IRA.A < IRB.B {
			nextPC = (PC + 2) % s.m
		}
	case BA:
		if IRA.B < IRB.A {
			nextPC = (PC + 2) % s.m
		}
	case F:
		fallthrough
	case I:
		if IRA.A < IRB.A && IRA.B < IRB.B {
			nextPC = (PC + 2) % s.m
		}

	case X:
		if IRA.A < IRB.B && IRA.B < IRB.A {
			nextPC = (PC + 2) % s.m
		}
	}
	s.Report(Report{Type: WarriorTaskPush, WarriorIndex: w.index, Address: nextPC})
	w.pq.Push(nextPC)
}

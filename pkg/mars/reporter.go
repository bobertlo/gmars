package mars

import (
	"fmt"
)

type ReportType uint8

const (
	CycleStart ReportType = iota
	CycleEnd
	WarriorSpawn
	WarriorTaskPop
	WarriorTaskTerminate
	WarriorTerminate
	WarriorRead
	WarriorWrite
	WarriorDecrement
	WarriorIncrement
)

type Report struct {
	Type         ReportType
	Cycle        int
	WarriorIndex int
	Address      Address
}

type Reporter interface {
	Report(r Report)
}

type debugReporter struct {
	s *Simulator
}

func NewDebugReporter(s *Simulator) Reporter {
	return &debugReporter{s: s}
}

func (r *debugReporter) Report(report Report) {
	switch report.Type {
	case CycleStart:
		fmt.Printf("Cycle %d\n", r.s.cycleCount)
	case WarriorTaskPop:
		fmt.Printf("W%02d %04d: %s\n", report.WarriorIndex, report.Address, r.s.mem[report.Address].NormString(r.s.m))
	case WarriorTaskTerminate:
		fmt.Printf("W%02d %04d: Task Terminated\n", report.WarriorIndex, report.Address)
	case WarriorTerminate:
		fmt.Printf("W%02d %04d: Warrior Terminated\n", report.WarriorIndex, report.Address)
	case WarriorRead:
		fmt.Printf("W%02d %04d: Read\n", report.WarriorIndex, report.Address)
	case WarriorWrite:
		fmt.Printf("W%02d %04d: Write\n", report.WarriorIndex, report.Address)
	case WarriorIncrement:
		fmt.Printf("W%02d %04d: Increment\n", report.WarriorIndex, report.Address)
	case WarriorDecrement:
		fmt.Printf("W%02d %04d: Decrement\n", report.WarriorIndex, report.Address)
	}
}

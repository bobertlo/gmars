package gmars

import (
	"fmt"
)

type ReportType uint8

const (
	SimReset ReportType = iota
	CycleStart
	CycleEnd
	WarriorSpawn
	WarriorTaskPop
	WarriorTaskPush
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
	s Simulator
}

func NewDebugReporter(s Simulator) Reporter {
	return &debugReporter{s: s}
}

func (r *debugReporter) Report(report Report) {
	switch report.Type {
	case SimReset:
		fmt.Printf("Simulator reset\n")
	case CycleStart:
		fmt.Printf("%d\n", r.s.CycleCount())
	case WarriorSpawn:
		fmt.Printf("w%02d %04d: Warrior Spawn\n", report.WarriorIndex, report.Address)
	case WarriorTaskPop:
		fmt.Printf("W%02d %04d: Exec %s\n", report.WarriorIndex, report.Address, r.s.GetMem(report.Address).NormString(r.s.CoreSize()))
	case WarriorTaskPush:
		fmt.Printf("W%02d: Task Push %04d\n", report.WarriorIndex, report.Address)
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

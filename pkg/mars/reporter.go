package mars

import (
	"fmt"
	"os"
)

type Reporter interface {
	AddressRead(wi int, a Address)
	AddressWrite(wi int, a Address)
	AddressIncrement(wi int, a Address)
	AddressDecrement(wi int, a Address)
	TaskTerminate(wi int, a Address)
	TurnStart(n int)
	ResetMars()
	WarriorAdd(wi int, name, author string)
	WarriorSpawn(wi int, origin, entry Address)
	WarriorTaskPop(wi int, pc Address)
	WarriorTaskPush(wi int, pc Address)
	WarriorTerminate(wi int)
}

type DebugReporter struct {
	s    MARS
	turn int
}

func NewDebugReporter(s MARS) Reporter {
	return &DebugReporter{s: s}
}

func (r *DebugReporter) AddressRead(wi int, a Address) {
	fmt.Fprintf(os.Stderr, "w%02d: %05d read\n", wi, a)
}

func (r *DebugReporter) AddressWrite(wi int, a Address) {
	fmt.Fprintf(os.Stderr, "w%02d: %05d write\n", wi, a)
}

func (r *DebugReporter) AddressIncrement(wi int, a Address) {
	fmt.Fprintf(os.Stderr, "w%02d: %05d++\n", wi, a)
}

func (r *DebugReporter) AddressDecrement(wi int, a Address) {
	fmt.Fprintf(os.Stderr, "w%02d: %05d++\n", wi, a)
}

func (r *DebugReporter) TaskTerminate(wi int, a Address) {

}

func (r *DebugReporter) TurnStart(n int) {}

func (r *DebugReporter) ResetMars() {
	fmt.Fprintf(os.Stderr, "MARS reset")
}

func (r *DebugReporter) WarriorAdd(wi int, name, author string) {}

func (r *DebugReporter) WarriorSpawn(wi int, origin, entry Address) {}
func (r *DebugReporter) WarriorTaskPop(wi int, pc Address)          {}
func (r *DebugReporter) WarriorTaskPush(wi int, pc Address)         {}
func (r *DebugReporter) WarriorTerminate(wi int)                    {}

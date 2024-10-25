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
	InitMars()
	ResetMars()
	WarriorAdd(wi int, name, author string)
	WarriorSpawn(wi int, origin, entry Address)
	WarriorTaskPop(wi int, pc Address)
	WarriorTaskPush(wi int, pc Address)
	WarriorTerminate(wi int)
}

type debugReporter struct {
	s    *Simulator
	turn int
}

func NewDebugReporter(s *Simulator) Reporter {
	return &debugReporter{s: s}
}

func (r *debugReporter) AddressRead(wi int, a Address) {
	fmt.Fprintf(os.Stderr, "w%02d: %05d read\n", wi, a)
}

func (r *debugReporter) AddressWrite(wi int, a Address) {
	fmt.Fprintf(os.Stderr, "w%02d: %05d write\n", wi, a)
}

func (r *debugReporter) AddressIncrement(wi int, a Address) {
	fmt.Fprintf(os.Stderr, "w%02d: %05d++\n", wi, a)
}

func (r *debugReporter) AddressDecrement(wi int, a Address) {
	fmt.Fprintf(os.Stderr, "w%02d: %05d++\n", wi, a)
}

func (r *debugReporter) TaskTerminate(wi int, a Address) {
	fmt.Fprintf(os.Stderr, "w%02d: %05d TERMINATE\n", wi, a)
}

func (r *debugReporter) TurnStart(n int) {
	fmt.Fprintf(os.Stderr, "TURN %05d\n", n)
}

func (r *debugReporter) ResetMars() {
	fmt.Fprintf(os.Stderr, "MARS reset\n")
}

func (r *debugReporter) InitMars() {
	fmt.Fprintf(os.Stderr, "MARS reset\n")
}

func (r *debugReporter) WarriorAdd(wi int, name, author string) {
	fmt.Fprintf(os.Stderr, "w%02d: ADD '%s' by '%s'\n", wi, name, author)
}

func (r *debugReporter) WarriorSpawn(wi int, origin, entry Address) {
	fmt.Fprintf(os.Stderr, "w%02d: SPAWN %05d START %05d\n", wi, origin, entry)
}
func (r *debugReporter) WarriorTaskPop(wi int, pc Address) {
	fmt.Fprintf(os.Stderr, "w%02d: EXEC %05d\n", wi, pc)
}
func (r *debugReporter) WarriorTaskPush(wi int, pc Address) {
	fmt.Fprintf(os.Stderr, "w%02d: PUSH %05d\n", wi, pc)
}
func (r *debugReporter) WarriorTerminate(wi int) {
	fmt.Fprintf(os.Stderr, "w%02d: TERMINATE\n", wi)
}

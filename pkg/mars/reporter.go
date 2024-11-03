package mars

import (
	"fmt"
	"os"
)

type Reporter interface {
	CycleStart(n int)
	CycleEnd(n int)
	WarriorAdd(wi int, name, author string)
	WarriorSpawn(wi int, origin, entry Address)
	WarriorTaskPop(wi int, pc Address)
	WarriorTaskPush(wi int, pc Address)
	WarriorTaskTerminate(wi int, a Address)
	WarriorTerminate(wi int)
	WarriorRead(wi int, a Address)
	WarriorWrite(wi int, a Address)
	WarriorDecrement(wi int, a Address)
	WarriorIncrement(wi int, a Address)
}

type debugReporter struct {
	s *Simulator
}

func NewDebugReporter(s *Simulator) Reporter {
	return &debugReporter{s: s}
}

func (r *debugReporter) CycleStart(n int) {
	fmt.Fprintf(os.Stderr, "TURN %05d\n", n)
}

func (r *debugReporter) CycleEnd(n int) {
	fmt.Fprintf(os.Stderr, "TURN %05d\n", n)
}

func (r *debugReporter) WarriorAdd(wi int, name, author string) {
	fmt.Fprintf(os.Stderr, "w%02d: ADD '%s' by '%s'\n", wi, name, author)
}

func (r *debugReporter) WarriorSpawn(wi int, origin, entry Address) {
	fmt.Fprintf(os.Stderr, "w%02d: SPAWN %05d START %05d\n", wi, origin, entry)
}

func (r *debugReporter) WarriorTaskPop(wi int, pc Address) {
	fmt.Fprintf(os.Stderr, "w%02d: EXEC %05d %s\n", wi, pc, r.s.mem[pc].NormString(r.s.m))
}

func (r *debugReporter) WarriorTaskPush(wi int, pc Address) {
	fmt.Fprintf(os.Stderr, "w%02d: PUSH %05d\n", wi, pc)
}

func (r *debugReporter) WarriorTaskTerminate(wi int, a Address) {
	fmt.Fprintf(os.Stderr, "w%02d: %05d TERMINATE\n", wi, a)
}

func (r *debugReporter) WarriorTerminate(wi int) {
	fmt.Fprintf(os.Stderr, "w%02d: TERMINATE\n", wi)
}

func (r *debugReporter) WarriorRead(wi int, a Address) {
	fmt.Fprintf(os.Stderr, "w%02d: %05d read\n", wi, a)
}

func (r *debugReporter) WarriorWrite(wi int, a Address) {
	fmt.Fprintf(os.Stderr, "w%02d: %05d write\n", wi, a)
}

func (r *debugReporter) WarriorIncrement(wi int, a Address) {
	fmt.Fprintf(os.Stderr, "w%02d: %05d++\n", wi, a)
}

func (r *debugReporter) WarriorDecrement(wi int, a Address) {
	fmt.Fprintf(os.Stderr, "w%02d: %05d++\n", wi, a)
}

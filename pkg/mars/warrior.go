package mars

import "fmt"

type WarriorState uint8

const (
	RESET WarriorState = iota
	ALIVE
	DEAD
)

type WarriorData struct {
	Name     string        // Warrior Name
	Author   string        // Author Name
	Strategy string        // Strategy including multiple lines
	Code     []Instruction // Program Instructions
	Start    int           // Program Entry Point
}

// Copy creates a deep copy of a WarriorData object
func (w *WarriorData) Copy() *WarriorData {
	codeCopy := make([]Instruction, len(w.Code))
	copy(codeCopy, w.Code)
	return &WarriorData{
		Name:     w.Name,
		Author:   w.Author,
		Strategy: w.Strategy,
		Code:     codeCopy,
		Start:    w.Start,
	}
}

// Warrior is a manifestation WarriorData in a Simulator
type Warrior struct {
	data  *WarriorData
	sim   *Simulator
	index int
	pq    *processQueue
	// pspace []Instruction
	state WarriorState
}

// Name returns the Warrior's Name
func (w *Warrior) Name() string {
	return w.data.Name
}

// Author returns the Warrior's Author
func (w *Warrior) Author() string {
	return w.data.Author
}

// Length returns the Warrior's code length
func (w *Warrior) Length() int {
	return len(w.data.Code)
}

func (w *Warrior) State() WarriorState {
	return w.state
}

// Alive returns true if the warrior is alive
func (w *Warrior) Alive() bool {
	return w.state == ALIVE
}

func (w *Warrior) ThreadCount() Address {
	return w.pq.Len()
}

func (w *Warrior) LoadCode() string {
	out := ""

	if len(w.data.Code) == 0 {
		return ""
	}

	if w.sim == nil || (w.sim != nil && !w.sim.legacy) {
		out += "       ORG      START\n"
	}
	for i, inst := range w.data.Code {
		start := "     "

		if i == int(w.data.Start) {
			start = "START"
		}
		opmode := ""
		if w.sim == nil || !w.sim.legacy {
			opmode = "." + inst.OpMode.String()
		}

		line := fmt.Sprintf("%s  %3s%-3s %1s %5d, %1s %5d     \n",
			start,
			inst.Op,
			opmode,
			inst.AMode,
			w.sim.addressSigned(inst.A),
			inst.BMode, w.sim.addressSigned(inst.B))
		out = out + line
	}
	if w.sim != nil && w.sim.legacy {
		out += "       END      START\n"
	}
	return out
}

func (w *Warrior) LoadCodePMARS() string {
	header := fmt.Sprintf("Program \"%s\" (length %d) by \"%s\"\n\n", w.Name(), w.Length(), w.Author())

	if len(w.data.Code) > 0 {
		return header + w.LoadCode() + "\n"
	}
	return header
}

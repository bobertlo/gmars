package gmars

type CoreState uint8

const (
	CoreEmpty CoreState = iota
	CoreExecuted
	CoreWritten
	CoreIncremented
	CoreDecremented
	CoreRead
	CoreTerminated
)

// HistoryEntry represents a single historical event in the core
type HistoryEntry struct {
	State CoreState
	Color int
}

// StateRecorder implements a Reporter which records operations performed at each core address
// and maintains a configurable history of events
type StateRecorder struct {
	sim         ReportingSimulator
	coresize    Address
	recordReads bool

	historySize int              // Maximum number of historical events to store per address
	history     [][]HistoryEntry // Slice of circular buffers, one for each address
	historyPos  []int            // Current position in each circular buffer
}

func NewStateRecorder(sim ReportingSimulator, historySize int) *StateRecorder {
	coresize := sim.CoreSize()

	history := make([][]HistoryEntry, coresize)
	historyPos := make([]int, coresize)

	for i := Address(0); i < coresize; i++ {
		history[i] = make([]HistoryEntry, historySize)
		for j := 0; j < historySize; j++ {
			history[i][j] = HistoryEntry{
				State: CoreEmpty,
				Color: -1,
			}
		}
	}

	return &StateRecorder{
		sim:         sim,
		coresize:    coresize,
		history:     history,
		historySize: historySize,
		historyPos:  historyPos,
	}
}

// Get a historical state, up to r.historySize.
// Use positive numbers as n, e.g. n==0 is current state, n==1 is one state back.
func (r *StateRecorder) GetMemStateN(a Address, n int) (CoreState, int) {
	pos := (r.historyPos[a] - (n + 1) + r.historySize) % r.historySize
	entry := r.history[a][pos]
	return entry.State, entry.Color
}

// GetMemState returns the current state and color of a memory address
// by looking at the most recent history entry.
func (r *StateRecorder) GetMemState(a Address) (CoreState, int) {
	return r.GetMemStateN(a, 0)
}

func (r *StateRecorder) SetRecordRead(val bool) {
	r.recordReads = val
}

// recordHistory adds a new entry to the history for the given address
func (r *StateRecorder) recordHistory(addr Address, state CoreState, color int, report Report) {
	entry := HistoryEntry{
		State: state,
		Color: color,
	}

	r.history[addr][r.historyPos[addr]] = entry
	r.historyPos[addr] = (r.historyPos[addr] + 1) % r.historySize
}

func (r *StateRecorder) reset() {
	for i := Address(0); i < r.coresize; i++ {
		r.historyPos[i] = 0
		for j := 0; j < r.historySize; j++ {
			r.history[i][j] = HistoryEntry{
				State: CoreEmpty,
				Color: -1,
			}
		}
	}
}

func (r *StateRecorder) Report(report Report) {
	affected := report.Address % r.coresize

	switch report.Type {
	case SimReset:
		r.reset()
	case WarriorSpawn:
		w := r.sim.GetWarrior(report.WarriorIndex)
		for i := report.Address; i < report.Address+Address(w.Length()); i++ {
			r.recordHistory(i%r.coresize, CoreWritten, report.WarriorIndex, report)
		}
	case WarriorTaskTerminate:
		r.recordHistory(affected, CoreTerminated, report.WarriorIndex, report)
	case WarriorTaskPop:
		r.recordHistory(affected, CoreExecuted, report.WarriorIndex, report)
	case WarriorWrite:
		r.recordHistory(affected, CoreWritten, report.WarriorIndex, report)
	case WarriorRead:
		if !r.recordReads {
			return
		}
		r.recordHistory(affected, CoreRead, report.WarriorIndex, report)
	case WarriorIncrement:
		r.recordHistory(affected, CoreIncremented, report.WarriorIndex, report)
	case WarriorDecrement:
		r.recordHistory(affected, CoreDecremented, report.WarriorIndex, report)
	}
}

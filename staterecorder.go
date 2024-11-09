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

// StateRecorder implements a Reporter which records the most recent operation
// performed each core address and the warrior index associated. The initial
// state of each address is CoreEmpty with a warrior index of -1.
type StateRecorder struct {
	sim         ReportingSimulator
	coresize    Address
	color       []int
	state       []CoreState
	recordReads bool
}

func NewStateRecorder(sim ReportingSimulator) *StateRecorder {
	coresize := sim.CoreSize()

	color := make([]int, coresize)
	for i := Address(0); i < coresize; i++ {
		color[i] = -1
	}

	state := make([]CoreState, coresize)

	return &StateRecorder{
		sim:      sim,
		coresize: coresize,
		color:    color,
		state:    state,
	}
}

func (r *StateRecorder) GetMemState(a Address) (CoreState, int) {
	return r.state[a], r.color[a]
}

func (r *StateRecorder) SetRecordRead(val bool) {
	r.recordReads = val
}

func (r *StateRecorder) reset() {
	for i := Address(0); i < r.coresize; i++ {
		r.state[i] = CoreEmpty
		r.color[i] = -1
	}
}

func (r *StateRecorder) Report(report Report) {
	switch report.Type {
	case SimReset:
		r.reset()
	case WarriorSpawn:
		w := r.sim.GetWarrior(report.WarriorIndex)
		for i := report.Address; i < report.Address+Address(w.Length()); i++ {
			r.color[i%r.coresize] = report.WarriorIndex
			r.state[i%r.coresize] = CoreWritten
		}
	case WarriorTaskTerminate:
		r.color[report.Address] = report.WarriorIndex
		r.state[report.Address] = CoreTerminated
	case WarriorTaskPop:
		r.color[report.Address] = report.WarriorIndex
		r.state[report.Address] = CoreExecuted
	case WarriorWrite:
		r.color[report.Address] = report.WarriorIndex
		r.state[report.Address] = CoreWritten
	case WarriorRead:
		if r.recordReads {
			r.color[report.Address] = report.WarriorIndex
			r.state[report.Address] = CoreRead
		}
	case WarriorIncrement:
		r.color[report.Address] = report.WarriorIndex
		r.state[report.Address] = CoreIncremented
	case WarriorDecrement:
		r.color[report.Address] = report.WarriorIndex
		r.state[report.Address] = CoreDecremented
	}
}

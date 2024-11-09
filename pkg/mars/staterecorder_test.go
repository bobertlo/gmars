package mars

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recorderTest struct {
	input     []string
	output    []string
	color     []int
	state     []CoreState
	coresize  Address
	processes Address
	start     Address
	turns     int
	offset    Address
	pq        []Address
}

func TestStateRecorderDwarf(t *testing.T) {
	tests := []recorderTest{
		{
			input: []string{
				"add.ab #4, $3", "mov.i  $2, @2", "jmp.b $-2, $0", "dat.f  #0, #0",
			},
			output: []string{
				"add.ab #4, $3", "mov.i  $2, @2", "jmp.b $-2, $0", "dat.f  #0, #12",
				"dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f #0, #4",
				"dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f #0, #8",
				"dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f #0, #12",
			},
			color: []int{
				0, 0, 0, 0,
				-1, -1, -1, 0,
				-1, -1, -1, 0,
				-1, -1, -1, 0,
			},
			state: []CoreState{
				CoreExecuted, CoreExecuted, CoreExecuted, CoreWritten,
				CoreEmpty, CoreEmpty, CoreEmpty, CoreWritten,
				CoreEmpty, CoreEmpty, CoreEmpty, CoreWritten,
				CoreEmpty, CoreEmpty, CoreEmpty, CoreWritten,
			},
			turns: 9,
			pq:    []Address{0},
		},
	}
	runStateRecorderTests(t, "dwarf recorder", tests)
}

func runStateRecorderTests(t *testing.T, set_name string, tests []recorderTest) {
	for i, test := range tests {
		coresize := test.coresize
		if coresize == 0 {
			coresize = Address(len(test.output))
		}

		processes := test.processes
		if processes == 0 {
			processes = coresize
		}

		turns := test.turns
		if turns == 0 {
			turns = 1
		}

		if len(test.input) > int(coresize) || len(test.output) > int(coresize) {
			t.Fatalf("%s test %d: invalid coresize", set_name, i)
		}

		code := make([]Instruction, len(test.input))
		for i, instring := range test.input {
			instruction := parseTestInstruction(t, instring, int(coresize))
			code[i] = instruction
		}

		expectedOutput := make([]Instruction, len(test.output))
		for i, instring := range test.output {
			instruction := parseTestInstruction(t, instring, int(coresize))
			expectedOutput[i] = instruction
		}

		config := NewQuickConfig(ICWS88, coresize, processes, Address(turns), coresize)
		config.Distance = 0

		sim, err := newReportSim(config)
		require.NoError(t, err)
		rec := NewStateRecorder(sim)
		sim.AddReporter(rec)
		w, err := sim.addWarrior(&WarriorData{Code: code, Start: int(test.offset)})
		require.NoError(t, err)
		err = sim.spawnWarrior(0, test.start)
		require.NoError(t, err)

		for i := 0; i < turns; i++ {
			sim.RunCycle()
		}

		for j, expected := range expectedOutput {
			state, color := rec.GetMemState(Address(j))
			assert.Equal(t, expected, sim.GetMem(Address(j)), fmt.Sprintf("%s test %d address %d value", set_name, i, j))
			assert.Equal(t, test.color[j], color, fmt.Sprintf("%s test %d address %d color", set_name, i, j))
			assert.Equal(t, test.state[j], state, fmt.Sprintf("%s test %d address %d state", set_name, i, j))

		}
		assert.Equal(t, test.pq, w.pq.Values(), fmt.Sprintf("%s test %d", set_name, i))
	}

}

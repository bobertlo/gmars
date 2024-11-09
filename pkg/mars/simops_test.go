package mars

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type redcodeTest struct {
	input     []string
	output    []string
	coresize  Address
	processes Address
	start     Address
	turns     int
	offset    Address
	pq        []Address
}

func TestFold(t *testing.T) {
	sim, err := newReportSim(SimulatorConfig{
		Mode:       NOP94,
		CoreSize:   8000,
		Processes:  8000,
		Cycles:     80000,
		ReadLimit:  8000,
		WriteLimit: 8000,
		Length:     100,
		Distance:   100,
	})
	require.NoError(t, err)
	for i := Address(0); i < 8000; i += 7 {
		assert.Equal(t, i, sim.readFold(i))
		assert.Equal(t, i, sim.writeFold(i))
	}
}

func TestFoldLimit(t *testing.T) {
	sim, err := newReportSim(SimulatorConfig{
		Mode:       NOP94,
		CoreSize:   8000,
		Cycles:     80000,
		Processes:  8000,
		ReadLimit:  1000,
		WriteLimit: 1000,
		Length:     100,
		Distance:   100,
	})
	require.NoError(t, err)
	assert.Equal(t, Address(400), sim.readFold(1400))
	assert.Equal(t, Address(400), sim.writeFold(1400))
	assert.Equal(t, Address(8000-400), sim.readFold(8000-1400))
	assert.Equal(t, Address(8000-400), sim.writeFold(8000-1400))
}

func parseTestAddres(t *testing.T, input string, M int) (AddressMode, Address) {
	var mode AddressMode
	if len(input) == 0 {
		t.Fatalf("empty address")
	}

	switch input[0] {
	case '#':
		mode = IMMEDIATE
	case '$':
		mode = DIRECT
	case '*':
		mode = A_INDIRECT
	case '@':
		mode = B_INDIRECT
	case '{':
		mode = A_DECREMENT
	case '}':
		mode = A_INCREMENT
	case '<':
		mode = B_DECREMENT
	case '>':
		mode = B_INCREMENT
	default:
		t.Fatalf("invalid address mode: '%s'", input)
	}

	input = input[1:]
	if len(input) == 0 {
		t.Fatalf("missing address")
	}

	val, err := strconv.ParseInt(input, 10, 32)
	if err != nil {
		t.Fatalf("error parsing addres: %s", err)
	}

	mval := int(val) % M
	if mval < 0 {
		mval = M + mval
	}

	return mode, Address(mval)
}

func parseTestInstruction(t *testing.T, input string, M int) Instruction {
	lower := strings.ToLower(input)
	nocomma := strings.ReplaceAll(lower, ",", " ")
	fields := strings.Fields(nocomma)

	if len(fields) != 3 {
		t.Fatalf("len(fields) != 3: '%s'", input)
	}

	opTokens := strings.Split(fields[0], ".")
	if len(opTokens) != 2 {
		t.Fatalf("invalid op: '%s", fields[0])
	}

	op, err := getOpCode(opTokens[0])
	if err != nil {
		t.Fatalf("error parsing '%s': %s", input, err)
	}
	opMode, err := getOpMode(opTokens[1])
	if err != nil {
		t.Fatalf("error parsing '%s': %s", input, err)
	}

	amode, a := parseTestAddres(t, fields[1], M)
	bmode, b := parseTestAddres(t, fields[2], M)

	return Instruction{Op: op, OpMode: opMode, AMode: amode, A: a, BMode: bmode, B: b}
}

func runTests(t *testing.T, set_name string, tests []redcodeTest) {
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
		w, err := sim.addWarrior(&WarriorData{Code: code, Start: int(test.offset)})
		require.NoError(t, err)
		err = sim.spawnWarrior(0, test.start)
		require.NoError(t, err)

		for i := 0; i < turns; i++ {
			sim.RunCycle()
		}

		for j, expected := range expectedOutput {
			assert.Equal(t, expected, sim.GetMem(Address(j)), fmt.Sprintf("%s test %d address %d", set_name, i, j))
		}
		assert.Equal(t, test.pq, w.pq.Values(), fmt.Sprintf("%s test %d", set_name, i))
	}

}

func TestDat(t *testing.T) {
	tests := []redcodeTest{
		{
			input:  []string{"dat.f $0, $0"},
			output: []string{"dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{},
		},
		{
			input:  []string{"dat.f <0, $0"},
			output: []string{"dat.f <0, $-1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{},
		},
		{
			input:  []string{"dat.f >0, $0"},
			output: []string{"dat.f >0, $1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{},
		},
		{
			input:  []string{"dat.f {0, $0"},
			output: []string{"dat.f {-1, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{},
		},
		{
			input:  []string{"dat.f }0, $0"},
			output: []string{"dat.f }1, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{},
		},
		{
			input:  []string{"dat.f $0, <0"},
			output: []string{"dat.f $0, <-1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{},
		},
		{
			input:  []string{"dat.f $0, >0"},
			output: []string{"dat.f $0, >1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{},
		},
		{
			input:  []string{"dat.f $0, {0"},
			output: []string{"dat.f $-1, {0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{},
		},
		{
			input:  []string{"dat.f $0, }0"},
			output: []string{"dat.f $1, }0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{},
		},
		{
			input:  []string{"dat.f $0, <-1"},
			output: []string{"dat.f $0, <-1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $-1"},
			pq:     []Address{},
		},
		{
			input:  []string{"dat.f $0, >-1"},
			output: []string{"dat.f $0, >-1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $1"},
			pq:     []Address{},
		},
	}
	runTests(t, "dat", tests)
}

func TestMov(t *testing.T) {
	tests := []redcodeTest{
		// immediate a
		{
			input:  []string{"mov.i #0, $1"},
			output: []string{"mov.i #0, $1", "mov.i #0, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"mov.a $1, $2", "dat.f #1, #2"},
			output: []string{"mov.a $1, $2", "dat.f #1, #2", "dat.f $1, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"mov.b $1, $2", "dat.f #1, #2"},
			output: []string{"mov.b $1, $2", "dat.f #1, #2", "dat.f $0, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"mov.ab $1, $2", "dat.f #1, #2"},
			output: []string{"mov.ab $1, $2", "dat.f #1, #2", "dat.f $0, $1", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"mov.ba $1, $2", "dat.f #1, #2"},
			output: []string{"mov.ba $1, $2", "dat.f #1, #2", "dat.f $2, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"mov.f $1, $2", "dat.f #1, #2"},
			output: []string{"mov.f $1, $2", "dat.f #1, #2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"mov.i $1, $2", "add.ab #1, #2"},
			output: []string{"mov.i $1, $2", "add.ab #1, #2", "add.ab #1, #2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"mov.x $1, $2", "dat.f #1, #2"},
			output: []string{"mov.x $1, $2", "dat.f #1, #2", "dat.f $2, $1", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"mov.f $1, $-1", "dat.f #1, #2"},
			output: []string{"mov.f $1, $-1", "dat.f #1, #2", "dat.f $0, $0", "dat.f $1, $2"},
			pq:     []Address{1},
		},

		// indirect modifiers
		{
			input:  []string{"mov.i <1, $3"},
			output: []string{"mov.i <1, $3", "dat.f $0, $-1", "dat.f $0, $0", "mov.i <1, $3"},
			pq:     []Address{1},
		},
		{
			input:  []string{"mov.i >1, $3", "dat.f $1, $0"},
			output: []string{"mov.i >1, $3", "dat.f $1, $1", "dat.f $0, $0", "dat.f $1, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"mov.i {1, $3", "dat.f $1, $1"},
			output: []string{"mov.i {1, $3", "dat.f $0, $1", "dat.f $0, $0", "dat.f $0, $1"},
			pq:     []Address{1},
		},
		{
			input:  []string{"mov.i }1, $3", "dat.f $1, $1", "dat.f $2, $0"},
			output: []string{"mov.i }1, $3", "dat.f $2, $1", "dat.f $2, $0", "dat.f $2, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"mov.i $0, <1", "dat.f $0, $0"},
			output: []string{"mov.i $0, <1", "dat.f $0, $-1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"mov.i $0, >1", "dat.f $0, $0"},
			output: []string{"mov.i $0, >1", "mov.i $0, >1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"mov.i $0, {1", "dat.f $0, $0"},
			output: []string{"mov.i $0, {1", "dat.f $-1, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"mov.i $0, }1", "dat.f $0, $0"},
			output: []string{"mov.i $0, }1", "mov.i $0, }1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		//random tests
		{
			input:    []string{"dat.f $0, $3", "mov.i $1, >-1", "spl.b $0, $0"},
			output:   []string{"dat.f $0, $4", "mov.i $1, >-1", "spl.b $0, $0", "spl.b $0, $0"},
			pq:       []Address{2},
			coresize: 20,
			offset:   1,
		},
	}
	runTests(t, "mov", tests)
}

func TestAdd(t *testing.T) {
	tests := []redcodeTest{
		// immidiate a
		{
			input:  []string{"add.a #2, $1"},
			output: []string{"add.a #2, $1", "dat.f $2, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"add.b #3, $1"},
			output: []string{"add.b #3, $1", "dat.f $0, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"add.ab #3, $1"},
			output: []string{"add.ab #3, $1", "dat.f $0, $3", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"add.ba #3, $1"},
			output: []string{"add.ba #3, $1", "dat.f $1, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"add.f #3, $1"},
			output: []string{"add.f #3, $1", "dat.f $3, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"add.i #3, $1"},
			output: []string{"add.i #3, $1", "dat.f $3, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"add.x #3, $1"},
			output: []string{"add.x #3, $1", "dat.f $1, $3", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},

		// immediate b
		{
			input:  []string{"add.a #2, #1"},
			output: []string{"add.a #4, #1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"add.b #3, #1"},
			output: []string{"add.b #3, #2", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"add.ab #3, #1"},
			output: []string{"add.ab #3, #4", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"add.ba #3, #1"},
			output: []string{"add.ba #4, #1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"add.f #2, #1"},
			output: []string{"add.f #0, #2", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"add.i #2, #1"},
			output: []string{"add.i #0, #2", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"add.x #2, #1"},
			output: []string{"add.x #3, #3", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},

		// direct
		{
			input:  []string{"add.a $1, $2", "dat.f #1, #2", "dat.f $0, $0"},
			output: []string{"add.a $1, $2", "dat.f #1, #2", "dat.f $1, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"add.b $1, $2", "dat.f #1, #2", "dat.f $0, $0"},
			output: []string{"add.b $1, $2", "dat.f #1, #2", "dat.f $0, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"add.ab $1, $2", "dat.f #1, #2", "dat.f $0, $0"},
			output: []string{"add.ab $1, $2", "dat.f #1, #2", "dat.f $0, $1", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"add.ba $1, $2", "dat.f #1, #2", "dat.f $0, $0"},
			output: []string{"add.ba $1, $2", "dat.f #1, #2", "dat.f $2, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"add.f $1, $2", "dat.f #1, #2", "dat.f $0, $0"},
			output: []string{"add.f $1, $2", "dat.f #1, #2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"add.i $1, $2", "dat.f #1, #2", "dat.f $0, $0"},
			output: []string{"add.i $1, $2", "dat.f #1, #2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"add.x $1, $2", "dat.f #1, #2", "dat.f $0, $0"},
			output: []string{"add.x $1, $2", "dat.f #1, #2", "dat.f $2, $1", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		// random tests
		{
			input:    []string{"spl.b #-3044, <3044", "mov.i >-3044, $3045", "add.f $-2, $-1"},
			output:   []string{"spl.b #-3044, <3044", "mov.i >1912, $6089", "add.f $-2, $-1"},
			pq:       []Address{3},
			offset:   2,
			coresize: 8000,
		},
	}
	runTests(t, "add", tests)
}

func TestSub(t *testing.T) {
	tests := []redcodeTest{
		// direct
		{
			input:  []string{"sub.a $1, $2", "dat.f #1, #2", "dat.f $3, $3"},
			output: []string{"sub.a $1, $2", "dat.f #1, #2", "dat.f $2, $3", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"sub.b $1, $2", "dat.f #1, #2", "dat.f $3, $3"},
			output: []string{"sub.b $1, $2", "dat.f #1, #2", "dat.f $3, $1", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"sub.ab $1, $2", "dat.f #1, #2", "dat.f $3, $3"},
			output: []string{"sub.ab $1, $2", "dat.f #1, #2", "dat.f $3, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"sub.ba $1, $2", "dat.f #1, #2", "dat.f $3, $3"},
			output: []string{"sub.ba $1, $2", "dat.f #1, #2", "dat.f $1, $3", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"sub.f $1, $2", "dat.f #1, #2", "dat.f $3, $3"},
			output: []string{"sub.f $1, $2", "dat.f #1, #2", "dat.f $2, $1", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"sub.i $1, $2", "dat.f #1, #2", "dat.f $3, $3"},
			output: []string{"sub.i $1, $2", "dat.f #1, #2", "dat.f $2, $1", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"sub.x $1, $2", "dat.f #1, #2", "dat.f $3, $3"},
			output: []string{"sub.x $1, $2", "dat.f #1, #2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},

		// negative result
		{
			input:  []string{"sub.a $1, $2", "dat.f #2, #2", "dat.f $1, $1"},
			output: []string{"sub.a $1, $2", "dat.f #2, #2", "dat.f $-1, $1", "dat.f $0, $0"},
			pq:     []Address{1},
		},

		// negative input
		{
			input:  []string{"sub.a $1, $2", "dat.f #-1, #-1", "dat.f $1, $1"},
			output: []string{"sub.a $1, $2", "dat.f #-1, #-1", "dat.f $2, $1", "dat.f $0, $0"},
			pq:     []Address{1},
		},
	}
	runTests(t, "sub", tests)
}

func TestMUL(t *testing.T) {
	tests := []redcodeTest{
		{
			input:    []string{"mul.a $1, $2", "dat.f #3, #4", "dat.f #5, #6"},
			output:   []string{"mul.a $1, $2", "dat.f #3, #4", "dat.f #15, #6"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"mul.b $1, $2", "dat.f #3, #4", "dat.f #5, #6"},
			output:   []string{"mul.b $1, $2", "dat.f #3, #4", "dat.f #5, #24"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"mul.ab $1, $2", "dat.f #3, #4", "dat.f #5, #6"},
			output:   []string{"mul.ab $1, $2", "dat.f #3, #4", "dat.f #5, #18"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"mul.ba $1, $2", "dat.f #3, #4", "dat.f #5, #6"},
			output:   []string{"mul.ba $1, $2", "dat.f #3, #4", "dat.f #20, #6"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"mul.f $1, $2", "dat.f #3, #4", "dat.f #5, #6"},
			output:   []string{"mul.f $1, $2", "dat.f #3, #4", "dat.f #15, #24"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"mul.x $1, $2", "dat.f #3, #4", "dat.f #5, #6"},
			output:   []string{"mul.x $1, $2", "dat.f #3, #4", "dat.f #20, #18"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"mul.i $1, $2", "dat.f #3, #4", "dat.f #5, #6"},
			output:   []string{"mul.i $1, $2", "dat.f #3, #4", "dat.f #15, #24"},
			pq:       []Address{1},
			coresize: 256,
		},
	}
	runTests(t, "mul", tests)
}

func TestDIV(t *testing.T) {
	tests := []redcodeTest{
		{
			input:    []string{"div.a $1, $2", "dat.f #2, #3", "dat.f #6, #12"},
			output:   []string{"div.a $1, $2", "dat.f #2, #3", "dat.f #3, #12"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"div.a $1, $2", "dat.f #0, #3", "dat.f #6, #12"},
			output:   []string{"div.a $1, $2", "dat.f #0, #3", "dat.f #6, #12"},
			pq:       []Address{},
			coresize: 256,
		},

		{
			input:    []string{"div.b $1, $2", "dat.f #2, #3", "dat.f #6, #12"},
			output:   []string{"div.b $1, $2", "dat.f #2, #3", "dat.f #6, #4"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"div.b $1, $2", "dat.f #2, #0", "dat.f #6, #12"},
			output:   []string{"div.b $1, $2", "dat.f #2, #0", "dat.f #6, #12"},
			pq:       []Address{},
			coresize: 256,
		},

		{
			input:    []string{"div.ab $1, $2", "dat.f #2, #3", "dat.f #6, #12"},
			output:   []string{"div.ab $1, $2", "dat.f #2, #3", "dat.f #6, #6"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"div.ab $1, $2", "dat.f #0, #3", "dat.f #6, #12"},
			output:   []string{"div.ab $1, $2", "dat.f #0, #3", "dat.f #6, #12"},
			pq:       []Address{},
			coresize: 256,
		},

		{
			input:    []string{"div.ba $1, $2", "dat.f #2, #3", "dat.f #6, #12"},
			output:   []string{"div.ba $1, $2", "dat.f #2, #3", "dat.f #2, #12"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"div.ba $1, $2", "dat.f #2, #0", "dat.f #6, #12"},
			output:   []string{"div.ba $1, $2", "dat.f #2, #0", "dat.f #6, #12"},
			pq:       []Address{},
			coresize: 256,
		},

		{
			input:    []string{"div.f $1, $2", "dat.f #2, #3", "dat.f #6, #12"},
			output:   []string{"div.f $1, $2", "dat.f #2, #3", "dat.f #3, #4"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"div.f $1, $2", "dat.f #0, #3", "dat.f #6, #12"},
			output:   []string{"div.f $1, $2", "dat.f #0, #3", "dat.f #6, #4"},
			pq:       []Address{},
			coresize: 256,
		},
		{
			input:    []string{"div.f $1, $2", "dat.f #2, #0", "dat.f #6, #12"},
			output:   []string{"div.f $1, $2", "dat.f #2, #0", "dat.f #3, #12"},
			pq:       []Address{},
			coresize: 256,
		},

		{
			input:    []string{"div.i $1, $2", "dat.f #2, #3", "dat.f #6, #12"},
			output:   []string{"div.i $1, $2", "dat.f #2, #3", "dat.f #3, #4"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"div.i $1, $2", "dat.f #0, #3", "dat.f #6, #12"},
			output:   []string{"div.i $1, $2", "dat.f #0, #3", "dat.f #6, #4"},
			pq:       []Address{},
			coresize: 256,
		},
		{
			input:    []string{"div.i $1, $2", "dat.f #2, #0", "dat.f #6, #12"},
			output:   []string{"div.i $1, $2", "dat.f #2, #0", "dat.f #3, #12"},
			pq:       []Address{},
			coresize: 256,
		},

		{
			input:    []string{"div.x $1, $2", "dat.f #2, #3", "dat.f #6, #12"},
			output:   []string{"div.x $1, $2", "dat.f #2, #3", "dat.f #2, #6"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"div.x $1, $2", "dat.f #0, #3", "dat.f #6, #12"},
			output:   []string{"div.x $1, $2", "dat.f #0, #3", "dat.f #2, #12"},
			pq:       []Address{},
			coresize: 256,
		},
		{
			input:    []string{"div.x $1, $2", "dat.f #2, #0", "dat.f #6, #12"},
			output:   []string{"div.x $1, $2", "dat.f #2, #0", "dat.f #6, #6"},
			pq:       []Address{},
			coresize: 256,
		},
	}
	runTests(t, "div", tests)
}

func TestMOD(t *testing.T) {
	tests := []redcodeTest{
		{
			input:    []string{"mod.a $1, $2", "dat.f #3, #4", "dat.f #10, #15"},
			output:   []string{"mod.a $1, $2", "dat.f #3, #4", "dat.f #1, #15"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"mod.a $1, $2", "dat.f #0, #4", "dat.f #10, #15"},
			output:   []string{"mod.a $1, $2", "dat.f #0, #4", "dat.f #10, #15"},
			pq:       []Address{},
			coresize: 256,
		},

		{
			input:    []string{"mod.b $1, $2", "dat.f #3, #4", "dat.f #10, #15"},
			output:   []string{"mod.b $1, $2", "dat.f #3, #4", "dat.f #10, #3"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"mod.b $1, $2", "dat.f #3, #0", "dat.f #10, #15"},
			output:   []string{"mod.b $1, $2", "dat.f #3, #0", "dat.f #10, #15"},
			pq:       []Address{},
			coresize: 256,
		},

		{
			input:    []string{"mod.ab $1, $2", "dat.f #3, #4", "dat.f #10, #15"},
			output:   []string{"mod.ab $1, $2", "dat.f #3, #4", "dat.f #10, #0"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"mod.ab $1, $2", "dat.f #0, #4", "dat.f #10, #15"},
			output:   []string{"mod.ab $1, $2", "dat.f #0, #4", "dat.f #10, #15"},
			pq:       []Address{},
			coresize: 256,
		},

		{
			input:    []string{"mod.ba $1, $2", "dat.f #3, #4", "dat.f #10, #15"},
			output:   []string{"mod.ba $1, $2", "dat.f #3, #4", "dat.f #2, #15"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"mod.ba $1, $2", "dat.f #3, #0", "dat.f #10, #15"},
			output:   []string{"mod.ba $1, $2", "dat.f #3, #0", "dat.f #10, #15"},
			pq:       []Address{},
			coresize: 256,
		},

		{
			input:    []string{"mod.f $1, $2", "dat.f #3, #4", "dat.f #10, #15"},
			output:   []string{"mod.f $1, $2", "dat.f #3, #4", "dat.f #1, #3"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"mod.f $1, $2", "dat.f #3, #0", "dat.f #10, #15"},
			output:   []string{"mod.f $1, $2", "dat.f #3, #0", "dat.f #1, #15"},
			pq:       []Address{},
			coresize: 256,
		},
		{
			input:    []string{"mod.f $1, $2", "dat.f #0, #4", "dat.f #10, #15"},
			output:   []string{"mod.f $1, $2", "dat.f #0, #4", "dat.f #10, #3"},
			pq:       []Address{},
			coresize: 256,
		},

		{
			input:    []string{"mod.i $1, $2", "dat.f #3, #4", "dat.f #10, #15"},
			output:   []string{"mod.i $1, $2", "dat.f #3, #4", "dat.f #1, #3"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"mod.i $1, $2", "dat.f #3, #0", "dat.f #10, #15"},
			output:   []string{"mod.i $1, $2", "dat.f #3, #0", "dat.f #1, #15"},
			pq:       []Address{},
			coresize: 256,
		},
		{
			input:    []string{"mod.i $1, $2", "dat.f #0, #4", "dat.f #10, #15"},
			output:   []string{"mod.i $1, $2", "dat.f #0, #4", "dat.f #10, #3"},
			pq:       []Address{},
			coresize: 256,
		},

		{
			input:    []string{"mod.x $1, $2", "dat.f #3, #4", "dat.f #10, #15"},
			output:   []string{"mod.x $1, $2", "dat.f #3, #4", "dat.f #2, #0"},
			pq:       []Address{1},
			coresize: 256,
		},
		{
			input:    []string{"mod.x $1, $2", "dat.f #3, #0", "dat.f #10, #15"},
			output:   []string{"mod.x $1, $2", "dat.f #3, #0", "dat.f #10, #0"},
			pq:       []Address{},
			coresize: 256,
		},
		{
			input:    []string{"mod.x $1, $2", "dat.f #0, #4", "dat.f #10, #15"},
			output:   []string{"mod.x $1, $2", "dat.f #0, #4", "dat.f #2, #15"},
			pq:       []Address{},
			coresize: 256,
		},
	}
	runTests(t, "mod", tests)
}

func TestJMP(t *testing.T) {
	tests := []redcodeTest{
		{
			input:  []string{"jmp.a $2, $0"},
			output: []string{"jmp.a $2, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmp.b $2, $0"},
			output: []string{"jmp.b $2, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmp.ab $2, $0"},
			output: []string{"jmp.ab $2, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmp.ba $2, $0"},
			output: []string{"jmp.ba $2, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmp.f $2, $0"},
			output: []string{"jmp.f $2, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmp.x $2, $0"},
			output: []string{"jmp.x $2, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmp.i $2, $0"},
			output: []string{"jmp.i $2, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		// non-zero PC
		{
			input:  []string{"dat.f $0, $0", "jmp.b $2, $0"},
			output: []string{"dat.f $0, $0", "jmp.b $2, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{3},
			offset: 1,
		},
		{
			input:  []string{"dat.f $0, $0", "jmp.b <1, $0"},
			output: []string{"dat.f $0, $0", "jmp.b <1, $0", "dat.f $0, $-1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
			offset: 1,
		},
		{
			input:  []string{"dat.f $0, $0", "jmp.b >1, $0"},
			output: []string{"dat.f $0, $0", "jmp.b >1, $0", "dat.f $0, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
			offset: 1,
		},
		{
			input:  []string{"dat.f $0, $0", "jmp.b >0, $0"},
			output: []string{"dat.f $0, $0", "jmp.b >0, $1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
			offset: 1,
		},
		{
			input:    []string{"dat.f $10, $0", "dat.f $0, $0", "jmp.b *-2, $0"},
			output:   []string{"dat.f $10, $0", "dat.f $0, $0", "jmp.b *-2, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:       []Address{10},
			offset:   2,
			coresize: 20,
		},
	}
	runTests(t, "jmp", tests)
}

func TestJMZ(t *testing.T) {
	tests := []redcodeTest{
		// postive cases all modes
		{
			input:  []string{"jmz.a $2, $1", "dat.f $0, $1"},
			output: []string{"jmz.a $2, $1", "dat.f $0, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmz.ba $2, $1", "dat.f $0, $1"},
			output: []string{"jmz.ba $2, $1", "dat.f $0, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmz.b $2, $1", "dat.f $1, $0"},
			output: []string{"jmz.b $2, $1", "dat.f $1, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmz.ab $2, $1", "dat.f $1, $0"},
			output: []string{"jmz.ab $2, $1", "dat.f $1, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmz.f $2, $1", "dat.f $0, $0"},
			output: []string{"jmz.f $2, $1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmz.x $2, $1", "dat.f $0, $0"},
			output: []string{"jmz.x $2, $1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmz.i $2, $1", "dat.f $0, $0"},
			output: []string{"jmz.i $2, $1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		// negative cases all modes
		{
			input:  []string{"jmz.a $2, $1", "dat.f $1, $0"},
			output: []string{"jmz.a $2, $1", "dat.f $1, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmz.ba $2, $1", "dat.f $1, $0"},
			output: []string{"jmz.ba $2, $1", "dat.f $1, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmz.a $2, $1", "dat.f $1, $0"},
			output: []string{"jmz.a $2, $1", "dat.f $1, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmz.b $2, $1", "dat.f $0, $1"},
			output: []string{"jmz.b $2, $1", "dat.f $0, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmz.ab $2, $1", "dat.f $0, $1"},
			output: []string{"jmz.ab $2, $1", "dat.f $0, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmz.f $2, $1", "dat.f $0, $1"},
			output: []string{"jmz.f $2, $1", "dat.f $0, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmz.f $2, $1", "dat.f $1, $0"},
			output: []string{"jmz.f $2, $1", "dat.f $1, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmz.f $2, $1", "dat.f $1, $1"},
			output: []string{"jmz.f $2, $1", "dat.f $1, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmz.x $2, $1", "dat.f $0, $1"},
			output: []string{"jmz.x $2, $1", "dat.f $0, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmz.x $2, $1", "dat.f $1, $0"},
			output: []string{"jmz.x $2, $1", "dat.f $1, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmz.x $2, $1", "dat.f $1, $1"},
			output: []string{"jmz.x $2, $1", "dat.f $1, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmz.i $2, $1", "dat.f $0, $1"},
			output: []string{"jmz.i $2, $1", "dat.f $0, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmz.i $2, $1", "dat.f $1, $0"},
			output: []string{"jmz.i $2, $1", "dat.f $1, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmz.i $2, $1", "dat.f $1, $1"},
			output: []string{"jmz.i $2, $1", "dat.f $1, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		// non zero PC
		{
			input:  []string{"dat.f $0, $0", "jmz.i $2, $1"},
			output: []string{"dat.f $0, $0", "jmz.i $2, $1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{3},
			offset: 1,
		},
	}
	runTests(t, "jmz", tests)
}

func TestJMN(t *testing.T) {
	tests := []redcodeTest{
		// positive cases all modes
		{
			input:  []string{"jmn.a $2, $1", "dat.f $1, $0"},
			output: []string{"jmn.a $2, $1", "dat.f $1, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmn.ba $2, $1", "dat.f $1, $0"},
			output: []string{"jmn.ba $2, $1", "dat.f $1, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmn.b $2, $1", "dat.f $0, $1"},
			output: []string{"jmn.b $2, $1", "dat.f $0, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmn.ab $2, $1", "dat.f $0, $1"},
			output: []string{"jmn.ab $2, $1", "dat.f $0, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmn.f $2, $1", "dat.f $1, $0"},
			output: []string{"jmn.f $2, $1", "dat.f $1, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmn.f $2, $1", "dat.f $0, $1"},
			output: []string{"jmn.f $2, $1", "dat.f $0, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmn.f $2, $1", "dat.f $1, $1"},
			output: []string{"jmn.f $2, $1", "dat.f $1, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmn.x $2, $1", "dat.f $1, $0"},
			output: []string{"jmn.x $2, $1", "dat.f $1, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmn.x $2, $1", "dat.f $0, $1"},
			output: []string{"jmn.x $2, $1", "dat.f $0, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmn.x $2, $1", "dat.f $1, $1"},
			output: []string{"jmn.x $2, $1", "dat.f $1, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmn.i $2, $1", "dat.f $1, $0"},
			output: []string{"jmn.i $2, $1", "dat.f $1, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmn.i $2, $1", "dat.f $0, $1"},
			output: []string{"jmn.i $2, $1", "dat.f $0, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"jmn.i $2, $1", "dat.f $1, $1"},
			output: []string{"jmn.i $2, $1", "dat.f $1, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		// negative cases all modes
		{
			input:  []string{"jmn.a $2, $1", "dat.f $0, $1"},
			output: []string{"jmn.a $2, $1", "dat.f $0, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmn.ba $2, $1", "dat.f $0, $1"},
			output: []string{"jmn.ba $2, $1", "dat.f $0, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmn.b $2, $1", "dat.f $1, $0"},
			output: []string{"jmn.b $2, $1", "dat.f $1, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmn.b $2, $1", "dat.f $1, $0"},
			output: []string{"jmn.b $2, $1", "dat.f $1, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmn.ab $2, $1", "dat.f $1, $0"},
			output: []string{"jmn.ab $2, $1", "dat.f $1, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmn.f $2, $1", "dat.f $0, $0"},
			output: []string{"jmn.f $2, $1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmn.x $2, $1", "dat.f $0, $0"},
			output: []string{"jmn.x $2, $1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"jmn.i $2, $1", "dat.f $0, $0"},
			output: []string{"jmn.i $2, $1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		// non-zero PC
		{
			input:  []string{"dat.f $0, $0", "jmn.i $2, $0"},
			output: []string{"dat.f $0, $0", "jmn.i $2, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{3},
			offset: 1,
		},
	}
	runTests(t, "jmn", tests)
}

func TestDNJ(t *testing.T) {
	tests := []redcodeTest{
		// positive cases all modes
		{
			input:  []string{"djn.a $2, $1", "dat.f $0, $1"},
			output: []string{"djn.a $2, $1", "dat.f $-1, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"djn.ba $2, $1", "dat.f $0, $1"},
			output: []string{"djn.ba $2, $1", "dat.f $-1, $1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"djn.b $2, $1", "dat.f $1, $0"},
			output: []string{"djn.b $2, $1", "dat.f $1, $-1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"djn.ab $2, $1", "dat.f $1, $0"},
			output: []string{"djn.ab $2, $1", "dat.f $1, $-1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"djn.f $2, $1", "dat.f $1, $0"},
			output: []string{"djn.f $2, $1", "dat.f $0, $-1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"djn.x $2, $1", "dat.f $1, $0"},
			output: []string{"djn.x $2, $1", "dat.f $0, $-1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"djn.i $2, $1", "dat.f $1, $0"},
			output: []string{"djn.i $2, $1", "dat.f $0, $-1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		// negative
		{
			input:  []string{"djn.a $2, $1", "dat.f $1, $0"},
			output: []string{"djn.a $2, $1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"djn.ba $2, $1", "dat.f $1, $0"},
			output: []string{"djn.ba $2, $1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"djn.b $2, $1", "dat.f $0, $1"},
			output: []string{"djn.b $2, $1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"djn.ab $2, $1", "dat.f $0, $1"},
			output: []string{"djn.ab $2, $1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"djn.f $2, $1", "dat.f $1, $1"},
			output: []string{"djn.f $2, $1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"djn.x $2, $1", "dat.f $1, $1"},
			output: []string{"djn.x $2, $1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"djn.i $2, $1", "dat.f $1, $1"},
			output: []string{"djn.i $2, $1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		// non-zero PC
		{
			input:  []string{"dat.f $0, $0", "djn.i $2, $1"},
			output: []string{"dat.f $0, $0", "djn.i $2, $1", "dat.f $-1, $-1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{3},
			offset: 1,
		},
	}

	runTests(t, "djn", tests)
}

func TestCMP(t *testing.T) {
	tests := []redcodeTest{
		// positive cases all modes
		{
			input:  []string{"cmp.a $1, $2", "dat.f $1, $2", "dat.f $1, $4"},
			output: []string{"cmp.a $1, $2", "dat.f $1, $2", "dat.f $1, $4", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"cmp.b $1, $2", "dat.f $1, $2", "dat.f $3, $2"},
			output: []string{"cmp.b $1, $2", "dat.f $1, $2", "dat.f $3, $2", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"cmp.ab $1, $2", "dat.f $1, $2", "dat.f $3, $1"},
			output: []string{"cmp.ab $1, $2", "dat.f $1, $2", "dat.f $3, $1", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"cmp.ba $1, $2", "dat.f $1, $2", "dat.f $2, $4"},
			output: []string{"cmp.ba $1, $2", "dat.f $1, $2", "dat.f $2, $4", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"cmp.f $1, $2", "dat.f $1, $2", "dat.f $1, $2"},
			output: []string{"cmp.f $1, $2", "dat.f $1, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"cmp.x $1, $2", "dat.f $1, $2", "dat.f $2, $1"},
			output: []string{"cmp.x $1, $2", "dat.f $1, $2", "dat.f $2, $1", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"cmp.i $1, $2", "dat.f $1, $2", "dat.f $1, $2"},
			output: []string{"cmp.i $1, $2", "dat.f $1, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"cmp.i $1, $2", "dat.f $1, $2", "dat.f $1, $2"},
			output: []string{"cmp.i $1, $2", "dat.f $1, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		// negative cases all modes
		{
			input:  []string{"cmp.a $1, $2", "dat.f $1, $2", "dat.f $3, $2"},
			output: []string{"cmp.a $1, $2", "dat.f $1, $2", "dat.f $3, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"cmp.b $1, $2", "dat.f $1, $2", "dat.f $1, $4"},
			output: []string{"cmp.b $1, $2", "dat.f $1, $2", "dat.f $1, $4", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"cmp.ab $1, $2", "dat.f $1, $1", "dat.f $1, $4"},
			output: []string{"cmp.ab $1, $2", "dat.f $1, $1", "dat.f $1, $4", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"cmp.ba $1, $2", "dat.f $1, $2", "dat.f $3, $1"},
			output: []string{"cmp.ba $1, $2", "dat.f $1, $2", "dat.f $3, $1", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"cmp.f $1, $2", "dat.f $1, $2", "dat.f $1, $3"},
			output: []string{"cmp.f $1, $2", "dat.f $1, $2", "dat.f $1, $3", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"cmp.f $1, $2", "dat.f $1, $2", "dat.f $3, $2"},
			output: []string{"cmp.f $1, $2", "dat.f $1, $2", "dat.f $3, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"cmp.x $1, $2", "dat.f $1, $2", "dat.f $3, $1"},
			output: []string{"cmp.x $1, $2", "dat.f $1, $2", "dat.f $3, $1", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"cmp.x $1, $2", "dat.f $1, $2", "dat.f $2, $4"},
			output: []string{"cmp.x $1, $2", "dat.f $1, $2", "dat.f $2, $4", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"cmp.i $1, $2", "add.f $1, $2", "dat.f $1, $2"},
			output: []string{"cmp.i $1, $2", "add.f $1, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"cmp.i $1, $2", "dat.a $1, $2", "dat.f $1, $2"},
			output: []string{"cmp.i $1, $2", "dat.a $1, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"cmp.i $1, $2", "dat.f #1, $2", "dat.f $1, $2"},
			output: []string{"cmp.i $1, $2", "dat.f #1, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"cmp.i $1, $2", "dat.f $2, $2", "dat.f $1, $2"},
			output: []string{"cmp.i $1, $2", "dat.f $2, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"cmp.i $1, $2", "dat.f $1, #2", "dat.f $1, $2"},
			output: []string{"cmp.i $1, $2", "dat.f $1, #2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"cmp.i $1, $2", "dat.f $1, $1", "dat.f $1, $2"},
			output: []string{"cmp.i $1, $2", "dat.f $1, $1", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
	}
	runTests(t, "cmp", tests)
}

func TestSEQ(t *testing.T) {
	tests := []redcodeTest{
		// positive cases all modes
		{
			input:  []string{"seq.a $1, $2", "dat.f $1, $2", "dat.f $1, $4"},
			output: []string{"seq.a $1, $2", "dat.f $1, $2", "dat.f $1, $4", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"seq.b $1, $2", "dat.f $1, $2", "dat.f $3, $2"},
			output: []string{"seq.b $1, $2", "dat.f $1, $2", "dat.f $3, $2", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"seq.ab $1, $2", "dat.f $1, $2", "dat.f $3, $1"},
			output: []string{"seq.ab $1, $2", "dat.f $1, $2", "dat.f $3, $1", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"seq.ba $1, $2", "dat.f $1, $2", "dat.f $2, $4"},
			output: []string{"seq.ba $1, $2", "dat.f $1, $2", "dat.f $2, $4", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"seq.f $1, $2", "dat.f $1, $2", "dat.f $1, $2"},
			output: []string{"seq.f $1, $2", "dat.f $1, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"seq.x $1, $2", "dat.f $1, $2", "dat.f $2, $1"},
			output: []string{"seq.x $1, $2", "dat.f $1, $2", "dat.f $2, $1", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"seq.i $1, $2", "dat.f $1, $2", "dat.f $1, $2"},
			output: []string{"seq.i $1, $2", "dat.f $1, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"seq.i $1, $2", "dat.f $1, $2", "dat.f $1, $2"},
			output: []string{"seq.i $1, $2", "dat.f $1, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		// negative cases all modes
		{
			input:  []string{"seq.a $1, $2", "dat.f $1, $2", "dat.f $3, $2"},
			output: []string{"seq.a $1, $2", "dat.f $1, $2", "dat.f $3, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"seq.b $1, $2", "dat.f $1, $2", "dat.f $1, $4"},
			output: []string{"seq.b $1, $2", "dat.f $1, $2", "dat.f $1, $4", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"seq.ab $1, $2", "dat.f $1, $1", "dat.f $1, $4"},
			output: []string{"seq.ab $1, $2", "dat.f $1, $1", "dat.f $1, $4", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"seq.ba $1, $2", "dat.f $1, $2", "dat.f $3, $1"},
			output: []string{"seq.ba $1, $2", "dat.f $1, $2", "dat.f $3, $1", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"seq.f $1, $2", "dat.f $1, $2", "dat.f $1, $3"},
			output: []string{"seq.f $1, $2", "dat.f $1, $2", "dat.f $1, $3", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"seq.f $1, $2", "dat.f $1, $2", "dat.f $3, $2"},
			output: []string{"seq.f $1, $2", "dat.f $1, $2", "dat.f $3, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"seq.x $1, $2", "dat.f $1, $2", "dat.f $3, $1"},
			output: []string{"seq.x $1, $2", "dat.f $1, $2", "dat.f $3, $1", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"seq.x $1, $2", "dat.f $1, $2", "dat.f $2, $4"},
			output: []string{"seq.x $1, $2", "dat.f $1, $2", "dat.f $2, $4", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"seq.i $1, $2", "add.f $1, $2", "dat.f $1, $2"},
			output: []string{"seq.i $1, $2", "add.f $1, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"seq.i $1, $2", "dat.a $1, $2", "dat.f $1, $2"},
			output: []string{"seq.i $1, $2", "dat.a $1, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"seq.i $1, $2", "dat.f #1, $2", "dat.f $1, $2"},
			output: []string{"seq.i $1, $2", "dat.f #1, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"seq.i $1, $2", "dat.f $2, $2", "dat.f $1, $2"},
			output: []string{"seq.i $1, $2", "dat.f $2, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"seq.i $1, $2", "dat.f $1, #2", "dat.f $1, $2"},
			output: []string{"seq.i $1, $2", "dat.f $1, #2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"seq.i $1, $2", "dat.f $1, $1", "dat.f $1, $2"},
			output: []string{"seq.i $1, $2", "dat.f $1, $1", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},

		// non-zero pc
		{
			input:    []string{"dat.f $0, $0", "seq.i $145, $140"},
			output:   []string{"dat.f $0, $0", "seq.i $145, $140"},
			pq:       []Address{3},
			coresize: 8000,
			offset:   1,
		},
		{
			input:    []string{"dat.f $1, $1", "seq.i $-1, $140"},
			output:   []string{"dat.f $1, $1", "seq.i $-1, $140"},
			pq:       []Address{2},
			coresize: 8000,
			offset:   1,
		},
	}
	runTests(t, "seq", tests)
}

func TestSNE(t *testing.T) {
	tests := []redcodeTest{
		// positive cases all modes
		{
			input:  []string{"sne.a $1, $2", "dat.f $1, $2", "dat.f $3, $2"},
			output: []string{"sne.a $1, $2", "dat.f $1, $2", "dat.f $3, $2", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"sne.b $1, $2", "dat.f $1, $2", "dat.f $1, $4"},
			output: []string{"sne.b $1, $2", "dat.f $1, $2", "dat.f $1, $4", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"sne.ab $1, $2", "dat.f $1, $2", "dat.f $2, $4"},
			output: []string{"sne.ab $1, $2", "dat.f $1, $2", "dat.f $2, $4", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"sne.ba $1, $2", "dat.f $1, $2", "dat.f $3, $1"},
			output: []string{"sne.ba $1, $2", "dat.f $1, $2", "dat.f $3, $1", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"sne.f $1, $2", "dat.f $1, $2", "dat.f $3, $2"},
			output: []string{"sne.f $1, $2", "dat.f $1, $2", "dat.f $3, $2", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"sne.f $1, $2", "dat.f $1, $2", "dat.f $1, $4"},
			output: []string{"sne.f $1, $2", "dat.f $1, $2", "dat.f $1, $4", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"sne.x $1, $2", "dat.f $1, $2", "dat.f $1, $1"},
			output: []string{"sne.x $1, $2", "dat.f $1, $2", "dat.f $1, $1", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"sne.x $1, $2", "dat.f $1, $2", "dat.f $2, $2"},
			output: []string{"sne.x $1, $2", "dat.f $1, $2", "dat.f $2, $2", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"sne.i $1, $2", "add.f $1, $2", "dat.f $1, $2"},
			output: []string{"sne.i $1, $2", "add.f $1, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"sne.i $1, $2", "dat.f #1, $2", "dat.f $1, $2"},
			output: []string{"sne.i $1, $2", "dat.f #1, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"sne.i $1, $2", "dat.f $2, $2", "dat.f $1, $2"},
			output: []string{"sne.i $1, $2", "dat.f $2, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"sne.i $1, $2", "dat.f $1, #2", "dat.f $1, $2"},
			output: []string{"sne.i $1, $2", "dat.f $1, #2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"sne.i $1, $2", "dat.f $1, $1", "dat.f $1, $2"},
			output: []string{"sne.i $1, $2", "dat.f $1, $1", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{2},
		},

		// negative cases all modes
		{
			input:  []string{"sne.a $1, $2", "dat.f $1, $2", "dat.f $1, $4"},
			output: []string{"sne.a $1, $2", "dat.f $1, $2", "dat.f $1, $4", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"sne.b $1, $2", "dat.f $1, $2", "dat.f $3, $2"},
			output: []string{"sne.b $1, $2", "dat.f $1, $2", "dat.f $3, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"sne.ab $1, $2", "dat.f $1, $2", "dat.f $3, $1"},
			output: []string{"sne.ab $1, $2", "dat.f $1, $2", "dat.f $3, $1", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"sne.ba $1, $2", "dat.f $1, $2", "dat.f $2, $4"},
			output: []string{"sne.ba $1, $2", "dat.f $1, $2", "dat.f $2, $4", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"sne.f $1, $2", "dat.f $1, $2", "dat.f $1, #2"},
			output: []string{"sne.f $1, $2", "dat.f $1, $2", "dat.f $1, #2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"sne.x $1, $2", "dat.f $1, $2", "dat.f $2, #1"},
			output: []string{"sne.x $1, $2", "dat.f $1, $2", "dat.f $2, #1", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"sne.i $1, $2", "dat.f $1, $2", "dat.f $1, $2"},
			output: []string{"sne.i $1, $2", "dat.f $1, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
	}
	runTests(t, "seq", tests)
}

func TestSLT(t *testing.T) {
	tests := []redcodeTest{
		// positive cases for all modes
		{
			input:  []string{"slt.a $1, $2", "dat.f $1, $2", "dat.f $2, $1"},
			output: []string{"slt.a $1, $2", "dat.f $1, $2", "dat.f $2, $1", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"slt.b $1, $2", "dat.f $2, $1", "dat.f $1, $2"},
			output: []string{"slt.b $1, $2", "dat.f $2, $1", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"slt.ab $1, $2", "dat.f $1, $2", "dat.f $1, $2"},
			output: []string{"slt.ab $1, $2", "dat.f $1, $2", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"slt.ba $1, $2", "dat.f $2, $1", "dat.f $2, $1"},
			output: []string{"slt.ba $1, $2", "dat.f $2, $1", "dat.f $2, $1", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"slt.f $1, $2", "dat.f $0, $2", "dat.f $1, $3"},
			output: []string{"slt.f $1, $2", "dat.f $0, $2", "dat.f $1, $3", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"slt.i $1, $2", "dat.f $0, $2", "dat.f $1, $3"},
			output: []string{"slt.i $1, $2", "dat.f $0, $2", "dat.f $1, $3", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		{
			input:  []string{"slt.x $1, $2", "dat.f $0, $2", "dat.f $3, $1"},
			output: []string{"slt.x $1, $2", "dat.f $0, $2", "dat.f $3, $1", "dat.f $0, $0"},
			pq:     []Address{2},
		},
		// negative cases for all modes
		{
			input:  []string{"slt.a $1, $2", "dat.f $1, $2", "dat.f $1, $3"},
			output: []string{"slt.a $1, $2", "dat.f $1, $2", "dat.f $1, $3", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"slt.b $1, $2", "dat.f $2, $1", "dat.f $3, $0"},
			output: []string{"slt.b $1, $2", "dat.f $2, $1", "dat.f $3, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"slt.ab $1, $2", "dat.f $1, $2", "dat.f $3, $1"},
			output: []string{"slt.ab $1, $2", "dat.f $1, $2", "dat.f $3, $1", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"slt.ba $1, $2", "dat.f $1, $3", "dat.f $2, $1"},
			output: []string{"slt.ba $1, $2", "dat.f $1, $3", "dat.f $2, $1", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"slt.f $1, $2", "dat.f $1, $2", "dat.f $2, $2"},
			output: []string{"slt.f $1, $2", "dat.f $1, $2", "dat.f $2, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"slt.f $1, $2", "dat.f $2, $1", "dat.f $2, $2"},
			output: []string{"slt.f $1, $2", "dat.f $2, $1", "dat.f $2, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"slt.i $1, $2", "dat.f $1, $2", "dat.f $2, $2"},
			output: []string{"slt.i $1, $2", "dat.f $1, $2", "dat.f $2, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"slt.i $1, $2", "dat.f $2, $1", "dat.f $2, $2"},
			output: []string{"slt.i $1, $2", "dat.f $2, $1", "dat.f $2, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"slt.x $1, $2", "dat.f $1, $1", "dat.f $1, $2"},
			output: []string{"slt.x $1, $2", "dat.f $1, $1", "dat.f $1, $2", "dat.f $0, $0"},
			pq:     []Address{1},
		},
		{
			input:  []string{"slt.x $1, $2", "dat.f $1, $1", "dat.f $2, $1"},
			output: []string{"slt.x $1, $2", "dat.f $1, $1", "dat.f $2, $1", "dat.f $0, $0"},
			pq:     []Address{1},
		},
	}
	runTests(t, "slt", tests)
}

func TestSPL(t *testing.T) {
	tests := []redcodeTest{
		{
			input:  []string{"spl.b $0, $0"},
			output: []string{"spl.b $0, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1, 0},
		},
		{
			input:  []string{"spl.b $0, $1"},
			output: []string{"spl.b $0, $1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1, 0},
		},
		{
			input:  []string{"spl.b <0, $1"},
			output: []string{"spl.b <0, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1, 0},
		},
		{
			input:  []string{"spl.b <1, $0"},
			output: []string{"spl.b <1, $0", "dat.f $0, $-1", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1, 0},
		},
		{
			input:  []string{"spl.b <0, $0"},
			output: []string{"spl.b <0, $-1", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1, 3},
		},
		// non-zero PC
		{
			input:  []string{"dat.f $0, $0", "spl.b $0, $0"},
			output: []string{"dat.f $0, $0", "spl.b $0, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2, 1},
			offset: 1,
		},
		{
			input:  []string{"dat.f $0, $0", "spl.b $2, $0"},
			output: []string{"dat.f $0, $0", "spl.b $2, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2, 3},
			offset: 1,
		},
		{
			input:  []string{"dat.f $0, $0", "spl.b $-1, $0"},
			output: []string{"dat.f $0, $0", "spl.b $-1, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{2, 0},
			offset: 1,
		},
		{
			input:    []string{"dat.f $10, $0", "spl.b *-1, $0"},
			output:   []string{"dat.f $10, $0", "spl.b *-1, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:       []Address{2, 10},
			offset:   1,
			coresize: 20,
		},
	}
	runTests(t, "spl", tests)
}

func TestNOP(t *testing.T) {
	tests := []redcodeTest{
		{
			input:  []string{"nop.b $0, $0"},
			output: []string{"nop.b $0, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0"},
			pq:     []Address{1},
		},
	}
	runTests(t, "nop", tests)
}

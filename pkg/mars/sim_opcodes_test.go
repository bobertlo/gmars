package mars

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type opcodeTest struct {
	input     []string
	output    []string
	coresize  Address
	processes Address
	start     Address
	turns     int
	pq        []Address
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
	case '@':
		mode = B_INDIRECT
	case '<':
		mode = B_DECREMENT
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
		mval = M - mval
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

func runTests(t *testing.T, set_name string, tests []opcodeTest) {
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

		sim := NewSimulator(coresize, processes, 1, coresize, coresize, false)
		_, err := sim.SpawnWarrior(&WarriorData{Code: code}, 0)
		require.NoError(t, err)

		for i := 0; i < turns; i++ {
			sim.run_turn()
		}

		for i, expected := range expectedOutput {
			require.Equal(t, expected, sim.mem[i])
		}
	}

}

func TestMov(t *testing.T) {
	tests := []opcodeTest{
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
	}
	runTests(t, "mov", tests)
}

func TestMovAB_Direct(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     MOV,
				OpMode: AB,
				AMode:  IMMEDIATE,
				A:      1,
				BMode:  DIRECT,
				B:      1,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, sim.mem[1], Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     3,
		BMode: IMMEDIATE,
		B:     1,
	})
}

func TestAddA(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     ADD,
				OpMode: A,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     8,
		BMode: IMMEDIATE,
		B:     6,
	}, sim.mem[2])
}

func TestAddB(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     ADD,
				OpMode: B,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     5,
		BMode: IMMEDIATE,
		B:     10,
	}, sim.mem[2])
}

func TestAddAB(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     ADD,
				OpMode: AB,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     5,
		BMode: IMMEDIATE,
		B:     9,
	}, sim.mem[2])
}

func TestAddBA(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     ADD,
				OpMode: BA,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     9,
		BMode: IMMEDIATE,
		B:     6,
	}, sim.mem[2])
}

func TestAddI(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     ADD,
				OpMode: I,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     8,
		BMode: IMMEDIATE,
		B:     10,
	}, sim.mem[2])
}

func TestAddF(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     ADD,
				OpMode: F,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     8,
		BMode: IMMEDIATE,
		B:     10,
	}, sim.mem[2])
}

func TestAddX(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     ADD,
				OpMode: X,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     5,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     9,
		BMode: IMMEDIATE,
		B:     8,
	}, sim.mem[2])
}

func TestSubA(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     SUB,
				OpMode: A,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     2,
		BMode: IMMEDIATE,
		B:     6,
	}, sim.mem[2])
}

func TestSubB(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     SUB,
				OpMode: B,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     5,
		BMode: IMMEDIATE,
		B:     2,
	}, sim.mem[2])
}

func TestSubAB(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     SUB,
				OpMode: AB,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     5,
		BMode: IMMEDIATE,
		B:     3,
	}, sim.mem[2])
}

func TestSubBA(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     SUB,
				OpMode: BA,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     1,
		BMode: IMMEDIATE,
		B:     6,
	}, sim.mem[2])
}

func TestSubI(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     SUB,
				OpMode: I,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     2,
		BMode: IMMEDIATE,
		B:     2,
	}, sim.mem[2])
}

func TestSubF(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     SUB,
				OpMode: F,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     2,
		BMode: IMMEDIATE,
		B:     2,
	}, sim.mem[2])
}

func TestSubX(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     SUB,
				OpMode: X,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     1,
		BMode: IMMEDIATE,
		B:     3,
	}, sim.mem[2])
}

func TestMulA(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     MUL,
				OpMode: A,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     15,
		BMode: IMMEDIATE,
		B:     6,
	}, sim.mem[2])
}

func TestMulB(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     MUL,
				OpMode: B,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     5,
		BMode: IMMEDIATE,
		B:     24,
	}, sim.mem[2])
}

func TestMulAB(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     MUL,
				OpMode: AB,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     5,
		BMode: IMMEDIATE,
		B:     18,
	}, sim.mem[2])
}

func TestMulBA(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     MUL,
				OpMode: BA,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     20,
		BMode: IMMEDIATE,
		B:     6,
	}, sim.mem[2])
}

func TestMulI(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     MUL,
				OpMode: I,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     15,
		BMode: IMMEDIATE,
		B:     24,
	}, sim.mem[2])
}

func TestMulF(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     MUL,
				OpMode: F,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     15,
		BMode: IMMEDIATE,
		B:     24,
	}, sim.mem[2])
}

func TestMulX(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     MUL,
				OpMode: X,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     20,
		BMode: IMMEDIATE,
		B:     18,
	}, sim.mem[2])
}

package gmars

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type warriorTestCase struct {
	filename     string
	loadFilename string
	output       WarriorData
	config       SimulatorConfig
	err          bool
}

func runWarriorTests(t *testing.T, tests []warriorTestCase) {
	for _, test := range tests {
		input, err := os.Open(test.filename)
		require.NoError(t, err)

		warriorData, err := CompileWarrior(input, test.config)
		if test.err {
			assert.Error(t, err, fmt.Sprintf("%s: error should be present", test.filename))
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.output, warriorData)
		}
	}
}

func runWarriorLoadFileTests(t *testing.T, tests []warriorTestCase) {
	for _, test := range tests {
		input, err := os.Open(test.filename)
		require.NoError(t, err)
		defer input.Close()

		warriorData, err := CompileWarrior(input, test.config)
		if test.err {
			assert.Error(t, err, fmt.Sprintf("%s: error should be present", test.filename))
		} else {
			require.NoError(t, err)
			loadInput, err := os.Open(test.loadFilename)
			require.NoError(t, err)
			defer loadInput.Close()
			expectedData, err := ParseLoadFile(loadInput, test.config)
			require.NoError(t, err)
			assert.Equal(t, expectedData.Code, warriorData.Code)
		}
	}
}

func TestCompileWarriors88(t *testing.T) {
	config := ConfigKOTH88()
	tests := []warriorTestCase{
		{
			filename: "warriors/88/imp.red",
			config:   config,
			output: WarriorData{
				Name:     "Imp",
				Author:   "A K Dewdney",
				Strategy: "this is the simplest program\nit was described in the initial articles\n",
				Start:    0,
				Code: []Instruction{
					{Op: MOV, OpMode: I, AMode: DIRECT, A: 0, BMode: DIRECT, B: 1},
				},
			},
		},
	}

	runWarriorTests(t, tests)
}

func TestCompileWarriors94(t *testing.T) {
	config := ConfigNOP94()
	tests := []warriorTestCase{
		{
			filename: "warriors/94/imp.red",
			config:   config,
			output: WarriorData{
				Name:     "Imp",
				Author:   "A K Dewdney",
				Strategy: "this is the simplest program\nit was described in the initial articles\n",
				Start:    0,
				Code: []Instruction{
					{Op: MOV, OpMode: I, AMode: IMMEDIATE, A: 0, BMode: DIRECT, B: 1},
				},
			},
		},
	}
	runWarriorTests(t, tests)
}

func TestCompileWarriorsFile94(t *testing.T) {
	config := ConfigNOP94()
	tests := []warriorTestCase{
		{
			filename:     "warriors/94/simpleshot.red",
			loadFilename: "test_files/simpleshot.rc",
			config:       config,
		},
		{
			filename:     "warriors/94/scaryvampire.red",
			loadFilename: "test_files/scaryvampire.rc",
			config:       config,
		},
	}

	runWarriorLoadFileTests(t, tests)
}

func TestCompileForLoop(t *testing.T) {
	config := ConfigNOP94()

	input := `
	dat 123, 123
	i j for 3
	dat i, j
	rof
	dat 123, 123	
`

	// lexer := newLexer(strings.NewReader(input))
	// parser := newParser(lexer)
	// src, data, err := parser.parse()
	// require.NoError(t, err)
	// compiler, err := newCompiler(src, data, config)

	// compiler.compile()
	fmt.Println("test start")
	w, err := CompileWarrior(strings.NewReader(input), config)
	require.NoError(t, err)
	assert.Equal(t, []Instruction{
		{Op: DAT, OpMode: F, AMode: DIRECT, A: 123, BMode: DIRECT, B: 123},
		{Op: DAT, OpMode: F, AMode: DIRECT, A: 0, BMode: DIRECT, B: 1},
		{Op: DAT, OpMode: F, AMode: DIRECT, A: 7999, BMode: DIRECT, B: 2},
		{Op: DAT, OpMode: F, AMode: DIRECT, A: 7998, BMode: DIRECT, B: 3},
		{Op: DAT, OpMode: F, AMode: DIRECT, A: 123, BMode: DIRECT, B: 123},
	}, w.Code)
	fmt.Println(w.Code)
}

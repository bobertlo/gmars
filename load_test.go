package gmars

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type instructionParseTestCase struct {
	input  string
	output Instruction
}

func TestLoadImp88(t *testing.T) {
	config := ConfigKOTH88

	reader := bytes.NewReader(imp_88_red)
	data, err := ParseLoadFile(reader, config)
	require.NoError(t, err)
	require.Equal(t, "Imp", data.Name)
	require.Equal(t, "A K Dewdney", data.Author)
	require.Equal(t, "this is the simplest program\nit was described in the initial articles\n", data.Strategy)
	require.Equal(t, 0, data.Start)
	require.Equal(t, 1, len(data.Code))
	require.Equal(t, Instruction{Op: MOV, OpMode: I, AMode: DIRECT, A: 0, BMode: DIRECT, B: 1}, data.Code[0])
}

func TestLoadImp94(t *testing.T) {
	config := ConfigNOP94

	reader := bytes.NewReader(imp_94_red)
	data, err := ParseLoadFile(reader, config)
	require.NoError(t, err)
	require.Equal(t, "Imp", data.Name)
	require.Equal(t, "A K Dewdney", data.Author)
	require.Equal(t, "this is the simplest program\nit was described in the initial articles\n", data.Strategy)
	require.Equal(t, 0, data.Start)
	require.Equal(t, 1, len(data.Code))
	require.Equal(t, Instruction{Op: MOV, OpMode: I, AMode: IMMEDIATE, A: 0, BMode: DIRECT, B: 1}, data.Code[0])
}

func TestLoadDwarf(t *testing.T) {
	config := ConfigKOTH88

	dwarf_code := `
	ADD #  4,  $  3
	MOV $  2,  @  2
	JMP $ -2,  $  0
	DAT #  0,  #  0
	`

	reader := strings.NewReader(dwarf_code)
	data, err := ParseLoadFile(reader, config)
	require.NoError(t, err)
	require.Equal(t, 0, data.Start)
	require.Equal(t, 4, len(data.Code))
	require.Equal(t, []Instruction{
		{Op: ADD, OpMode: AB, AMode: IMMEDIATE, A: 4, BMode: DIRECT, B: 3},
		{Op: MOV, OpMode: I, AMode: DIRECT, A: 2, BMode: B_INDIRECT, B: 2},
		{Op: JMP, OpMode: B, AMode: DIRECT, A: 8000 - 2, BMode: DIRECT, B: 0},
		{Op: DAT, OpMode: F, AMode: IMMEDIATE, A: 0, BMode: IMMEDIATE, B: 0},
	}, data.Code)
}

func TestValidInput(t *testing.T) {
	// random inputs that are valid but not worth validating output
	cases := []string{
		"END\n",
		"\n\n",
	}

	config := ConfigKOTH88
	for i, testCase := range cases {
		reader := strings.NewReader(testCase)
		_, err := ParseLoadFile(reader, config)
		assert.NoError(t, err, "test: %d' '%s'", i, testCase)
	}
}

func TestInvalidInput(t *testing.T) {
	// random inputs that will throw an error
	cases := []string{
		// random invalid token combinations
		"END invalid ;\n",
		"END 1 2\n",
		"OTHER ; bad short op code\n",
		"CMP $0, $ 0 ; missing space\n",
		"CMP $ 0, $0 ; missing space\n",
		"CMP $ 0 $ 0 ; no comma\n",
		"INV $ 0, $ 0 ; invalid opcode\n",
		// bad op address modes
		"DAT $ 0, # 0\n",
		"DAT @ 0, # 0\n",
		"DAT # 0, $ 0\n",
		"DAT # 0, @ 0\n",
		"CMP $ 0, # 0\n",
		"SLT $ 0, # 0\n",
		"ADD $ 0, # 0\n",
		"SUB $ 0, # 0\n",
		"JMP # 0, $ 0\n",
		"JMN # 0, $ 0\n",
		"JMZ # 0, $ 0\n",
		"DJN # 0, $ 0\n",
		"SPL # 0, $ 0\n",
		"MOV $ 0, $ 1\nEND 2 ; BAD END ADDRESS\n",
		"MOV $ 0, $ 1\nEND -2 ; BAD END ADDRESS\n",
		// invalid addresses and modes
		"MOV ! 0, $ 1\n",
		"MOV $ 0, ! 1\n",
		"MOV $ oops, $ 1\n",
		"MOV $ 0, $ ooops\n",
	}

	config := ConfigKOTH88

	for i, testCase := range cases {
		reader := strings.NewReader(testCase)
		out, err := ParseLoadFile(reader, config)
		assert.Error(t, err, fmt.Sprintf("test %d: '%s'", i, testCase))
		assert.Equal(t, 0, len(out.Code))
	}
}

func TestValidOpModeCombos88(t *testing.T) {
	// these are all valid op mode combinations
	// any op not included falls through to one of these

	testCases := []instructionParseTestCase{
		// DAT is special!
		{"DAT # 1, # 2\n", Instruction{Op: DAT, OpMode: F, AMode: IMMEDIATE, A: 1, BMode: IMMEDIATE, B: 2}},
		{"DAT # 1, < 2\n", Instruction{Op: DAT, OpMode: F, AMode: IMMEDIATE, A: 1, BMode: B_DECREMENT, B: 2}},
		{"DAT < 1, # 2\n", Instruction{Op: DAT, OpMode: F, AMode: B_DECREMENT, A: 1, BMode: IMMEDIATE, B: 2}},
		{"DAT < 1, < 2\n", Instruction{Op: DAT, OpMode: F, AMode: B_DECREMENT, A: 1, BMode: B_DECREMENT, B: 2}},

		// CMP, MOV
		{"MOV # 1, $ 2\n", Instruction{Op: MOV, OpMode: AB, AMode: IMMEDIATE, A: 1, BMode: DIRECT, B: 2}},
		{"MOV # 1, @ 2\n", Instruction{Op: MOV, OpMode: AB, AMode: IMMEDIATE, A: 1, BMode: B_INDIRECT, B: 2}},
		{"MOV # 1, < 2\n", Instruction{Op: MOV, OpMode: AB, AMode: IMMEDIATE, A: 1, BMode: B_DECREMENT, B: 2}},
		{"MOV $ 1, $ 2\n", Instruction{Op: MOV, OpMode: I, AMode: DIRECT, A: 1, BMode: DIRECT, B: 2}},
		{"MOV $ 1, @ 2\n", Instruction{Op: MOV, OpMode: I, AMode: DIRECT, A: 1, BMode: B_INDIRECT, B: 2}},
		{"MOV $ 1, < 2\n", Instruction{Op: MOV, OpMode: I, AMode: DIRECT, A: 1, BMode: B_DECREMENT, B: 2}},
		{"MOV @ 1, $ 2\n", Instruction{Op: MOV, OpMode: I, AMode: B_INDIRECT, A: 1, BMode: DIRECT, B: 2}},
		{"MOV @ 1, @ 2\n", Instruction{Op: MOV, OpMode: I, AMode: B_INDIRECT, A: 1, BMode: B_INDIRECT, B: 2}},
		{"MOV @ 1, < 2\n", Instruction{Op: MOV, OpMode: I, AMode: B_INDIRECT, A: 1, BMode: B_DECREMENT, B: 2}},
		{"MOV < 1, $ 2\n", Instruction{Op: MOV, OpMode: I, AMode: B_DECREMENT, A: 1, BMode: DIRECT, B: 2}},
		{"MOV < 1, @ 2\n", Instruction{Op: MOV, OpMode: I, AMode: B_DECREMENT, A: 1, BMode: B_INDIRECT, B: 2}},
		{"MOV < 1, < 2\n", Instruction{Op: MOV, OpMode: I, AMode: B_DECREMENT, A: 1, BMode: B_DECREMENT, B: 2}},

		// ADD, SUB
		{"ADD # 1, $ 2\n", Instruction{Op: ADD, OpMode: AB, AMode: IMMEDIATE, A: 1, BMode: DIRECT, B: 2}},
		{"ADD # 1, @ 2\n", Instruction{Op: ADD, OpMode: AB, AMode: IMMEDIATE, A: 1, BMode: B_INDIRECT, B: 2}},
		{"ADD # 1, < 2\n", Instruction{Op: ADD, OpMode: AB, AMode: IMMEDIATE, A: 1, BMode: B_DECREMENT, B: 2}},
		{"ADD $ 1, $ 2\n", Instruction{Op: ADD, OpMode: F, AMode: DIRECT, A: 1, BMode: DIRECT, B: 2}},
		{"ADD $ 1, @ 2\n", Instruction{Op: ADD, OpMode: F, AMode: DIRECT, A: 1, BMode: B_INDIRECT, B: 2}},
		{"ADD $ 1, < 2\n", Instruction{Op: ADD, OpMode: F, AMode: DIRECT, A: 1, BMode: B_DECREMENT, B: 2}},
		{"ADD @ 1, $ 2\n", Instruction{Op: ADD, OpMode: F, AMode: B_INDIRECT, A: 1, BMode: DIRECT, B: 2}},
		{"ADD @ 1, @ 2\n", Instruction{Op: ADD, OpMode: F, AMode: B_INDIRECT, A: 1, BMode: B_INDIRECT, B: 2}},
		{"ADD @ 1, < 2\n", Instruction{Op: ADD, OpMode: F, AMode: B_INDIRECT, A: 1, BMode: B_DECREMENT, B: 2}},
		{"ADD < 1, $ 2\n", Instruction{Op: ADD, OpMode: F, AMode: B_DECREMENT, A: 1, BMode: DIRECT, B: 2}},
		{"ADD < 1, @ 2\n", Instruction{Op: ADD, OpMode: F, AMode: B_DECREMENT, A: 1, BMode: B_INDIRECT, B: 2}},
		{"ADD < 1, < 2\n", Instruction{Op: ADD, OpMode: F, AMode: B_DECREMENT, A: 1, BMode: B_DECREMENT, B: 2}},

		// Just SLT
		{"SLT # 1, $ 2\n", Instruction{Op: SLT, OpMode: AB, AMode: IMMEDIATE, A: 1, BMode: DIRECT, B: 2}},
		{"SLT # 1, @ 2\n", Instruction{Op: SLT, OpMode: AB, AMode: IMMEDIATE, A: 1, BMode: B_INDIRECT, B: 2}},
		{"SLT # 1, < 2\n", Instruction{Op: SLT, OpMode: AB, AMode: IMMEDIATE, A: 1, BMode: B_DECREMENT, B: 2}},
		{"SLT $ 1, $ 2\n", Instruction{Op: SLT, OpMode: B, AMode: DIRECT, A: 1, BMode: DIRECT, B: 2}},
		{"SLT $ 1, @ 2\n", Instruction{Op: SLT, OpMode: B, AMode: DIRECT, A: 1, BMode: B_INDIRECT, B: 2}},
		{"SLT $ 1, < 2\n", Instruction{Op: SLT, OpMode: B, AMode: DIRECT, A: 1, BMode: B_DECREMENT, B: 2}},
		{"SLT @ 1, $ 2\n", Instruction{Op: SLT, OpMode: B, AMode: B_INDIRECT, A: 1, BMode: DIRECT, B: 2}},
		{"SLT @ 1, @ 2\n", Instruction{Op: SLT, OpMode: B, AMode: B_INDIRECT, A: 1, BMode: B_INDIRECT, B: 2}},
		{"SLT @ 1, < 2\n", Instruction{Op: SLT, OpMode: B, AMode: B_INDIRECT, A: 1, BMode: B_DECREMENT, B: 2}},
		{"SLT < 1, $ 2\n", Instruction{Op: SLT, OpMode: B, AMode: B_DECREMENT, A: 1, BMode: DIRECT, B: 2}},
		{"SLT < 1, @ 2\n", Instruction{Op: SLT, OpMode: B, AMode: B_DECREMENT, A: 1, BMode: B_INDIRECT, B: 2}},
		{"SLT < 1, < 2\n", Instruction{Op: SLT, OpMode: B, AMode: B_DECREMENT, A: 1, BMode: B_DECREMENT, B: 2}},

		// JMP, JMN, JMZ, DJN, SPL
		{"JMP $ 1, # 2\n", Instruction{Op: JMP, OpMode: B, AMode: DIRECT, A: 1, BMode: IMMEDIATE, B: 2}},
		{"JMP $ 1, $ 2\n", Instruction{Op: JMP, OpMode: B, AMode: DIRECT, A: 1, BMode: DIRECT, B: 2}},
		{"JMP $ 1, @ 2\n", Instruction{Op: JMP, OpMode: B, AMode: DIRECT, A: 1, BMode: B_INDIRECT, B: 2}},
		{"JMP $ 1, < 2\n", Instruction{Op: JMP, OpMode: B, AMode: DIRECT, A: 1, BMode: B_DECREMENT, B: 2}},
		{"JMP @ 1, # 2\n", Instruction{Op: JMP, OpMode: B, AMode: B_INDIRECT, A: 1, BMode: IMMEDIATE, B: 2}},
		{"JMP @ 1, $ 2\n", Instruction{Op: JMP, OpMode: B, AMode: B_INDIRECT, A: 1, BMode: DIRECT, B: 2}},
		{"JMP @ 1, @ 2\n", Instruction{Op: JMP, OpMode: B, AMode: B_INDIRECT, A: 1, BMode: B_INDIRECT, B: 2}},
		{"JMP @ 1, < 2\n", Instruction{Op: JMP, OpMode: B, AMode: B_INDIRECT, A: 1, BMode: B_DECREMENT, B: 2}},
		{"JMP < 1, # 2\n", Instruction{Op: JMP, OpMode: B, AMode: B_DECREMENT, A: 1, BMode: IMMEDIATE, B: 2}},
		{"JMP < 1, $ 2\n", Instruction{Op: JMP, OpMode: B, AMode: B_DECREMENT, A: 1, BMode: DIRECT, B: 2}},
		{"JMP < 1, @ 2\n", Instruction{Op: JMP, OpMode: B, AMode: B_DECREMENT, A: 1, BMode: B_INDIRECT, B: 2}},
		{"JMP < 1, < 2\n", Instruction{Op: JMP, OpMode: B, AMode: B_DECREMENT, A: 1, BMode: B_DECREMENT, B: 2}},
	}

	config := ConfigKOTH88

	for i, testCase := range testCases {
		reader := strings.NewReader(testCase.input)
		out, err := ParseLoadFile(reader, config)
		require.NoError(t, err, fmt.Sprintf("test %d: parsing '%s' failed: '%s", i, testCase.input, err))
		require.Equal(t, 1, len(out.Code), fmt.Sprintf("test %d: '%s'", i, testCase.input))
		assert.Equal(t, testCase.output, out.Code[0], fmt.Sprintf("test %d: '%s'", i, testCase.input))
	}
}

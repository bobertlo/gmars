package mars

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const imp88 = `;redcode
;name Imp
;author A K Dewdney
;strategy this is the simplest program
;strategy it was described in the initial articles

		MOV $ 0, $ 1
		END 0
`

func TestLoadImp(t *testing.T) {
	config := StandardConfig()

	reader := strings.NewReader(imp88)
	data, err := ParseLoadFile(reader, config)
	require.NoError(t, err)
	require.Equal(t, "Imp", data.Name)
	require.Equal(t, "A K Dewdney", data.Author)
	require.Equal(t, "this is the simplest program\nit was described in the initial articles\n", data.Strategy)
	require.Equal(t, 0, data.Start)
	require.Equal(t, 1, len(data.Code))
	require.Equal(t, Instruction{Op: MOV, OpMode: I, AMode: DIRECT, A: 0, BMode: DIRECT, B: 1}, data.Code[0])
}

func TestLoadDwarf(t *testing.T) {
	config := StandardConfig()

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

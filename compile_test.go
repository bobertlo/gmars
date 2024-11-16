package gmars

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type warriorTestCase struct {
	input  []byte
	name   string
	output WarriorData
	config SimulatorConfig
	err    bool
}

func runWarriorTests(t *testing.T, tests []warriorTestCase) {
	for _, test := range tests {
		warriorData, err := CompileWarrior(bytes.NewReader(test.input), test.config)
		if test.err {
			assert.Error(t, err, fmt.Sprintf("%s: error should be present", test.name))
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.output, warriorData)
		}
	}
}

func TestCompileWarriors88(t *testing.T) {
	config := ConfigKOTH88()
	tests := []warriorTestCase{
		{
			name:   "imp_88_red",
			input:  imp_88_red,
			config: config,
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

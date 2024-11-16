package gmars

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	l := newLexer(strings.NewReader("test equ 1\ntest2 equ test+start\nstart mov $0, $1"))
	p := newParser(l)
	source, err := p.parse()
	require.NoError(t, err)

	compiler, err := newCompiler(source, ConfigNOP94())
	require.NoError(t, err)

	out, err := compiler.compile()
	require.NoError(t, err)
	require.Equal(t, WarriorData{
		Code: []Instruction{
			{Op: MOV, OpMode: I, AMode: DIRECT, A: 0, BMode: DIRECT, B: 1},
		},
	}, out)
}

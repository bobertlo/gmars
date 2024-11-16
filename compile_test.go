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

	_, err = compiler.compile()
	require.Error(t, err)

	out, err := compiler.expandExpression([]token{{tokText, "start"}}, 1)
	require.NoError(t, err)
	require.Equal(t, []token{{tokNumber, "-1"}}, out)
}

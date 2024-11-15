package gmars

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	l := newLexer(strings.NewReader("test equ 1\nstart mov $0, $1"))
	p := newParser(l)
	source, err := p.parse()
	require.NoError(t, err)

	compiler, err := newCompiler(source, ConfigNOP94())
	require.NoError(t, err)

	_, err = compiler.compile()
	require.Error(t, err)
}

func TestContainsCycleNegative(t *testing.T) {
	refs := map[string][]string{
		"a": {"b"},
		"b": {"c", "d"},
		"c": {"d"},
	}

	cyclic, key := graphContainsCycle(refs)
	assert.False(t, cyclic)
	assert.Equal(t, "", key)
}

func TestContainsCyclePositive(t *testing.T) {
	refs := map[string][]string{
		"a": {"b"},
		"b": {"c", "d"},
		"c": {"b"},
	}
	cyclic, key := graphContainsCycle(refs)
	assert.True(t, cyclic)
	assert.Equal(t, "b", key)
}

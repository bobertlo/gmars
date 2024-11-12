package gmars

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	l := newLexer(strings.NewReader(";comment line\n"))
	p := newParser(l)

	source, err := p.parse()
	require.NoError(t, err)
	require.NotNil(t, source)

	require.Equal(t, 0, len(p.symbols))
	// require.Equal(t, 1, len(p.lines))
	require.Equal(t, []sourceLine{
		{
			line:     1,
			typ:      lineComment,
			comment:  ";comment line",
			newlines: 1,
		},
	}, p.lines)
}

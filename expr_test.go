package gmars

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpandExpressions(t *testing.T) {
	values := map[string][]token{
		"a": {{tokNumber, "1"}},
		"b": {{tokText, "a"}, {tokExprOp, "+"}, {tokNumber, "2"}},
		"c": {{tokText, "a"}, {tokExprOp, "*"}, {tokText, "b"}},
	}
	graph := map[string][]string{
		"b": {"a"},
		"c": {"b"},
	}

	output, err := expandExpressions(values, graph)
	require.NoError(t, err)
	require.Equal(t, map[string][]token{
		"a": {{tokNumber, "1"}},
		"b": {{tokNumber, "1"}, {tokExprOp, "+"}, {tokNumber, "2"}},
		"c": {{tokNumber, "1"}, {tokExprOp, "*"}, {tokNumber, "1"}, {tokExprOp, "+"}, {tokNumber, "2"}},
	}, output)
}

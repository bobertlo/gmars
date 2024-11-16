package gmars

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandExpressions(t *testing.T) {
	values := map[string][]token{
		"a": {{tokNumber, "1"}},
		"c": {{tokText, "a"}, {tokExprOp, "*"}, {tokText, "b"}},
		"b": {{tokText, "a"}, {tokExprOp, "+"}, {tokNumber, "2"}},
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

func TestEvaluateExpression(t *testing.T) {
	testCases := map[string]int{
		"1":       1,
		"2":       2,
		"-2":      -2,
		"1+2*3":   7,
		"1*2+3":   5,
		"(1+2)*3": 9,
	}

	for input, expected := range testCases {
		lexer := newLexer(strings.NewReader(input))
		tokens, err := lexer.Tokens()
		require.NoError(t, err)

		val, err := evaluateExpression(tokens)
		require.NoError(t, err)
		assert.Equal(t, expected, val)
	}
}

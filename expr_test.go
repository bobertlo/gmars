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
		"c": {{tokText, "a"}, {tokSymbol, "*"}, {tokText, "b"}},
		"b": {{tokText, "a"}, {tokSymbol, "+"}, {tokNumber, "2"}},
	}
	graph := map[string][]string{
		"b": {"a"},
		"c": {"b"},
	}

	output, err := expandExpressions(values, graph)
	require.NoError(t, err)
	require.Equal(t, map[string][]token{
		"a": {{tokNumber, "1"}},
		"b": {{tokNumber, "1"}, {tokSymbol, "+"}, {tokNumber, "2"}},
		"c": {{tokNumber, "1"}, {tokSymbol, "*"}, {tokNumber, "1"}, {tokSymbol, "+"}, {tokNumber, "2"}},
	}, output)
}

func TestCombineSigns(t *testing.T) {
	testCases := []struct {
		input  string
		output []token
	}{
		{
			input: "1++-2",
			output: []token{
				{tokNumber, "1"},
				{tokSymbol, "+"},
				{tokSymbol, "-"},
				{tokNumber, "2"},
			},
		},
		{
			input: "1-+-2",
			output: []token{
				{tokNumber, "1"},
				{tokSymbol, "-"},
				{tokSymbol, "-"},
				{tokNumber, "2"},
			},
		},
	}

	for _, test := range testCases {
		lexer := newLexer(strings.NewReader(test.input))
		tokens, err := lexer.Tokens()
		tokens = tokens[:len(tokens)-1]
		require.NoError(t, err)

		combinedTokens := combineSigns(tokens)
		require.Equal(t, test.output, combinedTokens)
	}
}

func TestFlipDoubleNegatives(t *testing.T) {
	testCases := []struct {
		input  string
		output []token
	}{
		{
			input: "1--1",
			output: []token{
				{tokNumber, "1"},
				{tokSymbol, "+"},
				{tokNumber, "1"},
			},
		},
	}

	for _, test := range testCases {
		lexer := newLexer(strings.NewReader(test.input))
		tokens, err := lexer.Tokens()
		tokens = tokens[:len(tokens)-1]
		require.NoError(t, err)

		combinedTokens := flipDoubleNegatives(tokens)
		require.Equal(t, test.output, combinedTokens)
	}
}

func TestEvaluateExpressionPositive(t *testing.T) {
	testCases := map[string]int{
		"1":       1,
		"2":       2,
		"-2":      -2,
		"1+2*3":   7,
		"1*2+3":   5,
		"(1+2)*3": 9,
		"1 * -1":  -1,
		"1 + -1":  0,

		// handle signs
		"1 - -1": 2,

		// logic
		"1 > 2":        0,
		"2 > 1":        1,
		"1 < 2":        1,
		"2 < 1":        0,
		"1 >= 1":       1,
		"2 <= 2":       1,
		"8000 == 8000": 1,
		"8000 == 800":  0,
		// hmmm, these need to be fixed
		// "1 && 1":           1,
		// "1 && 0":           0,
		// "1 || 1":           1,
		// "1 || 0":           0,
		"2 == 1 || 2 == 2": 1,
		"2 == 1 || 2 == 3": 0,
	}

	for input, expected := range testCases {
		lexer := newLexer(strings.NewReader(input))
		tokens, err := lexer.Tokens()
		require.NoError(t, err)

		// trim EOF from input
		tokens = tokens[:len(tokens)-1]

		val, err := evaluateExpression(tokens)
		require.NoError(t, err)
		assert.Equal(t, expected, val)
	}
}

func TestEvaluateExpressionNegative(t *testing.T) {
	cases := []string{
		")21",
		"2^3",
		"2{2",
	}

	for _, input := range cases {
		lexer := newLexer(strings.NewReader(input))
		tokens, err := lexer.Tokens()
		if err != nil {
			continue
		}

		val, err := evaluateExpression(tokens)
		assert.Error(t, err)
		assert.Equal(t, 0, val)
	}
}

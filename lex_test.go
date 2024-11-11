package gmars

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type lexTestCase struct {
	input    string
	expected []token
}

func runLexTests(t *testing.T, setName string, testCases []lexTestCase) {
	for i, test := range testCases {
		l := newLexer(strings.NewReader(test.input))
		out, err := l.Tokens()
		require.NoError(t, err, fmt.Errorf("%s test %d error: %s", setName, i, err))
		require.Equal(t, test.expected, out, fmt.Sprintf("%s test %d", setName, i))
	}
}

func TestLexer(t *testing.T) {
	testCases := []lexTestCase{
		{
			input: "\n",
			expected: []token{
				{typ: tokNewline},
				{typ: tokEOF},
			},
		},
		{
			input: "start mov # -1, $2 ; comment\n",
			expected: []token{
				{tokText, "start"},
				{tokText, "mov"},
				{tokAddressMode, "#"},
				{tokExprOp, "-"},
				{tokNumber, "1"},
				{tokComma, ","},
				{tokAddressMode, "$"},
				{tokNumber, "2"},
				{tokComment, "; comment"},
				{tokNewline, ""},
				{tokEOF, ""},
			},
		},
		{
			input: "step equ (1+3)-start\n",
			expected: []token{
				{tokText, "step"},
				{tokText, "equ"},
				{tokParenL, "("},
				{tokNumber, "1"},
				{tokExprOp, "+"},
				{tokNumber, "3"},
				{tokParenR, ")"},
				{tokExprOp, "-"},
				{tokText, "start"},
				{tokNewline, ""},
				{tokEOF, ""},
			},
		},
	}

	runLexTests(t, "TestLexer", testCases)
}

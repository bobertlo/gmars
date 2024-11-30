package gmars

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type forTestCase struct {
	input   string
	symbols map[string][]token
	output  []token
}

func runForExpanderTests(t *testing.T, cases []forTestCase) {
	for _, test := range cases {
		tokens, err := LexInput(strings.NewReader(test.input))
		require.NoError(t, err)
		require.NotNil(t, tokens)

		// scanner := newSymbolScanner(newBufTokenReader(tokens))
		expander := newForExpander(newBufTokenReader(tokens), test.symbols)
		outTokens, err := expander.Tokens()
		require.NoError(t, err)
		require.Equal(t, test.output, outTokens)
	}
}

func TestForExpander(t *testing.T) {
	tests := []forTestCase{
		{
			input: "i for 2\ndat 0, i\nrof\n",
			output: []token{
				{tokText, "dat"},
				{tokNumber, "0"},
				{tokComma, ","},
				{tokNumber, "1"},
				{tokNewline, ""},
				{tokText, "dat"},
				{tokNumber, "0"},
				{tokComma, ","},
				{tokNumber, "2"},
				{tokNewline, ""},
				{tokEOF, ""},
			},
		},
		{
			input: "i for 2\ndat 0, i\nrof\ndat 3, 4\n",
			output: []token{
				{tokText, "dat"},
				{tokNumber, "0"},
				{tokComma, ","},
				{tokNumber, "1"},
				{tokNewline, ""},
				{tokText, "dat"},
				{tokNumber, "0"},
				{tokComma, ","},
				{tokNumber, "2"},
				{tokNewline, ""},
				{tokText, "dat"},
				{tokNumber, "3"},
				{tokComma, ","},
				{tokNumber, "4"},
				{tokNewline, ""},
				{tokEOF, ""},
			},
		},
		// no for
		{
			input: "test equ 2\ndat 0, test\n",
			output: []token{
				{tokText, "test"},
				{tokText, "equ"},
				{tokNumber, "2"},
				{tokNewline, ""},
				{tokText, "dat"},
				{tokNumber, "0"},
				{tokComma, ","},
				{tokText, "test"},
				{tokNewline, ""},
				{tokEOF, ""},
			},
		},
		{
			input: "test test2 equ 2\ndat 0, test\n",
			output: []token{
				{tokText, "test"},
				{tokText, "test2"},
				{tokText, "equ"},
				{tokNumber, "2"},
				{tokNewline, ""},
				{tokText, "dat"},
				{tokNumber, "0"},
				{tokComma, ","},
				{tokText, "test"},
				{tokNewline, ""},
				{tokEOF, ""},
			},
		},
	}
	runForExpanderTests(t, tests)
}

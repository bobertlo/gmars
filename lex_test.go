package gmars

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type lexTestCase struct {
	input    string
	expected []token
	err      bool
}

func runLexTests(t *testing.T, setName string, testCases []lexTestCase) {
	for i, test := range testCases {
		l := newLexer(strings.NewReader(test.input))
		out, err := l.Tokens()
		if test.err {
			require.Error(t, err, fmt.Sprintf("%s test %d", setName, i))
			require.Equal(t, out, test.expected, fmt.Sprintf("%s test %d", setName, i))
		} else {
			require.NoError(t, err, fmt.Errorf("%s test %d: error: %s", setName, i, err))
			assert.Equal(t, test.expected, out, fmt.Sprintf("%s test %d", setName, i))
		}
	}
}

func TestLexer(t *testing.T) {
	testCases := []lexTestCase{
		{
			input: "",
			expected: []token{
				{tokEOF, ""},
			},
		},
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
				{tokSymbol, "#"},
				{tokSymbol, "-"},
				{tokNumber, "1"},
				{tokComma, ","},
				{tokSymbol, "$"},
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
				{tokSymbol, "+"},
				{tokNumber, "3"},
				{tokParenR, ")"},
				{tokSymbol, "-"},
				{tokText, "start"},
				{tokNewline, ""},
				{tokEOF, ""},
			},
		},
		{
			input: "111",
			expected: []token{
				{tokNumber, "111"},
				{tokEOF, ""},
			},
		},
		{
			input: "; comment",
			expected: []token{
				{tokComment, "; comment"},
				{tokEOF, ""},
			},
		},
		{
			input: "text",
			expected: []token{
				{tokText, "text"},
				{tokEOF, ""},
			},
		},
		{
			input: "#",
			expected: []token{
				{tokSymbol, "#"},
				{tokEOF, ""},
			},
		},
		{
			input: "underscore_text",
			expected: []token{
				{tokText, "underscore_text"},
				{tokEOF, ""},
			},
		},
		{
			input: "~",
			expected: []token{
				{tokError, "unexpected character: '~'"},
			},
		},
		{
			input: "label   ;\ndat 0",
			expected: []token{
				{tokText, "label"},
				{tokComment, ";"},
				{tokNewline, ""},
				{tokText, "dat"},
				{tokNumber, "0"},
				{tokEOF, ""},
			},
		},
		{
			input: "for CORESIZE==1\n",
			expected: []token{
				{tokText, "for"},
				{tokText, "CORESIZE"},
				{tokSymbol, "=="},
				{tokNumber, "1"},
				{tokNewline, ""},
				{tokEOF, ""},
			},
		},
		{
			input: "CORESIZE==8000||CORESIZE==800\n",
			expected: []token{
				{tokText, "CORESIZE"},
				{tokSymbol, "=="},
				{tokNumber, "8000"},
				{tokSymbol, "||"},
				{tokText, "CORESIZE"},
				{tokSymbol, "=="},
				{tokNumber, "800"},
				{tokNewline, ""},
				{tokEOF, ""},
			},
		},
		{
			input: "1&&2\n",
			expected: []token{
				{tokNumber, "1"},
				{tokSymbol, "&&"},
				{tokNumber, "2"},
				{tokNewline, ""},
				{tokEOF, ""},
			},
		},
	}

	runLexTests(t, "TestLexer", testCases)
}

func TestLexNegative(t *testing.T) {
	inputs := []string{
		"1 =! 0",
	}

	for _, input := range inputs {
		tokens, err := LexInput(strings.NewReader(input))
		require.NoError(t, err)
		require.Equal(t, tokError, tokens[len(tokens)-1].typ)
	}
}

func TestLexEnd(t *testing.T) {
	l := newLexer(strings.NewReader("test mov 0, 1\n"))

	_, err := l.Tokens()
	assert.NoError(t, err)

	tok, err := l.NextToken()
	assert.Error(t, err)
	assert.Equal(t, token{}, tok)

	tokens, err := l.Tokens()
	assert.Error(t, err)
	assert.Nil(t, tokens)

	r, eof := l.next()
	assert.True(t, eof)
	assert.Equal(t, r, '\x00')
}

func TestBufTokenReader(t *testing.T) {
	in := strings.NewReader("dat 0, 0\n")
	lexer := newLexer(in)
	tokens, err := lexer.Tokens()
	require.NoError(t, err)

	bReader := newBufTokenReader(tokens)
	bTokens, err := bReader.Tokens()
	require.NoError(t, err)

	require.Equal(t, tokens, bTokens)
}

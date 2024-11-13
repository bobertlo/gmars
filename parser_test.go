package gmars

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type parserTestCase struct {
	input  string
	output []sourceLine
	err    bool
}

func assertSourceLineEqual(t *testing.T, expected, value sourceLine) {
	assert.Equal(t, expected.line, value.line, "line")
	assert.Equal(t, expected.codeLine, value.codeLine, "codeline")
	assert.Equal(t, expected.typ, value.typ, "type")
	assert.Equal(t, expected.labels, value.labels)
	assert.Equal(t, expected.amode, value.amode, "amode")

	if expected.a == nil {
		assert.Nil(t, value.a, "a is nil")
	} else {
		require.NotNil(t, value.a, "a not nil")
		assert.Equal(t, expected.a.tokens, value.a.tokens, "a value")
	}

	if expected.b == nil {
		assert.Nil(t, value.b, "b is nil")
	} else {
		require.NotNil(t, value.b, "b not nil")
		assert.Equal(t, expected.b.tokens, value.b.tokens, "b value")
	}

	assert.Equal(t, expected.comment, value.comment, "comment")
	assert.Equal(t, expected.newlines, value.newlines, "newlines")
}

func runParserTests(t *testing.T, setName string, tests []parserTestCase) {
	for i, test := range tests {
		l := newLexer(strings.NewReader(test.input))
		p := newParser(l)

		source, err := p.parse()
		if test.err {
			assert.Error(t, err, fmt.Sprintf("%s test %d", setName, i))
		} else {
			require.NoError(t, err)
			require.Equal(t, len(test.output), len(source.lines))
			for i, line := range source.lines {
				assertSourceLineEqual(t, test.output[i], line)
			}
		}
	}
}

func TestParserPositive(t *testing.T) {
	testCases := []parserTestCase{
		{
			input:  "\n",
			output: []sourceLine{{line: 1, typ: lineEmpty, newlines: 1}},
		},
		{
			input:  "\n\n",
			output: []sourceLine{{line: 1, typ: lineEmpty, newlines: 2}},
		},
		{
			input:  "; comment line\n",
			output: []sourceLine{{line: 1, typ: lineComment, comment: "; comment line", newlines: 1}},
		},
		{
			input:  "end ; comment\n",
			output: []sourceLine{{line: 1, typ: linePseudoOp, op: "end", comment: "; comment", newlines: 1}},
		},
		{
			input: "mov $0, $1 ; comment\n",
			output: []sourceLine{{
				line:     1,
				typ:      lineInstruction,
				op:       "mov",
				amode:    "$",
				a:        &expression{tokens: []token{{typ: tokNumber, val: "0"}}},
				bmode:    "$",
				b:        &expression{tokens: []token{{typ: tokNumber, val: "1"}}},
				comment:  "; comment",
				newlines: 1,
			}},
		},
		{
			input: "a b mov $0, $1 ; comment\n",
			output: []sourceLine{{
				line:     1,
				labels:   []string{"a", "b"},
				typ:      lineInstruction,
				op:       "mov",
				amode:    "$",
				a:        &expression{tokens: []token{{typ: tokNumber, val: "0"}}},
				bmode:    "$",
				b:        &expression{tokens: []token{{typ: tokNumber, val: "1"}}},
				comment:  "; comment",
				newlines: 1,
			}},
		},
		{
			input: "mov $ -1, $ 2 + 2\n",
			output: []sourceLine{{
				line:  1,
				typ:   lineInstruction,
				op:    "mov",
				amode: "$",
				a: &expression{tokens: []token{
					{typ: tokExprOp, val: "-"},
					{typ: tokNumber, val: "1"},
				}},
				bmode: "$",
				b: &expression{tokens: []token{
					{typ: tokNumber, val: "2"},
					{typ: tokExprOp, val: "+"},
					{typ: tokNumber, val: "2"},
				}},
				comment:  "",
				newlines: 1,
			}},
		},
		{
			input: "\n\nmov $0, $1 ; comment\n\nmov $0, $1 ; comment\n",
			output: []sourceLine{
				{line: 1, typ: lineEmpty, newlines: 2},
				{
					line:     3,
					codeLine: 0,
					typ:      lineInstruction,
					op:       "mov",
					amode:    "$",
					a:        &expression{tokens: []token{{typ: tokNumber, val: "0"}}},
					bmode:    "$",
					b:        &expression{tokens: []token{{typ: tokNumber, val: "1"}}},
					comment:  "; comment",
					newlines: 1,
				},
				{line: 4, typ: lineEmpty, newlines: 1},
				{
					line:     5,
					codeLine: 1,
					typ:      lineInstruction,
					op:       "mov",
					amode:    "$",
					a:        &expression{tokens: []token{{typ: tokNumber, val: "0"}}},
					bmode:    "$",
					b:        &expression{tokens: []token{{typ: tokNumber, val: "1"}}},
					comment:  "; comment",
					newlines: 1,
				},
			},
		},
	}

	runParserTests(t, "parser positive", testCases)
}

func TestParserNegative(t *testing.T) {
	testCases := []parserTestCase{
		{
			input: "invalid\n",
			err:   true,
		},
		{
			input: "invalid $1, $2\n",
			err:   true,
		},
	}

	runParserTests(t, "parser negative", testCases)
}

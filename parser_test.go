package gmars

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type parserTestCase struct {
	input  string
	output []sourceLine
	err    bool
}

func runParserTests(t *testing.T, setName string, tests []parserTestCase) {
	for i, test := range tests {
		l := newLexer(strings.NewReader(test.input))
		p := newParser(l)

		source, _, err := p.parse()
		if test.err {
			assert.Error(t, err, fmt.Sprintf("%s test %d: error should be present", setName, i))
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.output, source)
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
				a:        []token{{typ: tokNumber, val: "0"}},
				bmode:    "$",
				b:        []token{{typ: tokNumber, val: "1"}},
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
				a:        []token{{typ: tokNumber, val: "0"}},
				bmode:    "$",
				b:        []token{{typ: tokNumber, val: "1"}},
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
				a: []token{
					{typ: tokExprOp, val: "-"},
					{typ: tokNumber, val: "1"},
				},
				bmode: "$",
				b: []token{
					{typ: tokNumber, val: "2"},
					{typ: tokExprOp, val: "+"},
					{typ: tokNumber, val: "2"},
				},
				comment:  "",
				newlines: 1,
			}},
		},
		{
			input: "mov * -1, * -1\n",
			output: []sourceLine{{
				line:  1,
				typ:   lineInstruction,
				op:    "mov",
				amode: "*",
				a: []token{
					{typ: tokExprOp, val: "-"},
					{typ: tokNumber, val: "1"},
				},
				bmode: "*",
				b: []token{
					{typ: tokExprOp, val: "-"},
					{typ: tokNumber, val: "1"},
				},
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
					a:        []token{{typ: tokNumber, val: "0"}},
					bmode:    "$",
					b:        []token{{typ: tokNumber, val: "1"}},
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
					a:        []token{{typ: tokNumber, val: "0"}},
					bmode:    "$",
					b:        []token{{typ: tokNumber, val: "1"}},
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
		{
			input: "mov $undefined, $2\n",
			err:   true,
		},
		{
			input: "mov $1, $undefined\n",
			err:   true,
		},
		{
			input: "redefined mov $0, $1\nredefined mov $0, $1\n",
			err:   true,
		},
		{
			input: "mov\n",
			err:   true,
		},
		{
			input: "mov ;comment\n",
			err:   true,
		},
		{
			input: "mov",
			err:   true,
		},
		{
			input: "mov $1,\n",
			err:   true,
		},
		{
			input: "mov,;comment\n",
			err:   true,
		},
	}

	runParserTests(t, "parser negative", testCases)
}

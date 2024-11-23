package gmars

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type symbolScannerTestCase struct {
	input  string
	output map[string][]token
}

func runSymbolScannerTests(t *testing.T, cases []symbolScannerTestCase) {
	for _, test := range cases {
		tokens, err := LexInput(strings.NewReader(test.input))
		require.NoError(t, err)
		require.NotNil(t, tokens)

		scanner := newSymbolScanner(newBufTokenReader(tokens))
		symbols, err := scanner.ScanInput()
		require.NoError(t, err)
		require.NotNil(t, symbols)

		require.Equal(t, test.output, symbols)
	}
}

func TestSymbolScanner(t *testing.T) {
	tests := []symbolScannerTestCase{
		{
			input: "test equ 2\ndat 0, test\n",
			output: map[string][]token{
				"test": {{tokNumber, "2"}},
			},
		},
		{
			input:  "dat 0, 0",
			output: map[string][]token{},
		},
		{
			input: "test\ntest2\nequ 2",
			output: map[string][]token{
				"test":  {{tokNumber, "2"}},
				"test2": {{tokNumber, "2"}},
			},
		},
		{
			// ignore symbols inside for loops because they could be redifined.
			// will just re-scan after expanding for loops
			input: "test equ 2\nfor 0\nq equ 1\nrof\nfor 1\nq equ 2\nrof\n",
			output: map[string][]token{
				"test": {{tokNumber, "2"}},
			},
		},
		{
			input:  "for 1\nend\nrof\n ~",
			output: map[string][]token{},
		},
	}
	runSymbolScannerTests(t, tests)
}

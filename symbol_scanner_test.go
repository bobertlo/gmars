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
	}
	runSymbolScannerTests(t, tests)
}

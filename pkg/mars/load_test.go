package mars

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const imp88 = `;redcode
;name Imp
;author A K Dewdney
;strategy this is the simplest program
;strategy it was described in the initial articles

		MOV # 0, $ 1
		END 0
`

func TestLoadImp(t *testing.T) {
	config := StandardConfig()

	reader := strings.NewReader(imp88)
	data, err := ParseLoadFile(reader, config)
	require.NoError(t, err)
	require.Equal(t, "Imp", data.Name)
	require.Equal(t, "A K Dewdney", data.Author)
	require.Equal(t, "this is the simplest program\nit was described in the initial articles\n", data.Strategy)
}

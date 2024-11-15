package gmars

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsCycleNegative(t *testing.T) {
	refs := map[string][]string{
		"a": {"b"},
		"b": {"c", "d"},
		"c": {"d"},
	}

	cyclic, key := graphContainsCycle(refs)
	assert.False(t, cyclic)
	assert.Equal(t, "", key)
}

func TestContainsCyclePositive(t *testing.T) {
	refs := map[string][]string{
		"a": {"b"},
		"b": {"c", "d"},
		"c": {"b"},
	}
	cyclic, key := graphContainsCycle(refs)
	assert.True(t, cyclic)
	assert.Equal(t, "b", key)
}

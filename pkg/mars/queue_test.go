package mars

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueueEmpty(t *testing.T) {
	pq := newProcessQueue(2)
	_, err := pq.Pop()
	require.Error(t, err)
}

func TestQueue(t *testing.T) {
	pq := newProcessQueue(2)

	pq.Push(1)
	pq.Push(2)
	pq.Push(3)

	out, err := pq.Pop()
	require.NoError(t, err)
	require.Equal(t, 1, int(out))

	out, err = pq.Pop()
	require.NoError(t, err)
	require.Equal(t, 2, int(out))

	_, err = pq.Pop()
	require.Error(t, err)
}

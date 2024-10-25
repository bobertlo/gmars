package mars

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func makeSim94() *Simulator {
	return NewSimulator(8000, 8000, 80000, 8000, 8000, false)
}

func makeSim88() *Simulator {
	return NewSimulator(8000, 8000, 80000, 8000, 8000, true)
}

func TestDwarf(t *testing.T) {
	wdata := makeDwarfData()

	sim := makeSim88()
	sim.reporters = append(sim.reporters, NewDebugReporter(sim))
	w, err := sim.SpawnWarrior(wdata, 0)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 4; i++ {
		sim.run_turn()
		require.True(t, w.Alive())
		require.Equal(t, w.pq.Len(), Address(1))
	}

	require.Equal(t, Instruction{
		Op:     DAT,
		OpMode: F,
		AMode:  IMMEDIATE,
		A:      0,
		BMode:  IMMEDIATE,
		B:      8,
	},
		sim.mem[3])

	require.Equal(t, Instruction{
		Op:     DAT,
		OpMode: F,
		AMode:  IMMEDIATE,
		A:      0,
		BMode:  IMMEDIATE,
		B:      4,
	},
		sim.mem[7])

	for i := 0; i < 4; i++ {
		sim.run_turn()
		require.True(t, w.Alive())
		require.Equal(t, w.pq.Len(), Address(1))
	}

	n, _ := w.pq.Pop()
	require.Equal(t, 2, int(n))
}

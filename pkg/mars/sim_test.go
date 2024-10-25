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

func TestDat(t *testing.T) {
	wdata := &WarriorData{
		Code: []Instruction{
			{
				Op:     DAT,
				OpMode: F,
			},
		},
	}

	sim := makeSim88()
	w, err := sim.SpawnWarrior(wdata, 0)
	if err != nil {
		t.Fatal(err)
	}

	sim.run_turn()
	require.True(t, w.Alive())
	require.Equal(t, w.pq.Len(), Address(0))

	sim.run_turn()
	require.False(t, w.Alive())
}

func TestDatDecA(t *testing.T) {
	wdata := &WarriorData{
		Code: []Instruction{
			{
				Op:     DAT,
				OpMode: F,
				AMode:  B_DECREMENT,
				A:      1,
				BMode:  IMMEDIATE,
				B:      0,
			},
		},
	}

	sim := makeSim88()
	w, err := sim.SpawnWarrior(wdata, 0)
	if err != nil {
		t.Fatal(err)
	}

	sim.run_turn()
	require.True(t, w.Alive())
	require.Equal(t, w.pq.Len(), Address(0))

	require.Equal(t, Instruction{
		B: 8000 - 1,
	}, sim.mem[1])

	sim.run_turn()
	require.False(t, w.Alive())
}

func TestDatDecB(t *testing.T) {
	wdata := &WarriorData{
		Code: []Instruction{
			{
				Op:     DAT,
				OpMode: F,
				AMode:  IMMEDIATE,
				A:      0,
				BMode:  B_DECREMENT,
				B:      1,
			},
		},
	}

	sim := makeSim88()
	w, err := sim.SpawnWarrior(wdata, 0)
	if err != nil {
		t.Fatal(err)
	}

	sim.run_turn()
	require.True(t, w.Alive())
	require.Equal(t, w.pq.Len(), Address(0))

	require.Equal(t, Instruction{
		B: 8000 - 1,
	}, sim.mem[1])

	sim.run_turn()
	require.False(t, w.Alive())
}

func TestDwarf(t *testing.T) {
	wdata := makeDwarfData()

	sim := makeSim88()
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

package mars

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMovI(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     MOV,
				OpMode: I,
				AMode:  DIRECT,
				A:      1,
				BMode:  DIRECT,
				B:      2,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, data.Code[1], sim.mem[1])
	require.Equal(t, data.Code[1], sim.mem[2])
}

func TestMovAB_Direct(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     MOV,
				OpMode: AB,
				AMode:  IMMEDIATE,
				A:      1,
				BMode:  DIRECT,
				B:      1,
			},
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     3,
				BMode: IMMEDIATE,
				B:     4,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, sim.mem[1], Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     3,
		BMode: IMMEDIATE,
		B:     1,
	})
}

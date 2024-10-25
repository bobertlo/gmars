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

func TestAddA(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     ADD,
				OpMode: A,
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
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     8,
		BMode: IMMEDIATE,
		B:     6,
	}, sim.mem[2])
}

func TestAddB(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     ADD,
				OpMode: B,
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
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     5,
		BMode: IMMEDIATE,
		B:     10,
	}, sim.mem[2])
}

func TestAddAB(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     ADD,
				OpMode: AB,
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
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     5,
		BMode: IMMEDIATE,
		B:     9,
	}, sim.mem[2])
}

func TestAddBA(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     ADD,
				OpMode: BA,
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
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     9,
		BMode: IMMEDIATE,
		B:     6,
	}, sim.mem[2])
}

func TestAddI(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     ADD,
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
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     8,
		BMode: IMMEDIATE,
		B:     10,
	}, sim.mem[2])
}

func TestAddF(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     ADD,
				OpMode: F,
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
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     8,
		BMode: IMMEDIATE,
		B:     10,
	}, sim.mem[2])
}

func TestAddX(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     ADD,
				OpMode: X,
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
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     5,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     9,
		BMode: IMMEDIATE,
		B:     8,
	}, sim.mem[2])
}

func TestSubA(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     SUB,
				OpMode: A,
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
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     2,
		BMode: IMMEDIATE,
		B:     6,
	}, sim.mem[2])
}

func TestSubB(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     SUB,
				OpMode: B,
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
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     5,
		BMode: IMMEDIATE,
		B:     2,
	}, sim.mem[2])
}

func TestSubAB(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     SUB,
				OpMode: AB,
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
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     5,
		BMode: IMMEDIATE,
		B:     3,
	}, sim.mem[2])
}

func TestSubBA(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     SUB,
				OpMode: BA,
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
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     1,
		BMode: IMMEDIATE,
		B:     6,
	}, sim.mem[2])
}

func TestSubI(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     SUB,
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
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     2,
		BMode: IMMEDIATE,
		B:     2,
	}, sim.mem[2])
}

func TestSubF(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     SUB,
				OpMode: F,
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
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     2,
		BMode: IMMEDIATE,
		B:     2,
	}, sim.mem[2])
}

func TestSubX(t *testing.T) {
	sim := makeSim88()

	data := &WarriorData{
		Name:   "test",
		Author: "test",
		Code: []Instruction{
			{
				Op:     SUB,
				OpMode: X,
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
			{
				Op:    DAT,
				AMode: IMMEDIATE,
				A:     5,
				BMode: IMMEDIATE,
				B:     6,
			},
		},
	}
	w, _ := sim.SpawnWarrior(data, 0)
	sim.run_turn()

	require.True(t, w.Alive())
	n, _ := w.pq.Pop()
	require.Equal(t, n, Address(1))

	require.Equal(t, Instruction{
		Op:    DAT,
		AMode: IMMEDIATE,
		A:     1,
		BMode: IMMEDIATE,
		B:     3,
	}, sim.mem[2])
}

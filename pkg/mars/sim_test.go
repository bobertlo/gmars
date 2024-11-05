package mars

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImp(t *testing.T) {
	tests := []redcodeTest{
		{
			input: []string{
				"mov.i #0, $1",
			},
			output: []string{
				"mov.i #0, $1", "mov.i #0, $1", "mov.i #0, $1", "mov.i #0, $1",
				"mov.i #0, $1", "mov.i #0, $1", "mov.i #0, $1", "mov.i #0, $1",
				"mov.i #0, $1", "mov.i #0, $1", "mov.i #0, $1", "mov.i #0, $1",
				"mov.i #0, $1", "mov.i #0, $1", "mov.i #0, $1", "mov.i #0, $1",
			},
			turns: 16,
			pq:    []Address{0},
		},
	}
	runTests(t, "dwarf", tests)
}

func TestDwarf(t *testing.T) {
	tests := []redcodeTest{
		{
			input: []string{
				"add.ab #4, $3", "mov.i  $2, @2", "jmp.b $-2, $0", "dat.f  #0, #0",
			},
			output: []string{
				"add.ab #4, $3", "mov.i  $2, @2", "jmp.b $-2, $0", "dat.f  #0, #12",
				"dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f #0, #4",
				"dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f #0, #8",
				"dat.f $0, $0", "dat.f $0, $0", "dat.f $0, $0", "dat.f #0, #12",
			},
			turns: 9,
			pq:    []Address{0},
		},
	}
	runTests(t, "dwarf", tests)
}

func TestRunImp(t *testing.T) {
	reader := strings.NewReader(imp88)
	config := ConfigKOTH88()
	impdata, err := ParseLoadFile(reader, config)
	require.NoError(t, err)

	sim, err := NewSimulator(config)
	require.NoError(t, err)
	w, err := sim.AddWarrior(&impdata)
	require.NoError(t, err)
	err = sim.SpawnWarrior(0, 0)
	require.NoError(t, err)

	state := sim.Run()
	require.Equal(t, 1, len(state))
	require.True(t, state[0])
	require.True(t, w.Alive())
	require.Equal(t, 80000, sim.CycleCount())
}

func TestRunTwoImps(t *testing.T) {
	reader := strings.NewReader(imp88)
	config := ConfigKOTH88()
	impdata, err := ParseLoadFile(reader, config)
	require.NoError(t, err)

	sim, err := NewSimulator(config)
	require.NoError(t, err)
	w, err := sim.AddWarrior(&impdata)
	require.NoError(t, err)
	err = sim.SpawnWarrior(0, 0)
	require.NoError(t, err)
	w2, err := sim.AddWarrior(&impdata)
	require.NoError(t, err)
	err = sim.SpawnWarrior(1, 200)
	require.NoError(t, err)

	state := sim.Run()
	require.Equal(t, 2, len(state))
	require.True(t, state[0])
	require.True(t, state[1])
	require.True(t, w.Alive())
	require.True(t, w2.Alive())
	require.Equal(t, 80000, sim.CycleCount())
}

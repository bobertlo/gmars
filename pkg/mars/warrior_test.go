package mars

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func makeDwarfData() *WarriorData {
	return &WarriorData{
		Name:   "Dwarf",
		Author: "A K Dewdney",
		Code: []Instruction{
			{
				Op:     ADD,
				OpMode: AB,
				AMode:  IMMEDIATE,
				A:      4,
				BMode:  DIRECT,
				B:      3,
			},
			{
				Op:     MOV,
				OpMode: I,
				AMode:  DIRECT,
				A:      2,
				BMode:  B_INDIRECT,
				B:      2,
			},
			{
				Op:     JMP,
				OpMode: B,
				AMode:  DIRECT,
				A:      8000 - 2,
				BMode:  DIRECT,
				B:      0,
			},
			{
				Op:     DAT,
				OpMode: F,
				AMode:  IMMEDIATE,
				A:      0,
				BMode:  IMMEDIATE,
				B:      0,
			},
		}}
}

func TestWarriorMethodLoadCode88(t *testing.T) {
	wdata := makeDwarfData()
	sim := makeSim88()
	w, err := sim.SpawnWarrior(wdata, 0)
	if err != nil {
		t.Fatal(err)
	}

	loadCodeStr := w.LoadCodePMARS()
	dat, err := os.ReadFile("test/dwarf_pmars88.rc")
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, string(dat), loadCodeStr)
}

func TestWarriorMethodLoadCode94(t *testing.T) {
	wdata := makeDwarfData()
	sim := makeSim94()
	w, err := sim.SpawnWarrior(wdata, 0)
	if err != nil {
		t.Fatal(err)
	}

	loadCodeStr := w.LoadCodePMARS()
	dat, err := os.ReadFile("test/dwarf_pmars94.rc")
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, string(dat), loadCodeStr)
}

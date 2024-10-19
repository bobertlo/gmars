package main

import (
	"fmt"
	"os"

	"github.com/bobertlo/gmars/pkg/mars"
)

func print_from_memory() {
	sim := mars.NewMARS(8000, 8000, 80000, 8000, 8000, true)

	code := []mars.Instruction{
		{
			Op:     mars.ADD,
			OpMode: mars.AB,
			AMode:  mars.IMMEDIATE,
			A:      4,
			BMode:  mars.DIRECT,
			B:      3,
		},
		{
			Op:     mars.MOV,
			OpMode: mars.I,
			AMode:  mars.DIRECT,
			A:      2,
			BMode:  mars.B_INDIRECT,
			B:      2,
		},
		{
			Op:     mars.JMP,
			OpMode: mars.B,
			AMode:  mars.DIRECT,
			A:      8000 - 2,
			BMode:  mars.DIRECT,
			B:      0,
		},
		{
			Op:    mars.DAT,
			AMode: mars.IMMEDIATE,
			BMode: mars.IMMEDIATE,
		},
	}

	wdata := &mars.WarriorData{
		Name:   "Dwarf",
		Author: "A K Dewdney",
		Code:   code,
		Start:  0,
	}

	warrior, err := sim.AddWarrior(wdata, 0)
	if err != nil {
		panic(err)
	}

	fmt.Printf(`Program "%s" (length %d) by "%s"`+"\n\n", warrior.Name(), warrior.Length(), warrior.Author())
	fmt.Println(warrior.LoadCode())
}

func load_and_print(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	sim := mars.NewMARS(8000, 8000, 80000, 8000, 8000, true)
	warrior, err := sim.LoadWarrior(f)
	if err != nil {
		return err
	}

	fmt.Println(warrior.LoadCodePMARS())

	return nil
}

func main() {
	// print_from_memory()

	err := load_and_print("warriors/dwarf_88.rc")
	if err != nil {
		fmt.Println(err)
	}
}

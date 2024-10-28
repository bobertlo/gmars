package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/bobertlo/gmars/pkg/mars"
)

func main() {
	// use88 := flag.Bool("-8", false, "Enforce ICWS'88 rules")
	sizeFlag := flag.Int("s", 8000, "Size of core")
	procFlag := flag.Int("p", 8000, "Max. Processes")
	cycleFlag := flag.Int("c", 80000, "Cycles until tie")
	lenFlag := flag.Int("l", 100, "Max. warrior length")
	fixedFlag := flag.Int("F", 0, "fixed position of warrior #2")
	flag.Parse()

	coresize := mars.Address(*sizeFlag)
	processes := mars.Address(*procFlag)
	cycles := mars.Address(*cycleFlag)
	length := mars.Address(*lenFlag)

	config := mars.BasicConfig(mars.ICWS88, coresize, processes, cycles, length)
	sim := mars.NewSimulator(config)

	args := flag.Args()

	if len(args) < 2 || len(args) > 2 {
		fmt.Println("only 2 warrior battles supported")
		os.Exit(1)
	}

	w1file, err := os.Open(args[0])
	if err != nil {
		fmt.Printf("error opening warrior file '%s': %s\n", args[0], err)
		os.Exit(1)
	}
	defer w1file.Close()
	w1data, err := mars.ParseLoadFile(w1file, config)
	if err != nil {
		fmt.Printf("error parsing warrior file '%s': %s\n", args[0], err)
		os.Exit(1)
	}
	w1file.Close()

	w2file, err := os.Open(args[1])
	if err != nil {
		fmt.Printf("error opening warrior file '%s': %s\n", args[1], err)
		os.Exit(1)
	}
	defer w1file.Close()
	w2data, err := mars.ParseLoadFile(w2file, config)
	if err != nil {
		fmt.Printf("error parsing warrior file '%s': %s\n", args[1], err)
	}
	w1file.Close()

	w2start := *fixedFlag

	if w2start == 0 {
		minStart := 2 * config.Length
		maxStart := config.CoreSize - config.Length - 1
		startRange := maxStart - minStart
		w2start = rand.Intn(int(startRange)+1) + int(minStart)
	}

	w1, err := sim.SpawnWarrior(&w1data, 0)
	if err != nil {
		fmt.Printf("error spawning warrior 1: %n", err)
	}

	w2, err := sim.SpawnWarrior(&w2data, mars.Address(w2start))
	if err != nil {
		fmt.Printf("error spawning warrior 2: %n", err)
	}

	sim.Run()

	w1win := 0
	w1tie := 0
	if w1.Alive() {
		if w2.Alive() {
			w1tie = 1
		} else {
			w1win = 1
		}
	}
	w2win := 0
	w2tie := 0
	if w2.Alive() {
		if w1.Alive() {
			w2tie = 1
		} else {
			w2win = 1
		}
	}

	fmt.Printf("%d %d\n", w1win, w1tie)
	fmt.Printf("%d %d\n", w2win, w2tie)
}

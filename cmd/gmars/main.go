package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/bobertlo/gmars"
)

func main() {
	use88Flag := flag.Bool("8", false, "Enforce ICWS'88 rules")
	sizeFlag := flag.Int("s", 8000, "Size of core")
	procFlag := flag.Int("p", 8000, "Max. Processes")
	cycleFlag := flag.Int("c", 80000, "Cycles until tie")
	lenFlag := flag.Int("l", 100, "Max. warrior length")
	fixedFlag := flag.Int("F", 0, "fixed position of warrior #2")
	roundFlag := flag.Int("r", 1, "Rounds to play")
	debugFlag := flag.Bool("debug", false, "Dump verbose reporting of simulator state")
	assembleFlag := flag.Bool("A", false, "Assemble and output warriors only")
	presetFlag := flag.String("preset", "", "Load named preset config (and ignore other flags)")
	flag.Parse()

	var config gmars.SimulatorConfig
	if *presetFlag != "" {
		presetConfig, err := gmars.PresetConfig(*presetFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading config: %s\n", err)
			os.Exit(1)
		}
		config = presetConfig
	} else {
		var mode gmars.SimulatorMode
		if *use88Flag {
			mode = gmars.ICWS88
		} else {
			mode = gmars.ICWS94
		}
		coresize := gmars.Address(*sizeFlag)
		processes := gmars.Address(*procFlag)
		cycles := gmars.Address(*cycleFlag)
		length := gmars.Address(*lenFlag)
		config = gmars.NewQuickConfig(mode, coresize, processes, cycles, length)
	}

	args := flag.Args()

	if *assembleFlag {
		if len(args) != 1 {
			fmt.Println("wrong number of arguments")
			os.Exit(1)
		}
	} else if len(args) < 2 || len(args) > 2 {
		fmt.Println("only 2 warrior battles supported")
		os.Exit(1)
	}

	w1file, err := os.Open(args[0])
	if err != nil {
		fmt.Printf("error opening warrior file '%s': %s\n", args[0], err)
		os.Exit(1)
	}
	defer w1file.Close()
	w1data, err := gmars.CompileWarrior(w1file, config)
	if err != nil {
		fmt.Printf("error parsing warrior file '%s': %s\n", args[0], err)
		os.Exit(1)
	}
	w1file.Close()

	if *assembleFlag {
		sim, err := gmars.NewSimulator(config)
		if err != nil {
			fmt.Printf("error creating sim: %s", err)
		}
		w1, err := sim.AddWarrior(&w1data)
		if err != nil {
			fmt.Printf("error loading warrior: %s", err)
		}
		fmt.Println(w1.LoadCode())
		return
	}

	w2file, err := os.Open(args[1])
	if err != nil {
		fmt.Printf("error opening warrior file '%s': %s\n", args[1], err)
		os.Exit(1)
	}
	defer w1file.Close()
	w2data, err := gmars.CompileWarrior(w2file, config)
	if err != nil {
		fmt.Printf("error parsing warrior file '%s': %s\n", args[1], err)
	}
	w1file.Close()

	rounds := *roundFlag

	w1win := 0
	w1tie := 0
	w2win := 0
	w2tie := 0
	for i := 0; i < rounds; i++ {
		sim, err := gmars.NewReportingSimulator(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating sim: %s", err)
		}
		if *debugFlag {
			sim.AddReporter(gmars.NewDebugReporter(sim))
		}

		w2start := *fixedFlag
		if w2start == 0 {
			minStart := 2 * config.Length
			maxStart := config.CoreSize - config.Length - 1
			startRange := maxStart - minStart
			w2start = rand.Intn(int(startRange)+1) + int(minStart)
		}

		w1, err := sim.AddWarrior(&w1data)
		if err != nil {
			fmt.Printf("error adding warrior 1: %s", err)
		}
		err = sim.SpawnWarrior(0, 0)
		if err != nil {
			fmt.Printf("error adding warrior 1: %s", err)
		}

		w2, err := sim.AddWarrior(&w2data)
		if err != nil {
			fmt.Printf("error adding warrior 2: %s", err)
		}
		err = sim.SpawnWarrior(1, gmars.Address(w2start))
		if err != nil {
			fmt.Printf("error spawning warrior 1: %s", err)
		}

		sim.Run()

		if w1.Alive() {
			if w2.Alive() {
				w1tie += 1
			} else {
				w1win += 1
			}
		}

		if w2.Alive() {
			if w1.Alive() {
				w2tie += 1
			} else {
				w2win += 1
			}
		}
	}
	fmt.Printf("%d %d\n", w1win, w1tie)
	fmt.Printf("%d %d\n", w2win, w2tie)
}

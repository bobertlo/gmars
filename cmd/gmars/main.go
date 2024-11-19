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

	if len(args) > 2 {
		fmt.Fprintf(os.Stderr, "only 2 warrior battles supported")
		os.Exit(1)
	}

	warriors := make([]gmars.WarriorData, 0)
	for _, arg := range args {
		in, err := os.Open(arg)
		if err != nil {
			fmt.Printf("error opening warrior file '%s': %s\n", arg, err)
			os.Exit(1)
		}
		defer in.Close()

		warrior, err := gmars.CompileWarrior(in, config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing warrior file '%s': %s\n", arg, err)
			os.Exit(1)
		}

		warriors = append(warriors, warrior)
	}

	if len(warriors) == 0 {
		fmt.Fprintf(os.Stderr, "no warriors specified\n")
		os.Exit(1)
	}
	if *assembleFlag {
		sim, err := gmars.NewSimulator(config)
		if err != nil {
			fmt.Printf("error creating sim: %s", err)
		}

		for _, warriorData := range warriors {
			w, err := sim.AddWarrior(&warriorData)
			if err != nil {
				fmt.Printf("error loading warrior: %s", err)
			}
			fmt.Println(w.LoadCode())
		}

		return
	}

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

		w1, err := sim.AddWarrior(&warriors[0])
		if err != nil {
			fmt.Printf("error adding warrior 1: %s", err)
			os.Exit(1)
		}
		err = sim.SpawnWarrior(0, 0)
		if err != nil {
			fmt.Printf("error adding warrior 1: %s", err)
			os.Exit(1)
		}

		var w2 gmars.Warrior
		if len(warriors) > 1 {
			w, err := sim.AddWarrior(&warriors[1])
			if err != nil {
				fmt.Printf("error adding warrior 2: %s", err)
				os.Exit(1)
			}
			err = sim.SpawnWarrior(1, gmars.Address(w2start))
			if err != nil {
				fmt.Printf("error spawning warrior 1: %s", err)
				os.Exit(1)
			}
			w2 = w
		}

		sim.Run()

		if len(warriors) == 1 {
			if w1.Alive() {
				w1win += 1
			}
		} else if len(warriors) == 2 {
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
	}
	fmt.Printf("%d %d\n", w1win, w1tie)
	if len(warriors) > 1 {
		fmt.Printf("%d %d\n", w2win, w2tie)
	}
}

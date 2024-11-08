//go:generate fyne bundle -o data.go Icon.png

// Package main launches the life app
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/bobertlo/gmars/pkg/mars"
)

const (
	minXCount = 50
	minYCount = 40
)

func show(sim mars.ReportingSimulator, app fyne.App) {
	game := newGame(sim)

	window := app.NewWindow("Corewar")

	window.SetContent(game.buildUI())
	window.Canvas().SetOnTypedRune(game.typedRune)

	// start the board animation before we show the window - it will block
	game.animate()

	window.Show()
}

func main() {
	use88Flag := flag.Bool("8", false, "Enforce ICWS'88 rules")
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

	var mode mars.SimulatorMode

	if *use88Flag {
		mode = mars.ICWS88
	} else {
		mode = mars.ICWS94
	}
	config := mars.NewQuickConfig(mode, coresize, processes, cycles, length)

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

	sim, err := mars.NewReportingSimulator(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating sim: %s", err)
	}

	w2start := *fixedFlag
	if w2start == 0 {
		minStart := 2 * config.Length
		maxStart := config.CoreSize - config.Length - 1
		startRange := maxStart - minStart
		w2start = rand.Intn(int(startRange)+1) + int(minStart)
	}

	_, err = sim.AddWarrior(&w1data)
	if err != nil {
		fmt.Printf("error adding warrior 1: %s", err)
	}
	err = sim.SpawnWarrior(0, 0)
	if err != nil {
		fmt.Printf("error adding warrior 1: %s", err)
	}

	_, err = sim.AddWarrior(&w2data)
	if err != nil {
		fmt.Printf("error adding warrior 2: %s", err)
	}
	err = sim.SpawnWarrior(1, mars.Address(w2start))
	if err != nil {
		fmt.Printf("error spawning warrior 1: %s", err)
	}

	app := app.New()
	app.Settings().SetTheme(&gameTheme{})

	show(sim, app)
	app.Run()
}

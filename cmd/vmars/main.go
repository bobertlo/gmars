package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"image"
	"log"
	"math/rand"
	"os"

	"github.com/bobertlo/gmars"
	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

const (
	tileSize         = 6
	defaultSpeedStep = 6
)

var (
	mplusFaceSource *text.GoTextFaceSource

	//go:embed assets/tiles_6.png
	tiles_png []byte

	tilesImage *ebiten.Image

	speeds = []int{-64, -32, -16, -8, -4, -2, 1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192}
)

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}
	mplusFaceSource = s

	img, _, err := image.Decode(bytes.NewReader(tiles_png))
	if err != nil {
		log.Fatal(err)
	}
	tilesImage = ebiten.NewImageFromImage(img)

}

func main() {
	use88Flag := flag.Bool("8", false, "Enforce ICWS'88 rules")
	sizeFlag := flag.Int("s", 8000, "Size of core")
	procFlag := flag.Int("p", 8000, "Max. Processes")
	cycleFlag := flag.Int("c", 80000, "Cycles until tie")
	lenFlag := flag.Int("l", 100, "Max. warrior length")
	fixedFlag := flag.Int("F", 0, "fixed position of warrior #2")
	// roundFlag := flag.Int("r", 1, "Rounds to play")
	showReadFlag := flag.Bool("showread", false, "display reads in the visualizer")
	debugFlag := flag.Bool("debug", false, "Dump verbose reporting of simulator state")
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

	sim, err := gmars.NewReportingSimulator(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating sim: %s", err)
	}
	if *debugFlag {
		sim.AddReporter(gmars.NewDebugReporter(sim))
	}
	rec := gmars.NewStateRecorder(sim)
	rec.SetRecordRead(*showReadFlag)
	sim.AddReporter(rec)

	sim.AddWarrior(&warriors[0])
	if len(warriors) > 1 {
		sim.AddWarrior(&warriors[1])
	}

	sim.SpawnWarrior(0, 0)

	if len(warriors) > 1 {
		w2start := *fixedFlag
		if w2start == 0 {
			minStart := 2 * config.Length
			maxStart := config.CoreSize - config.Length - 1
			startRange := maxStart - minStart
			w2start = rand.Intn(int(startRange)+1) + int(minStart)
		}
		sim.SpawnWarrior(1, gmars.Address(w2start))
	}

	game := &Game{
		sim:        sim,
		config:     config,
		rec:        *rec,
		speedStep:  defaultSpeedStep,
		fixedStart: *fixedFlag,
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	if len(warriors) > 1 {
		ebiten.SetWindowTitle(fmt.Sprintf("gMARS - '%s' vs '%s'", warriors[0].Name, warriors[1].Name))
	} else {
		ebiten.SetWindowTitle(fmt.Sprintf("gMARS - '%s'", warriors[0].Name))
	}
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

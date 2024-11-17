package main

import (
	"errors"
	"image/color"

	"github.com/bobertlo/gmars"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func NewGame(config gmars.SimulatorConfig, sim gmars.ReportingSimulator, rec *gmars.StateRecorder, defaultSpeedStep int) *Game {
	game := &Game{
		config:    config,
		sim:       sim,
		rec:       *rec,
		speedStep: defaultSpeedStep,
		hues:      []float64{0.0, 250.0},
	}
	game.cacheSprites()
	game.cacheColors()
	return game
}

type Game struct {
	config      gmars.SimulatorConfig
	sim         gmars.ReportingSimulator
	rec         gmars.StateRecorder
	running     bool
	finished    bool
	speedStep   int
	counter     int
	spriteCache [][]*ebiten.Image
	colorCache  []color.Color
	hues        []float64
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) slowDown() {
	g.speedStep--
	if g.speedStep < 0 {
		g.speedStep = 0
	}
}

func (g *Game) speedUp() {
	g.speedStep++
	if g.speedStep >= len(speeds) {
		g.speedStep = len(speeds) - 1
	}
}

func (g *Game) handleInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.running = !g.running
	} else if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		w2Start := g.config.GetW2Start()
		g.sim.Reset()
		g.sim.SpawnWarrior(0, 0)
		g.sim.SpawnWarrior(1, w2Start)
		g.finished = false
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.slowDown()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.speedUp()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		g.running = false
	} else if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		if g.running {
			g.running = false
		} else {
			for i := 0; i < speeds[g.speedStep]; i++ {
				g.runCycle()
			}
		}
	}
}

func (g *Game) runCycle() {
	if g.finished {
		return
	}

	if g.sim.WarriorLivingCount() > 1 && g.sim.CycleCount() < g.sim.MaxCycles() {
		g.sim.RunCycle()
	} else {
		g.finished = true
	}
}

func (g *Game) Update() error {
	speed := speeds[g.speedStep]

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return errors.New("game ended by player")
	}

	if g.running {
		if speed < 0 {
			if g.counter%speed == 0 {
				g.runCycle()
			}
		} else {
			for i := 0; i < speeds[g.speedStep]; i++ {
				g.runCycle()
			}
		}
	}

	g.handleInput()

	g.counter++

	return nil
}

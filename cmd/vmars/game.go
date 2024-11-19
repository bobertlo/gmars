package main

import (
	"errors"
	"math/rand"

	"github.com/bobertlo/gmars"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	sim        gmars.ReportingSimulator
	config     gmars.SimulatorConfig
	fixedStart int
	rec        gmars.StateRecorder
	running    bool
	finished   bool
	speedStep  int
	counter    int
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
		g.sim.Reset()
		g.sim.SpawnWarrior(0, 0)

		if g.sim.WarriorCount() > 1 {
			w2start := g.fixedStart
			if w2start == 0 {
				minStart := 2 * g.config.Length
				maxStart := g.config.CoreSize - g.config.Length - 1
				startRange := maxStart - minStart
				w2start = rand.Intn(int(startRange)+1) + int(minStart)
			}
			g.sim.SpawnWarrior(1, gmars.Address(w2start))
		}

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

	count := g.sim.WarriorCount()
	living := g.sim.WarriorLivingCount()
	if ((count > 1 && living > 1) || living > 0) && g.sim.CycleCount() < g.sim.MaxCycles() {
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

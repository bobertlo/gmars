package main

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"image"
	"image/color"
	"log"
	"math/rand"
	"os"

	"github.com/bobertlo/gmars/pkg/mars"
	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

const (
	tileSize         = 6
	defaultSpeedStep = 8
)

type Game struct {
	sim       mars.ReportingSimulator
	rec       mars.StateRecorder
	running   bool
	speedStep int
	counter   int
}

var (
	mplusFaceSource *text.GoTextFaceSource

	//go:embed assets/tiles_6.png
	tiles_png []byte

	tilesImage *ebiten.Image

	speeds = []int{-60, -30, -15, -4, -2, 1, 2, 4, 16, 32, 64, 128, 256, 512, 1024}
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
		g.sim.SpawnWarrior(1, mars.Address(rand.Intn(7000)+200))
	} else if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		g.slowDown()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		g.speedUp()
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
				g.sim.RunCycle()
			}
		} else {
			for i := 0; i < speeds[g.speedStep]; i++ {
				g.sim.RunCycle()
			}
		}
	}

	g.handleInput()

	g.counter++

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	scales := make([]ebiten.ColorScale, 3)
	scales[0].Scale(1, 1, 1, 1)
	scales[1].Scale(1, 1, 0, 1)
	scales[2].Scale(0, 1, 1, 1)

	w := tilesImage.Bounds().Dx()
	tileXCount := w / tileSize

	const xCount = screenWidth / tileSize

	for i := 0; i < int(g.sim.CoreSize()); i++ {
		state, color := g.rec.GetMemState(mars.Address(i))

		// fmt.Println(i, state)
		if state == mars.CoreEmpty {
			continue
		}
		t := int(state)

		op := &ebiten.DrawImageOptions{ColorScale: scales[color+1]}
		op.GeoM.Translate(float64((i%xCount)*tileSize), float64((i/xCount)*tileSize))

		sx := (t % tileXCount) * tileSize
		sy := (t / tileXCount) * tileSize
		screen.DrawImage(tilesImage.SubImage(image.Rect(sx, sy, sx+tileSize, sy+tileSize)).(*ebiten.Image), op)
	}

	// Draw info
	msg := fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS())
	op := &text.DrawOptions{}
	op.GeoM.Translate(560, 460)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, msg, &text.GoTextFace{
		Source: mplusFaceSource,
		Size:   12,
	}, op)

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	config := mars.ConfigNOP94()
	sim, err := mars.NewReportingSimulator(config)
	if err != nil {
		log.Fatal(err)
	}
	rec := mars.NewStateRecorder(sim)
	sim.AddReporter(rec)
	// sim.AddReporter(mars.NewDebugReporter(sim))

	w1file, err := os.Open("warriors/k94/julietstorm.red")
	if err != nil {
		log.Fatal(err)
	}
	w1data, err := mars.ParseLoadFile(w1file, config)
	if err != nil {
		log.Fatal(err)
	}
	w1file.Close()

	w2file, err := os.Open("warriors/k94/timescape10.red")
	if err != nil {
		log.Fatal(err)
	}
	w2data, err := mars.ParseLoadFile(w2file, config)
	if err != nil {
		log.Fatal(err)
	}
	w2file.Close()

	sim.AddWarrior(&w1data)
	sim.AddWarrior(&w2data)

	sim.SpawnWarrior(0, 0)
	sim.SpawnWarrior(1, 5555)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("gMARS")

	game := &Game{
		sim:       sim,
		rec:       *rec,
		speedStep: defaultSpeedStep,
	}
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

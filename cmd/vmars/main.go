package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"

	"github.com/bobertlo/gmars/pkg/mars"
	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

const (
	tileSize = 6
)

type Game struct {
	counter int
	sim     mars.ReportingSimulator
	rec     mars.StateRecorder
}

var (
	mplusFaceSource *text.GoTextFaceSource

	//go:embed assets/tiles_6.png
	tiles_png []byte

	tilesImage *ebiten.Image
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

func (g *Game) Update() error {
	for i := 0; i < 100; i++ {
		g.sim.RunCycle()
	}
	g.counter++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	w := tilesImage.Bounds().Dx()
	tileXCount := w / tileSize

	const xCount = screenWidth / tileSize

	for i := 0; i < int(g.sim.CoreSize()); i++ {
		state, _ := g.rec.GetMemState(mars.Address(i))
		// fmt.Println(i, state)
		if state == mars.CoreEmpty {
			continue
		}
		t := int(state)

		op := &ebiten.DrawImageOptions{}
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

	w1file, err := os.Open("warriors/94/blur2.rc")
	if err != nil {
		log.Fatal(err)
	}
	w1data, err := mars.ParseLoadFile(w1file, config)
	if err != nil {
		log.Fatal(err)
	}
	w1file.Close()

	w2file, err := os.Open("warriors/94/npaper2.rc")
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
		sim: sim,
		rec: *rec,
	}
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

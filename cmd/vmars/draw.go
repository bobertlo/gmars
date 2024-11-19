package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/bobertlo/gmars"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func (g *Game) Draw(screen *ebiten.Image) {
	warriorColors := make([]ebiten.ColorScale, 3)
	warriorColors[0].Scale(1, 1, 1, 1)
	warriorColors[1].Scale(1, 1, 0, 1)
	warriorColors[2].Scale(0, 1, 1, 1)

	warriorQueueColors := make([]ebiten.ColorScale, 2)
	warriorQueueColors[0].Scale(0.5, 0.5, 0.15, 1)
	warriorQueueColors[1].Scale(0.15, 0.5, 0.5, 1)

	execColor := ebiten.ColorScale{}
	execColor.Scale(1, 0.25, 0.25, 1)

	w := tilesImage.Bounds().Dx()
	tileXCount := w / tileSize

	const xCount = screenWidth / tileSize

	for i := 0; i < int(g.sim.CoreSize()); i++ {
		state, color := g.rec.GetMemState(gmars.Address(i))

		if state == gmars.CoreEmpty {
			continue
		}
		t := int(state)

		op := &ebiten.DrawImageOptions{ColorScale: warriorColors[color+1]}
		op.GeoM.Translate(float64((i%xCount)*tileSize), float64((i/xCount)*tileSize))

		sx := (t % tileXCount) * tileSize
		sy := (t / tileXCount) * tileSize
		screen.DrawImage(tilesImage.SubImage(image.Rect(sx, sy, sx+tileSize, sy+tileSize)).(*ebiten.Image), op)
	}

	for i := 0; i < int(g.sim.WarriorCount()); i++ {
		t := 0

		w := g.sim.GetWarrior(i)
		if w == nil || !w.Alive() {
			continue
		}

		for _, pc := range w.Queue()[1:] {

			op := &ebiten.DrawImageOptions{ColorScale: warriorQueueColors[i]}
			op.GeoM.Translate(float64((pc%xCount)*tileSize), float64((pc/xCount)*tileSize))

			sx := (t % tileXCount) * tileSize
			sy := (t / tileXCount) * tileSize
			screen.DrawImage(tilesImage.SubImage(image.Rect(sx, sy, sx+tileSize, sy+tileSize)).(*ebiten.Image), op)
		}

		t = 1
		pc, _ := w.NextPC()
		op := &ebiten.DrawImageOptions{ColorScale: execColor}
		op.GeoM.Translate(float64((pc%xCount)*tileSize), float64((pc/xCount)*tileSize))

		sx := (t % tileXCount) * tileSize
		sy := (t / tileXCount) * tileSize
		screen.DrawImage(tilesImage.SubImage(image.Rect(sx, sy, sx+tileSize, sy+tileSize)).(*ebiten.Image), op)

	}

	// Draw info
	msg := fmt.Sprintf("FPS: %0.2f", ebiten.ActualTPS())
	op := &text.DrawOptions{}
	op.GeoM.Translate(585, 465)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, msg, &text.GoTextFace{
		Source: mplusFaceSource,
		Size:   10,
	}, op)

	msg = fmt.Sprintf("Cycle: %05d (%dx)", g.sim.CycleCount(), speeds[g.speedStep])
	op = &text.DrawOptions{}
	op.GeoM.Translate(5, 465)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, msg, &text.GoTextFace{
		Source: mplusFaceSource,
		Size:   10,
	}, op)

	// draw results if finished
	if g.finished {
		if g.sim.WarriorCount() > 1 {
			w1a := g.sim.GetWarrior(0).Alive()
			w2a := g.sim.GetWarrior(1).Alive()

			if w1a || w2a {

				var msg string
				op = &text.DrawOptions{}
				op.GeoM.Translate(115, 465)
				if w1a && w2a {
					op.ColorScale = warriorColors[0]
					msg = "tie"
				} else if w1a {
					op.ColorScale = warriorColors[1]
					msg = fmt.Sprintf("%s wins", g.sim.GetWarrior(0).Name())
				} else if w2a {
					op.ColorScale = warriorColors[2]
					msg = fmt.Sprintf("%s wins", g.sim.GetWarrior(1).Name())
				}
				text.Draw(screen, msg, &text.GoTextFace{
					Source: mplusFaceSource,
					Size:   10,
				}, op)
			}
		}
	} else if !g.running {
		op = &text.DrawOptions{}
		op.GeoM.Translate(115, 465)
		text.Draw(screen, "PAUSED", &text.GoTextFace{
			Source: mplusFaceSource,
			Size:   10,
		}, op)
	}

}

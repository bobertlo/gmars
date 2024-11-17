package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/bobertlo/gmars"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/colornames"
)

const (
	SpriteFirst = iota
	SpriteHead
	SpriteHeadActive
	SpriteExecuted
	SpriteCode
	SpriteData
	SpriteIncr
	SpriteDecr
	SpriteRead
	SpriteDead
	SpriteLast
)

const screenXTileCount = screenWidth / tileSize
const minimumHeadTransparency = 0.5
const offsetHeadSprites = tileSize / 2.0

// cacheSprites buffers specific tiles from the tile image in different hue
// variants.  This helps performance a lot, because we don't need to shuffle so
// much memory on ever Update anymore.
func (g *Game) cacheSprites() {
	g.spriteCache = make([][]*ebiten.Image, SpriteLast)
	sprnum := SpriteFirst + 1

	tileXCount := tilesImage.Bounds().Dx() / tileSize
	for y := 0; y < tileXCount; y++ {
		for x := 0; x < tileXCount; x++ {
			sx := x * tileSize
			sy := y * tileSize
			rect := image.Rect(sx, sy, sx+tileSize, sy+tileSize)

			subImage := tilesImage.SubImage(rect).(*ebiten.Image)

			g.spriteCache[sprnum] = make([]*ebiten.Image, len(g.hues))
			for i, hue := range g.hues {
				var c colorm.ColorM
				c.ChangeHSV(hue/180.0*math.Pi, 1.0, 1.0)
				g.spriteCache[sprnum][i] = ebiten.NewImage(tileSize, tileSize)
				colorm.DrawImage(g.spriteCache[sprnum][i], subImage, c, nil)
			}

			sprnum++
			if sprnum >= SpriteLast {
				goto stop
			}
		}
	}
stop:
}

// cacheColors precomputes colors to use for static display, such as text.  This
// color matches dominant color of the sprite tileset.
func (g *Game) cacheColors() {
	g.colorCache = make([]color.Color, SpriteLast-1)
	// This color should match the tiles_X.png sprite
	baseColor := color.RGBA{106, 190, 48, 255}
	for i, hue := range g.hues {
		var cm colorm.ColorM
		cm.ChangeHSV(hue/180.0*math.Pi, 1.0, 1.0)
		g.colorCache[i] = cm.Apply(baseColor)
	}
}

// blitSpriteAlpha draws specific sprite to memory cell with alpha blending,
// taking into account sprite specifics.
func (g *Game) blitSpriteAlpha(screen *ebiten.Image, sprnum int, a gmars.Address, hue int, alpha float32) {
	var offsetY float64
	if sprnum != SpriteHead && sprnum != SpriteHeadActive && sprnum != SpriteDead {
		offsetY = offsetHeadSprites
	}
	var scale ebiten.ColorScale
	scale.ScaleAlpha(alpha)
	op := &ebiten.DrawImageOptions{ColorScale: scale}
	op.GeoM.Translate(float64((a%screenXTileCount)*tileSize), float64((a/screenXTileCount)*tileSize)+offsetY)
	screen.DrawImage(g.spriteCache[sprnum][hue], op)
}

// blitSprite draws specific sprite to memory cell taking into account sprite
// specifics.
func (g *Game) blitSprite(screen *ebiten.Image, sprnum int, a gmars.Address, hue int) {
	var offsetY float64
	if sprnum != SpriteHead && sprnum != SpriteHeadActive && sprnum != SpriteDead {
		offsetY = offsetHeadSprites
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64((a%screenXTileCount)*tileSize), float64((a/screenXTileCount)*tileSize)+offsetY)
	screen.DrawImage(g.spriteCache[sprnum][hue], op)
}

// queuePosition is used for sorting memory addresses by earlies position in the
// queue.
type queuePosition struct {
	PC               gmars.Address
	EarliestPosition int
}

func getSprite(state gmars.CoreState, instr gmars.Instruction) int {
	if state == gmars.CoreRead {
		return SpriteRead
	} else if state == gmars.CoreWritten {
		if instr.Op == gmars.DAT {
			return SpriteData
		} else {
			return SpriteCode
		}
	} else if state == gmars.CoreExecuted {
		return SpriteExecuted
	} else if state == gmars.CoreDecremented {
		return SpriteDecr
	} else if state == gmars.CoreIncremented {
		return SpriteIncr
	}

	return -1
}

func (g *Game) Draw(screen *ebiten.Image) {
	for i := 0; i < int(g.sim.CoreSize()); i++ {
		a := gmars.Address(i)
		mem := g.sim.GetMem(a)
		curState, curWar := g.rec.GetMemState(gmars.Address(i))

		sprite := getSprite(curState, mem)
		if sprite >= SpriteFirst {
			g.blitSprite(screen, sprite, a, curWar)
		}

		// Display CoreTerminated above the memory cell, as we do with heads.
		// We can then display the cause of death, if we look two steps behind.
		// (One step behind is always going to be warrior's own execute, so not interesting.)
		if curState == gmars.CoreTerminated {
			prevState, prevWar := g.rec.GetMemStateN(gmars.Address(i), 2)
			sprite := getSprite(prevState, mem)
			if sprite >= SpriteFirst {
				g.blitSprite(screen, sprite, a, prevWar)
			}

			g.blitSprite(screen, SpriteDead, a, curWar)
		}
	}

	wc := g.sim.WarriorCount()
	for wid := 0; wid < int(wc); wid++ {
		w := g.sim.GetWarrior(wid)
		if w == nil || !w.Alive() {
			continue
		}

		// The idea here is to show the PC queue using transparency.  Since
		// warriors often SPL thousands of times, often on top of the same
		// memory cells, we instead opt to show how soon *any* PC (head) on a
		// memory cell will execute.  So if a cell has a thousand heads, but
		// none of them is scheduled any time soon, the head will be shown
		// semi-transparent.
		earliestOccurence := make(map[gmars.Address]int)
		position := 0
		for _, pc := range w.Queue()[1:] {
			if _, ok := earliestOccurence[pc]; !ok {
				earliestOccurence[pc] = position
			}
			position++
		}

		queuePositions := make([]queuePosition, len(earliestOccurence))
		i := 0
		for pc, occurence := range earliestOccurence {
			queuePositions[i] = queuePosition{
				PC:               pc,
				EarliestPosition: occurence,
			}
			i++
		}

		sort.Slice(queuePositions, func(i, j int) bool {
			return queuePositions[i].EarliestPosition < queuePositions[j].EarliestPosition
		})

		valDecr := (1.0 - minimumHeadTransparency) / float64(len(queuePositions))
		val := 1.0
		for _, qp := range queuePositions {
			g.blitSpriteAlpha(screen, SpriteHead, qp.PC, wid, float32(val*val))
			val -= valDecr
		}

		pc, _ := w.NextPC()
		g.blitSprite(screen, SpriteHeadActive, pc, wid)
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
		w1a := g.sim.GetWarrior(0).Alive()
		w2a := g.sim.GetWarrior(1).Alive()

		if w1a || w2a {

			var msg string
			color := color.Color(colornames.White)
			op = &text.DrawOptions{}
			op.GeoM.Translate(115, 465)
			if w1a && w2a {
				msg = "tie"
			} else if w1a {
				color = g.colorCache[0]
				msg = fmt.Sprintf("%s wins", g.sim.GetWarrior(0).Name())
			} else if w2a {
				color = g.colorCache[1]
				msg = fmt.Sprintf("%s wins", g.sim.GetWarrior(1).Name())
			}

			op.ColorScale.ScaleWithColor(color)
			text.Draw(screen, msg, &text.GoTextFace{
				Source: mplusFaceSource,
				Size:   10,
			}, op)
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

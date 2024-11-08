package main

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/bobertlo/gmars/pkg/mars"
	"golang.org/x/image/colornames"
)

const cellSize = 4

type gameRenderer struct {
	render   *canvas.Raster
	objects  []fyne.CanvasObject
	imgCache *image.RGBA

	emptyColor       color.Color
	writtenColor     color.Color
	executedColor    color.Color
	readColor        color.Color
	incrementedColor color.Color
	decrementedColor color.Color

	game *game
}

func (g *gameRenderer) MinSize() fyne.Size {
	return fyne.NewSize(float32(g.game.widthInCells()*cellSize), float32(g.game.heightInCells()*cellSize))
}

func (g *gameRenderer) Layout(size fyne.Size) {
	g.render.Resize(size)
}

func (g *gameRenderer) ApplyTheme() {
	g.emptyColor = theme.Color(theme.ColorNameBackground)
	g.writtenColor = theme.Color(theme.ColorNameForeground)
	g.executedColor = theme.Color(theme.ColorGreen)
	g.readColor = theme.Color(theme.ColorBlue)
	g.incrementedColor = theme.Color(theme.ColorOrange)
	g.decrementedColor = theme.Color(theme.ColorPurple)
}

func (g *gameRenderer) Refresh() {
	canvas.Refresh(g.render)
}

func (g *gameRenderer) Objects() []fyne.CanvasObject {
	return g.objects
}

func (g *gameRenderer) Destroy() {
}

func (g *gameRenderer) draw(w, h int) image.Image {
	pixDensity := g.game.pixelDensity()
	pixW, pixH := g.game.cellForCoord(w, h, pixDensity)

	img := g.imgCache
	if img == nil || img.Bounds().Size().X != pixW || img.Bounds().Size().Y != pixH {
		img = image.NewRGBA(image.Rect(0, 0, pixW, pixH))
		g.imgCache = img
	}

	var memState mars.CoreState
	for y := 0; y < int(g.game.heightInCells()); y++ {
		for x := 0; x < int(g.game.widthInCells()); x++ {
			addr := y*80 + x
			if addr < int(g.game.sim.CoreSize()) {
				memState, _ = g.game.rec.GetMemState(mars.Address(addr))

				if memState == mars.CoreEmpty {
					img.Set(x, y, g.emptyColor)
				} else if memState == mars.CoreWritten {
					img.Set(x, y, g.writtenColor)
				} else if memState == mars.CoreExecuted {
					img.Set(x, y, g.executedColor)
				} else if memState == mars.CoreRead {
					img.Set(x, y, g.readColor)
				} else if memState == mars.CoreIncremented {
					img.Set(x, y, g.incrementedColor)
				} else if memState == mars.CoreDecremented {
					img.Set(x, y, g.decrementedColor)
				} else {
					img.Set(x, y, g.emptyColor)
				}
			}
		}
	}

	return img
}

type game struct {
	widget.BaseWidget

	genText *widget.Label
	sim     mars.Simulator
	rec     *mars.StateRecorder
	paused  bool
	speed   float64
}

func (g *game) CreateRenderer() fyne.WidgetRenderer {
	renderer := &gameRenderer{game: g}

	render := canvas.NewRaster(renderer.draw)
	render.ScaleMode = canvas.ImageScalePixels
	renderer.render = render
	renderer.objects = []fyne.CanvasObject{render}
	renderer.ApplyTheme()

	return renderer
}

func (g *game) cellForCoord(x, y int, density float32) (int, int) {
	xpos := int(float32(x) / float32(cellSize) / density)
	ypos := int(float32(y) / float32(cellSize) / density)

	return xpos, ypos
}

func (g *game) toggleRun() {
	g.paused = !g.paused
}

func (g *game) animate() {
	go func() {
		tick := time.NewTicker(time.Second / 6)

		for range tick.C {
			if g.paused {
				continue
			}

			g.sim.RunCycle()
			g.updateGeneration()
			g.Refresh()

			if g.speed == 1000.0 {
				tick.Reset(time.Nanosecond)
			} else {
				tick.Reset(time.Duration(1000/g.speed) * time.Millisecond)
			}
		}
	}()
}

func (g *game) typedRune(r rune) {
	if r == ' ' {
		g.toggleRun()
	}
}

func (g *game) Tapped(ev *fyne.PointEvent) {
}

func (g *game) TappedSecondary(ev *fyne.PointEvent) {
}

func (g *game) buildUI() fyne.CanvasObject {
	var pause *widget.Button
	pause = widget.NewButton("Pause", func() {
		g.paused = !g.paused

		if g.paused {
			pause.SetText("Resume")
		} else {
			pause.SetText("Pause")
		}
	})

	slider := &widget.Slider{Step: 1, Min: 1, Max: 1000, OnChanged: func(f float64) {
		g.speed = f
	}}

	title := container.NewGridWithColumns(2, g.genText, pause, slider)
	return container.NewBorder(title, nil, nil, nil, g)
}

func (g *game) updateGeneration() {
	g.genText.SetText(fmt.Sprintf("Cycle %d", g.sim.CycleCount()))
}

func (g *game) pixelDensity() float32 {
	c := fyne.CurrentApp().Driver().CanvasForObject(g)
	if c == nil {
		return 1.0
	}

	pixW, _ := c.PixelCoordinateForPosition(fyne.NewPos(cellSize, cellSize))
	return float32(pixW) / float32(cellSize)
}

func (g *game) widthInCells() uint64 {
	return uint64(80)
}
func (g *game) heightInCells() uint64 {
	return (uint64(g.sim.CoreSize()) / 80) + 1
}

func newGame(sim mars.ReportingSimulator) *game {
	rec := mars.NewStateRecorder(sim)
	sim.AddReporter(rec)

	g := &game{sim: sim, rec: rec, speed: 1.0, genText: widget.NewLabel("Generation 0")}
	g.ExtendBaseWidget(g)

	return g
}

type gameTheme struct {
}

var _ fyne.Theme = (*gameTheme)(nil)

func (m gameTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameBackground {
		if variant == theme.VariantLight {
			return color.White
		}
		return color.Black
	}

	if name == theme.ColorGreen {
		return colornames.Green
	}
	if name == theme.ColorBlue {
		return colornames.Blue
	}
	if name == theme.ColorOrange {
		return colornames.Orange
	}
	if name == theme.ColorPurple {
		return colornames.Purple
	}

	return theme.DefaultTheme().Color(name, variant)
}

func (m gameTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m gameTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m gameTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

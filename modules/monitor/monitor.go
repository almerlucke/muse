package monitor

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/almerlucke/muse"
	"github.com/fogleman/gg"
)

type Monitor struct {
	*muse.BaseModule
	context *gg.Context
	width   int
	height  int
	raster  *canvas.Raster
}

func NewMonitor(width int, height int) *Monitor {
	ctx := gg.NewContext(width, height)

	raster := canvas.NewRasterFromImage(ctx.Image())
	raster.ScaleMode = canvas.ImageScaleFastest

	m := &Monitor{
		BaseModule: muse.NewBaseModule(1, 0),
		context:    ctx,
		raster:     raster,
		width:      width,
		height:     height,
	}

	m.SetSelf(m)

	return m
}

func (m *Monitor) MustSynthesize() bool {
	return true
}

func (m *Monitor) Synthesize() bool {
	if !m.BaseModule.Synthesize() {
		return false
	}

	m.context.SetRGB(1, 1, 1)
	m.context.Clear()
	m.context.SetRGB(0, 0, 0)
	m.context.SetLineWidth(1.0)

	yHalf := float64(m.height) / 2.0
	xStep := float64(m.width) / float64(m.Config.BufferSize)

	for i := 0; i < m.Config.BufferSize; i++ {
		if i == 0 {
			m.context.MoveTo(float64(i)*xStep, yHalf+m.Inputs[0].Buffer[i]*yHalf)
		} else {
			m.context.LineTo(float64(i)*xStep, yHalf+m.Inputs[0].Buffer[i]*yHalf)
		}
	}

	m.context.Stroke()
	m.raster.Refresh()

	return true
}

func (m *Monitor) UI() fyne.CanvasObject {
	return m.raster
}

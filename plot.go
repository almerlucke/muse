package muse

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

type PlotModule struct {
	*BaseModule
	samples []float64
	index   int
}

func NewPlotModule(n int, config *Configuration) *PlotModule {
	pc := &PlotModule{
		BaseModule: NewBaseModule(1, 0, config, ""),
		samples:    make([]float64, n),
	}

	pc.SetSelf(pc)

	return pc
}

func (pm *PlotModule) ReceiveControlValue(value any, index int) {
	pm.samples[pm.index] = value.(float64)
	pm.index++
}

func (pm *PlotModule) points(asControl bool) plotter.XYs {
	timeStep := 1.0

	if asControl {
		timeStep = (float64(pm.Config.BufferSize) / pm.Config.SampleRate) * 1000.0
	}

	pts := make(plotter.XYs, len(pm.samples))

	for i, sample := range pm.samples {
		pts[i] = plotter.XY{X: float64(i) * timeStep, Y: sample}
	}

	return pts
}

func (pm *PlotModule) MustSynthesize() bool {
	return true
}

func (pm *PlotModule) Synthesize() bool {
	if !pm.BaseModule.Synthesize() {
		return false
	}

	if pm.Inputs[0].IsConnected() {
		for i := 0; i < pm.Config.BufferSize; i++ {
			pm.samples[pm.index] = pm.Inputs[0].Buffer[i]
			pm.index++
		}
	}

	return true
}

func (pm *PlotModule) Save(w float64, h float64, asControl bool, filePath string) error {
	p := plot.New()

	l, _ := plotter.NewLine(pm.points(asControl))

	p.Add(l)

	wp := vg.Points(w)
	hp := vg.Points(h)

	p.Draw(draw.New(vgimg.New(wp, hp)))

	return p.Save(wp, hp, filePath)
}

func PlotVector(v []float64, w float64, h float64, filePath string) error {
	pts := make(plotter.XYs, len(v))

	for i, sample := range v {
		pts[i] = plotter.XY{X: float64(i), Y: sample}
	}

	return PlotPoints(pts, w, h, filePath)
}

func PlotPoints(points plotter.XYs, w float64, h float64, filePath string) error {
	p := plot.New()

	l, _ := plotter.NewLine(points)

	p.Add(l)

	wp := vg.Points(w)
	hp := vg.Points(h)

	p.Draw(draw.New(vgimg.New(wp, hp)))

	return p.Save(wp, hp, filePath)
}

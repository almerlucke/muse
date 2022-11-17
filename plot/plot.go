package plot

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

func PlotVector(v []float64, w float64, h float64, filePath string) error {
	pts := make(plotter.XYs, len(v))

	for i, sample := range v {
		pts[i] = plotter.XY{X: float64(i), Y: sample}
	}

	return PlotLine(pts, w, h, filePath)
}

func PlotLine(points plotter.XYs, w float64, h float64, filePath string) error {
	p := plot.New()

	l, _ := plotter.NewLine(points)

	p.Add(l)

	wp := vg.Points(w)
	hp := vg.Points(h)

	p.Draw(draw.New(vgimg.New(wp, hp)))

	return p.Save(wp, hp, filePath)
}

func PlotPoints(points plotter.XYs, w float64, h float64, filePath string) error {
	p := plot.New()

	sc, _ := plotter.NewScatter(points)
	p.Add(sc)

	wp := vg.Points(w)
	hp := vg.Points(h)

	p.Draw(draw.New(vgimg.New(wp, hp)))

	return p.Save(wp, hp, filePath)
}

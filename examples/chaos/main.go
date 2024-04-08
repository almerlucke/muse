package main

import (
	"github.com/almerlucke/genny/float/interp"
	"github.com/almerlucke/genny/float/iter"
	"github.com/almerlucke/genny/float/iter/updaters/chaos"
	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/float/shape/shapers/linear"
	"github.com/almerlucke/genny/float/shape/shapers/mirror"
	"github.com/almerlucke/genny/float/shape/shapers/series"

	"log"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/controls/fgen"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/modules/generator"
	"github.com/almerlucke/muse/modules/osc"
	"github.com/almerlucke/muse/plot"

	"gonum.org/v1/plot/plotter"
)

func genPlot() {
	f := func(x float64) float64 { return x * x } // Aronson adjusted
	iter := iter.New([]float64{0.4, 0.4}, chaos.NewAronsonWithFunc(1.978, f))

	pts := make(plotter.XYs, 20000)
	vecX := make([]float64, 20000)
	vecY := make([]float64, 20000)

	for i := 0; i < 20000; i++ {
		v := iter.Generate()
		pts[i].X = v[0]
		pts[i].Y = v[1]
		vecX[i] = v[0]
		vecY[i] = v[1]
		log.Printf("x: %f", v[0])
		log.Printf("y: %f", v[1])
	}

	_ = plot.Vector(vecX[:1600], 1600, 400, "/Users/almerlucke/Desktop/aronX.png")
	_ = plot.Vector(vecY[:1600], 1600, 400, "/Users/almerlucke/Desktop/aronY.png")
	_ = plot.Points(pts, 400, 400, "/Users/almerlucke/Desktop/aron.png")
}

func genSound() {
	root := muse.New(1)

	f := func(x float64) float64 { return x * x } // Aronson adjusted
	it := iter.New([]float64{0.4, 0.4}, chaos.NewAronsonWithFunc(1.878, f))
	mir := mirror.New(-1.0, 1.0)
	wrapper := interp.New(
		shape.New(it, mir),
		interp.Cubic,
		1.0/60,
	)

	sgen := root.AddModule(generator.NewBasic(wrapper))

	sgen.Connect(0, root, 0)
	// sgen.Connect(1, env, 1)

	_ = root.RenderAudio()
}

func genFreq() {
	root := muse.New(1)

	f := func(x float64) float64 { return x * x } // Aronson adjusted
	aron := chaos.NewAronsonWithFunc(1.698, f)
	it := iter.New([]float64{0.4, 0.4}, aron)
	mir := mirror.New(-1.0, 1.0)
	uni := linear.NewUnipolar()
	scale := linear.New(1400.0, 50.0)
	chain := series.New(mir, uni, scale)
	wrapper := interp.New(
		shape.New(it, chain),
		interp.Cubic,
		1.0/10.0,
	)

	cgen := root.AddControl(fgen.NewX(wrapper, func(value any, index int) {
		if index == 0 {
			aron.A = value.(float64)
		} else if index == 1 {
			scale.Scale = value.(float64)
		}
	}, nil))

	chaosLfo := root.AddControl(lfo.NewBasicControlLFO(0.089, 1.467, 1.998))
	freqLfo := root.AddControl(lfo.NewBasicControlLFO(0.067, 200.0, 2300.0))

	chaosLfo.CtrlConnect(0, cgen, 0)
	freqLfo.CtrlConnect(0, cgen, 1)

	osc1 := root.AddModule(osc.New(100.0, 0.0))

	cgen.CtrlConnect(1, osc1, 0)

	osc1.Connect(2, root, 0)

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/lozi.aiff", 20.0, 44100.0, false, sndfile.SF_FORMAT_AIFF)
	_ = root.RenderAudio()
}

func main() {
	genFreq()
}

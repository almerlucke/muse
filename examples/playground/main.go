package main

import (
	"log"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/interpolator"
	"github.com/almerlucke/muse/components/iterator"
	"github.com/almerlucke/muse/components/iterator/chaos"
	"github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/controls/gen"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/modules/blosc"
	"github.com/almerlucke/muse/modules/generator"
	"github.com/almerlucke/muse/plot"
	"gonum.org/v1/plot/plotter"
)

func genPlot() {
	f := func(x float64) float64 { return x * x } // Aronson adjusted
	iter := iterator.NewIterator([]float64{0.4, 0.4}, chaos.NewAronsonWithFunc(1.978, f))

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

	plot.PlotVector(vecX[:1600], 1600, 400, "/Users/almerlucke/Desktop/aronX.png")
	plot.PlotVector(vecY[:1600], 1600, 400, "/Users/almerlucke/Desktop/aronY.png")
	plot.PlotPoints(pts, 400, 400, "/Users/almerlucke/Desktop/aron.png")
}

func genSound() {
	env := muse.NewEnvironment(1, 44100.0, 128)

	f := func(x float64) float64 { return x * x } // Aronson adjusted
	iter := iterator.NewIterator([]float64{0.4, 0.4}, chaos.NewAronsonWithFunc(1.878, f))
	mirror := waveshaping.NewMirror(-1.0, 1.0)
	wrapper := interpolator.NewInterpolator(
		waveshaping.NewGeneratorWrapper(iter, []waveshaping.Shaper{mirror, mirror}),
		interpolator.Cubic,
		60,
	)

	sgen := env.AddModule(generator.NewBasicGenerator(wrapper, env.Config, ""))

	sgen.Connect(0, env, 0)
	// sgen.Connect(1, env, 1)

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/aronson.aiff", 20.0, 44100.0, false, sndfile.SF_FORMAT_AIFF)
	env.QuickPlayAudio()
}

func genFreq() {
	env := muse.NewEnvironment(1, 44100.0, 128)

	f := func(x float64) float64 { return x * x } // Aronson adjusted
	aron := chaos.NewAronsonWithFunc(1.698, f)
	iter := iterator.NewIterator([]float64{0.4, 0.4}, aron)
	mirror := waveshaping.NewMirror(-1.0, 1.0)
	uni := waveshaping.NewUnipolar()
	scale := waveshaping.NewLinear(1400.0, 50.0)
	chain := waveshaping.NewChain(mirror, uni, scale)
	wrapper := interpolator.NewInterpolator(
		waveshaping.NewGeneratorWrapper(iter, []waveshaping.Shaper{chain, chain}),
		interpolator.Cubic,
		5,
	)

	cgen := env.AddControl(gen.NewGen(wrapper, func(value any, index int) {
		if index == 0 {
			aron.A = value.(float64)
		} else if index == 1 {
			scale.Scale = value.(float64)
		}
	}, nil, ""))

	chaosLfo := env.AddControl(lfo.NewBasicControlLFO(0.089, 1.467, 1.998, env.Config, ""))
	freqLfo := env.AddControl(lfo.NewBasicControlLFO(0.067, 200.0, 2300.0, env.Config, ""))

	chaosLfo.CtrlConnect(0, cgen, 0)
	freqLfo.CtrlConnect(0, cgen, 1)

	osc1 := env.AddModule(blosc.NewOsc(100.0, 0.0, env.Config, ""))

	cgen.CtrlConnect(1, osc1, 0)

	osc1.Connect(2, env, 0)

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/lozi.aiff", 20.0, 44100.0, false, sndfile.SF_FORMAT_AIFF)
	env.QuickPlayAudio()
}

func main() {
	genFreq()
}

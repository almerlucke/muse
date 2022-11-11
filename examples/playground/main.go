package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/phasor"
	"github.com/almerlucke/muse/components/waveshaping"
)

func main() {
	phasor := phasor.NewPhasor(100.0, 44100.0, 0.0)
	chain := waveshaping.NewChain(waveshaping.NewSineTable(512), waveshaping.NewMult(2.3146), waveshaping.NewMirror(-0.6, 0.6))

	xs := make([]float64, 1000)
	for i := 0; i < 1000; i++ {
		xs[i] = chain.Shape(phasor.Tick())
	}

	muse.PlotVector(xs, 400, 200, "/Users/almerlucke/Desktop/mirror.png")
}

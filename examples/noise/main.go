package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/noise"
	"github.com/mkb218/gosndfile/sndfile"
)

func main() {
	env := muse.NewEnvironment(1, 44100.0, 512)

	n := env.AddModule(noise.NewNoise(1, env.Config))
	m := env.AddModule(functor.NewFunctor(1, func(v []float64) float64 { return v[0] * 0.1 }, env.Config))

	muse.Connect(n, 0, m, 0)
	muse.Connect(m, 0, env, 0)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/noise.aiff", 10.0, 44100.0, false, sndfile.SF_FORMAT_AIFF)
}

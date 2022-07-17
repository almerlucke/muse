package main

import (
	"math/rand"
	"time"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/mkb218/gosndfile/sndfile"

	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/blosc"
	"github.com/almerlucke/muse/modules/functor"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	env := muse.NewEnvironment(1, 44100.0, 512)

	adsrAmp := env.AddModule(adsr.NewADSR([]adsrc.Step{
		{Level: 1.0, Duration: 20, Shape: 0.0}, {Level: 0.3, Duration: 20, Shape: 0.0},
		{Duration: 20},
		{Duration: 20, Shape: 0.1}}, adsrc.Absolute, adsrc.Automatic, 1.0, env.Config, "adsr"))
	osc := env.AddModule(blosc.NewBloscModule(300.0, 0, 1, env.Config, "osc"))
	mul := env.AddModule(functor.NewFunctor(2, functor.FunctorMult, env.Config, "mul"))

	env.AddMessenger(stepper.NewStepper(swing.New(120.0, 4.0, []*swing.Step{
		{}, {Shuffle: 0.3}, {Skip: true}, {Shuffle: 0.3, ShuffleRand: 0.2}, {Skip: true}, {Shuffle: 0.1}, {}, {SkipFactor: 0.4, Shuffle: 0.2},
	}), []string{"adsr"}, ""))

	muse.Connect(adsrAmp, 0, mul, 0)
	muse.Connect(osc, 0, mul, 1)
	muse.Connect(mul, 0, env, 0)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/beats.aiff", 10.0, env.Config.SampleRate, sndfile.SF_FORMAT_AIFF)
}

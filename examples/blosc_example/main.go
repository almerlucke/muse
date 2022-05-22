package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/sequencer"
	"github.com/almerlucke/muse/messengers/stepper"
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/blosc"
	"github.com/almerlucke/muse/modules/common"
	"github.com/almerlucke/muse/modules/filters/moog/stilson"
	"github.com/mkb218/gosndfile/sndfile"
)

func main() {
	env := muse.NewEnvironment(2, 44100, 128)

	sequence1, _ := sequencer.ReadSequence("examples/blosc_example/sequence1.json")
	sequence2, _ := sequencer.ReadSequence("examples/blosc_example/sequence2.json")

	env.AddMessenger(sequencer.NewSequencer(sequence1, "sequencer1"))
	env.AddMessenger(sequencer.NewSequencer(sequence2, "sequencer2"))

	env.AddMessenger(stepper.NewStepper(
		stepper.NewSliceProvider([]float64{250, -125, 250, 250, -125, 125, -125, 250}),
		[]string{"sequencer1", "adsr1"}, "",
	))
	env.AddMessenger(stepper.NewStepper(
		stepper.NewSliceProvider([]float64{-125, 125, 125, 125, -125, 250, -125}),
		[]string{"sequencer2", "adsr2"}, "",
	))

	steps := []adsr.ADSRStep{
		{LevelRatio: 1.0, DurationRatio: 0.1, Shape: 0.1},
		{LevelRatio: 0.4, DurationRatio: 0.1, Shape: -0.1},
		{DurationRatio: 0.1},
		{DurationRatio: 0.3, Shape: -0.1},
	}

	adsrEnv1 := env.AddModule(adsr.NewADSRModule(steps, 1.0, env.Config, "adsr1"))
	adsrEnv2 := env.AddModule(adsr.NewADSRModule(steps, 1.0, env.Config, "adsr2"))
	mult1 := env.AddModule(common.NewMult(2, env.Config, ""))
	mult2 := env.AddModule(common.NewMult(2, env.Config, ""))
	osc1 := env.AddModule(blosc.NewBloscModule(100.0, 0.0, 1.0, env.Config, "blosc1"))
	osc2 := env.AddModule(blosc.NewBloscModule(400.0, 0.0, 1.0, env.Config, "blosc2"))
	filter1 := env.AddModule(stilson.NewStilsonMoog(600.0, 0.3, env.Config, "filter1"))
	filter2 := env.AddModule(stilson.NewStilsonMoog(700.0, 0.3, env.Config, "filter2"))

	muse.Connect(osc1, 2, mult1, 0)
	muse.Connect(osc2, 2, mult2, 0)
	muse.Connect(adsrEnv1, 0, mult1, 1)
	muse.Connect(adsrEnv2, 0, mult2, 1)
	muse.Connect(mult1, 0, filter1, 0)
	muse.Connect(mult2, 0, filter2, 0)
	muse.Connect(filter1, 0, env, 0)
	muse.Connect(filter2, 0, env, 1)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/test.aiff", 15.0, sndfile.SF_FORMAT_AIFF)
}

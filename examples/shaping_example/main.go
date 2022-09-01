package main

import (
	"github.com/almerlucke/muse"

	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	shapingc "github.com/almerlucke/muse/components/shaping"
	"github.com/almerlucke/muse/utils"
	"github.com/almerlucke/muse/values"

	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/allpass"
	"github.com/almerlucke/muse/modules/filters/moog/moog1"
	"github.com/almerlucke/muse/modules/freeverb"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/phasor"
	"github.com/almerlucke/muse/modules/shaper"
	"github.com/almerlucke/muse/modules/vartri"
	"github.com/mkb218/gosndfile/sndfile"
)

func msgMapper(msg any, shaper shapingc.Shaper) {
	// params, ok := msg.(map[string]any)
	// if ok {
	// 	s, ok := params["sync"]
	// 	if ok {
	// 	}
	// }
}

func paramMapper(param int, value float64, shaper shapingc.Shaper) {
	if param == 0 {
		shaper.(*shapingc.Chain).SetSuperSawM1(value)
	}
}

func main() {
	env := muse.NewEnvironment(2, 3*44100, 512)

	sequence := values.NewSequence(utils.ReadJSONNull[[][]*muse.Message]("examples/shaping_example/sequence1.json"), true)

	env.AddMessenger(banger.NewValueGenerator(sequence, "sequencer1"))

	env.AddMessenger(stepper.NewStepper(
		stepper.NewValueStepper(values.NewSequence([]float64{250, -125, 250, 250, -125, 125, -125, 250}, true)),
		[]string{"sequencer1", "adsr1"}, "",
	))

	steps := []adsrc.Step{
		{Level: 1.0, Duration: 20, Shape: 0.0},
		{Level: 0.4, Duration: 20, Shape: 0.0},
		{Duration: 20},
		{Duration: 100, Shape: 0.0},
	}

	paramVarTri1 := env.AddModule(vartri.NewVarTri(0.25, 0.0, 0.5, env.Config, "vartri1"))
	paramVarTri2 := env.AddModule(vartri.NewVarTri(0.325, 0.0, 0.5, env.Config, "vartri2"))
	superSawParam := env.AddModule(functor.NewFunctor(1, func(vec []float64) float64 { return vec[0]*0.82 + 0.15 }, env.Config, "superSaw"))
	adsrEnv1 := env.AddModule(adsr.NewADSR(steps, adsrc.Absolute, adsrc.Automatic, 1.0, env.Config, "adsr1"))
	mult1 := env.AddModule(functor.NewFunctor(2, functor.FunctorMult, env.Config, ""))
	filterParam := env.AddModule(functor.NewFunctor(1, func(vec []float64) float64 { return vec[0]*2200.0 + 40.0 }, env.Config, ""))
	osc1 := env.AddModule(phasor.NewPhasor(140.0, 0.0, env.Config, "osc1"))
	shaper1 := env.AddModule(shaper.NewShaper(shapingc.NewSuperSaw(), 1, paramMapper, nil, env.Config, "shaper1"))
	allpass := env.AddModule(allpass.NewAllpass(375.0, 375.0, 0.3, env.Config, "allpass"))
	allpassAmp := env.AddModule(functor.NewFunctor(1, func(vec []float64) float64 { return vec[0] * 0.5 }, env.Config, ""))
	// filter := env.AddModule(butterworth.NewButterworth(300.0, 0.4, env.Config, "filter"))
	// filter := env.AddModule(rbj.NewRBJFilter(rbjc.Lowpass, 300.0, 10.0, env.Config, "filter"))
	filter := env.AddModule(moog1.NewMoog1(300.0, 0.45, 1.75, env.Config, "filter"))
	reverb := env.AddModule(freeverb.NewFreeVerb(env.Config, "freeverb"))

	reverb.(*freeverb.FreeVerb).SetDamp(0.1)
	reverb.(*freeverb.FreeVerb).SetDry(0.7)
	reverb.(*freeverb.FreeVerb).SetWet(0.1)
	reverb.(*freeverb.FreeVerb).SetRoomSize(0.8)
	reverb.(*freeverb.FreeVerb).SetWidth(0.8)

	muse.Connect(paramVarTri1, 0, superSawParam, 0)
	muse.Connect(paramVarTri2, 0, filterParam, 0)
	muse.Connect(osc1, 0, shaper1, 0)
	muse.Connect(superSawParam, 0, shaper1, 1)
	muse.Connect(shaper1, 0, mult1, 0)
	muse.Connect(adsrEnv1, 0, mult1, 1)
	muse.Connect(mult1, 0, filter, 0)
	muse.Connect(filterParam, 0, filter, 1)
	muse.Connect(filter, 0, allpass, 0)
	muse.Connect(allpass, 0, allpassAmp, 0)
	muse.Connect(filter, 0, reverb, 0)
	muse.Connect(allpassAmp, 0, reverb, 1)
	muse.Connect(reverb, 0, env, 0)
	muse.Connect(reverb, 1, env, 1)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/shaper.aiff", 24.0, 44100.0, sndfile.SF_FORMAT_AIFF)
}

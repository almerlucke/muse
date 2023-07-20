package main

import (
	"github.com/almerlucke/muse"

	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	shaping "github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/utils"
	"github.com/almerlucke/muse/value"

	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/allpass"
	"github.com/almerlucke/muse/modules/filters/moog"
	"github.com/almerlucke/muse/modules/freeverb"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/phasor"
	"github.com/almerlucke/muse/modules/vartri"
	"github.com/almerlucke/muse/modules/waveshaper"
	"github.com/mkb218/gosndfile/sndfile"
)

func msgMapper(msg any, shaper shaping.Shaper) {
	// params, ok := msg.(map[string]any)
	// if ok {
	// 	s, ok := params["sync"]
	// 	if ok {
	// 	}
	// }
}

func paramMapper(param int, value any, shaper shaping.Shaper) {
	if param == 0 {
		shaper.(*shaping.Serial).SetSuperSawM1(value.(float64))
	}
}

func main() {
	env := muse.NewEnvironment(2, 3*44100, 512)

	sequence := value.NewSequence(utils.ReadJSONNull[[][]*muse.Message]("examples/shaping_example/sequence1.json"))

	env.AddMessenger(banger.NewValueGenerator(sequence, "sequencer1"))

	env.AddMessenger(stepper.NewStepper(
		stepper.NewValueStepProvider(value.NewSequence([]float64{250, -125, 250, 250, -125, 125, -125, 250})),
		[]string{"sequencer1", "adsr1"}, "",
	))

	steps := []adsrc.Step{
		{Level: 1.0, Duration: 20, Shape: 0.0},
		{Level: 0.4, Duration: 20, Shape: 0.0},
		{Duration: 20},
		{Duration: 100, Shape: 0.0},
	}

	paramVarTri1 := vartri.NewVarTri(0.25, 0.0, 0.5, env.Config).Named("vartri1").Add(env)
	paramVarTri2 := vartri.NewVarTri(0.325, 0.0, 0.5, env.Config).Named("vartri2").Add(env)
	superSawParam := env.AddModule(functor.NewFunctor(1, func(vec []float64) float64 { return vec[0]*0.82 + 0.15 }, env.Config))
	adsrEnv1 := env.AddModule(adsr.NewADSR(steps, adsrc.Absolute, adsrc.Automatic, 1.0, env.Config, "adsr1"))
	mult1 := env.AddModule(functor.NewFunctor(2, functor.FunctorMult, env.Config))
	filterParam := env.AddModule(functor.NewFunctor(1, func(vec []float64) float64 { return vec[0]*2200.0 + 40.0 }, env.Config))
	osc1 := env.AddModule(phasor.NewPhasor(140.0, 0.0, env.Config, "osc1"))
	shaper1 := env.AddModule(waveshaper.NewWaveShaper(shaping.NewSuperSaw(1.5, 0.25, 0.88), 1, paramMapper, nil, env.Config, "shaper1"))
	allpass := env.AddModule(allpass.NewAllpass(375.0, 375.0, 0.3, env.Config, "allpass"))
	allpassAmp := env.AddModule(functor.NewFunctor(1, func(vec []float64) float64 { return vec[0] * 0.5 }, env.Config))
	// filter := env.AddModule(butterworth.NewButterworth(300.0, 0.4, env.Config, "filter"))
	// filter := env.AddModule(rbj.NewRBJFilter(rbjc.Lowpass, 300.0, 10.0, env.Config, "filter"))
	filter := env.AddModule(moog.NewMoog(300.0, 0.45, 1.0, env.Config, "filter"))
	reverb := freeverb.NewFreeVerb(env.Config).Named("freeverb").Add(env)

	reverb.(*freeverb.FreeVerb).SetDamp(0.1)
	reverb.(*freeverb.FreeVerb).SetDry(0.7)
	reverb.(*freeverb.FreeVerb).SetWet(0.1)
	reverb.(*freeverb.FreeVerb).SetRoomSize(0.8)
	reverb.(*freeverb.FreeVerb).SetWidth(0.8)

	paramVarTri1.Connect(0, superSawParam, 0)
	paramVarTri2.Connect(0, filterParam, 0)
	osc1.Connect(0, shaper1, 0)
	superSawParam.Connect(0, shaper1, 1)
	shaper1.Connect(0, mult1, 0)
	adsrEnv1.Connect(0, mult1, 1)
	mult1.Connect(0, filter, 0)
	filterParam.Connect(0, filter, 1)
	filter.Connect(0, allpass, 0)
	allpass.Connect(0, allpassAmp, 0)
	filter.Connect(0, reverb, 0)
	allpassAmp.Connect(0, reverb, 1)
	reverb.Connect(0, env, 0)
	reverb.Connect(1, env, 1)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/shaper.aiff", 24.0, 44100.0, true, sndfile.SF_FORMAT_AIFF)
}

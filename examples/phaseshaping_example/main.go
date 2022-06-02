package main

import (
	"math"

	"github.com/almerlucke/muse"
	pshape "github.com/almerlucke/muse/components/phaseshaping"
	"github.com/almerlucke/muse/messengers/sequencer"
	"github.com/almerlucke/muse/messengers/stepper"
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/allpass"
	"github.com/almerlucke/muse/modules/filters/butterworth"
	"github.com/almerlucke/muse/modules/freeverb"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/phaseshaper"
	"github.com/almerlucke/muse/modules/vartri"
	"github.com/mkb218/gosndfile/sndfile"
)

func minimoogVoyagerSawtooth(sr float64) *pshape.PhaseDistortion {
	pd := pshape.NewPhaseDistortion(&pshape.Phasor{Delta: 140.0 / sr, Phase: 0.0})
	pd.Shapers = []pshape.Shaper{
		pshape.NewLinear(0.25, 0.0),
		pshape.NewFunction(func(s float64) float64 { return math.Sin(2.0 * math.Pi * s) }),
		pshape.NewBipolar(),
	}

	return pd
}

func hardSync(sr float64) *pshape.PhaseDistortion {
	pd := pshape.NewPhaseDistortion(&pshape.Phasor{Delta: 140.0 / sr, Phase: 0.0})
	pd.Shapers = []pshape.Shaper{
		pshape.NewLinear(2.5, 0.0),
		pshape.NewMod1(),
		pshape.NewBipolar()}

	return pd
}

func softSyncTriangle(sr float64) *pshape.PhaseDistortion {
	pd := pshape.NewPhaseDistortion(&pshape.Phasor{Delta: 80.0 / sr, Phase: 0.0})
	pd.Shapers = []pshape.Shaper{
		pshape.NewBipolar(),
		pshape.NewAbs(),
		pshape.NewLinear(1.25, 0.0),
		pshape.NewMod1(),
		pshape.NewTri(),
		pshape.NewBipolar(),
	}

	return pd
}

func jp8000triMod(sr float64) *pshape.PhaseDistortion {
	pd := pshape.NewPhaseDistortion(&pshape.Phasor{Delta: 140.0 / sr, Phase: 0.0})
	pd.Shapers = []pshape.Shaper{
		pshape.NewBipolar(),
		pshape.NewAbs(),
		pshape.NewLinear(2.0, -1.0),
		pshape.NewMod1(),
		pshape.NewMult(0.7),
		pshape.NewFunction(
			func(x float64) float64 {
				return 2.0 * (x - math.Ceil(x-0.5))
			},
		),
	}

	return pd
}

func pulseWidthMod(sr float64) *pshape.PhaseDistortion {
	pd := pshape.NewPhaseDistortion(&pshape.Phasor{Delta: 140.0 / sr, Phase: 0.0})
	pd.Shapers = []pshape.Shaper{
		pshape.NewLinear(1.25, 0.0),
		pshape.NewMod1(),
		pshape.NewPulse(0.4),
		pshape.NewBipolar(),
	}

	return pd
}

func msgMapper(msg any, dist *pshape.PhaseDistortion) {
	params, ok := msg.(map[string]any)
	if ok {
		s, ok := params["sync"]
		if ok {
			dist.Shapers[2].(*pshape.LinearShape).Scale = s.(float64)
		}
	}
}

func paramMapper(param int, value float64, dist *pshape.PhaseDistortion) {
	if param == 0 {
		lin := dist.Shapers[2].(*pshape.LinearShape)
		lin.Scale = value
	}
}

func main() {
	env := muse.NewEnvironment(2, 3*44100, 512)

	sequence1, _ := sequencer.ReadSequence("examples/phaseshaping_example/sequence1.json")

	env.AddMessenger(sequencer.NewSequencer(sequence1, "sequencer1"))

	env.AddMessenger(stepper.NewStepper(
		stepper.NewSliceProvider([]float64{250, -125, 250, 250, -125, 125, -125, 250}),
		[]string{"sequencer1", "adsr1"}, "",
	))

	steps := []adsr.ADSRStep{
		{LevelRatio: 1.0, DurationRatio: 0.05, Shape: 0.1},
		{LevelRatio: 0.3, DurationRatio: 0.05, Shape: -0.1},
		{DurationRatio: 0.1},
		{DurationRatio: 0.35, Shape: -0.1},
	}

	paramVarTri := env.AddModule(vartri.NewVarTri(0.2, 0.0, 0.5, env.Config, "vartri"))
	syncParam := env.AddModule(functor.NewFunctor(1, func(vec []float64) float64 { return vec[0]*3.0 + 1.1 }, env.Config, "adsr1"))
	adsrEnv1 := env.AddModule(adsr.NewADSRModule(steps, 1.0, env.Config, "adsr1"))
	mult1 := env.AddModule(functor.NewFunctor(2, functor.FunctorMult, env.Config, ""))
	filterParam := env.AddModule(functor.NewFunctor(1, func(vec []float64) float64 { return vec[0]*1200.0 + 40.0 }, env.Config, ""))
	osc1 := env.AddModule(phaseshaper.NewPhaseShaper(100.0, 0.0, softSyncTriangle(env.Config.SampleRate), 1, paramMapper, msgMapper, env.Config, "osc1"))
	allpass := env.AddModule(allpass.NewAllpass(375.0, 375.0, 0.4, env.Config, "allpass"))
	allpassAmp := env.AddModule(functor.NewFunctor(1, func(vec []float64) float64 { return vec[0] * 0.5 }, env.Config, ""))
	filter := env.AddModule(butterworth.NewButterworth(300.0, 0.4, env.Config, "filter"))
	// filter := env.AddModule(rbj.NewRBJFilter(rbjc.Lowpass, 300.0, 10.0, env.Config, "filter"))
	reverb := env.AddModule(freeverb.NewFreeVerb(env.Config, "freeverb"))

	reverb.(*freeverb.FreeVerb).SetDamp(0.1)
	reverb.(*freeverb.FreeVerb).SetDry(0.7)
	reverb.(*freeverb.FreeVerb).SetWet(0.2)
	reverb.(*freeverb.FreeVerb).SetRoomSize(0.8)
	reverb.(*freeverb.FreeVerb).SetWidth(0.8)

	muse.Connect(paramVarTri, 0, syncParam, 0)
	muse.Connect(paramVarTri, 0, filterParam, 0)
	muse.Connect(syncParam, 0, osc1, 1)
	muse.Connect(osc1, 0, mult1, 0)
	muse.Connect(adsrEnv1, 0, mult1, 1)
	muse.Connect(mult1, 0, filter, 0)
	muse.Connect(filterParam, 0, filter, 1)
	muse.Connect(filter, 0, allpass, 0)
	muse.Connect(allpass, 0, allpassAmp, 0)
	muse.Connect(filter, 0, reverb, 0)
	muse.Connect(allpassAmp, 0, reverb, 0)
	muse.Connect(reverb, 0, env, 0)
	muse.Connect(reverb, 1, env, 1)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/shaper.aiff", 24.0, sndfile.SF_FORMAT_AIFF)
}

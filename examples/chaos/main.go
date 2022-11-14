package main

import (
	"math/rand"

	"github.com/almerlucke/muse"
	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	"github.com/almerlucke/muse/components/interpolator"
	"github.com/almerlucke/muse/components/iterator"
	"github.com/almerlucke/muse/components/iterator/chaos/verhulst"
	"github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/controls/val"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/timer"
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/chorus"
	"github.com/almerlucke/muse/modules/filters/korg35"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/generator"
	"github.com/almerlucke/muse/modules/pan"
	"github.com/almerlucke/muse/modules/polyphony"
	"github.com/almerlucke/muse/modules/waveshaper"
	"github.com/almerlucke/muse/value"
	"github.com/almerlucke/muse/value/template"
	"github.com/mkb218/gosndfile/sndfile"
)

type ChaosVoice struct {
	*muse.BasePatch
	verhulst       *verhulst.Verhulst
	iter           *iterator.Iterator
	interpol       *interpolator.Interpolator
	ampEnvSteps    adsrc.StepProvider
	filterEnvSteps adsrc.StepProvider
	filter         *korg35.Korg35LPF
	ampEnv         *adsr.ADSR
	filterEnv      *adsr.ADSR
	genMod         *generator.Generator
	waveShape      *waveshaper.WaveShaper
	panner         *pan.StereoPan
}

func NewChaosVoice(config *muse.Configuration, ampEnvSteps adsrc.StepProvider, filterEnvSteps adsrc.StepProvider) *ChaosVoice {
	verhulst := verhulst.NewVerhulstWithFunc(3.6951, verhulst.Iter1a)
	iter := iterator.NewIterator([]float64{0.1231}, verhulst)
	interpol := interpolator.NewInterpolator(iter, interpolator.Cubic, 120)

	voice := &ChaosVoice{
		BasePatch:      muse.NewPatch(0, 2, config, ""),
		verhulst:       verhulst,
		iter:           iter,
		interpol:       interpol,
		ampEnvSteps:    ampEnvSteps,
		filterEnvSteps: filterEnvSteps,
		panner:         pan.NewStereoPan(0.5, config, ""),
		filter:         korg35.NewKorg35LPF(1500.0, 1.2, 1.0, config, "filter"),
		genMod:         generator.NewGenerator(interpol, nil, nil, config, ""),
		waveShape:      waveshaper.NewWaveShaper(waveshaping.NewChain(waveshaping.NewMirror(0.0, 1.0), waveshaping.NewBipolar()), 0, nil, nil, config, ""),
		ampEnv:         adsr.NewADSR(ampEnvSteps.GetSteps(), adsrc.Absolute, adsrc.Duration, 1.0, config, "ampEnv"),
		filterEnv:      adsr.NewADSR(filterEnvSteps.GetSteps(), adsrc.Absolute, adsrc.Duration, 1.0, config, "filterEnv"),
	}

	voice.SetSelf(voice)

	voice.AddModule(voice.ampEnv)
	voice.AddModule(voice.filterEnv)
	voice.AddModule(voice.genMod)
	voice.AddModule(voice.waveShape)
	voice.AddModule(voice.filter)
	voice.AddModule(voice.panner)

	filterScaler := voice.AddModule(functor.NewFunctor(1, func(v []float64) float64 {
		min := 50.0
		max := 12000.0
		if min > max {
			tmp := max
			max = min
			min = tmp
		}
		return v[0]*(max-min) + min
	}, config))

	ampVCA := voice.AddModule(functor.NewMult(2, config))

	voice.genMod.Connect(0, voice.waveShape, 0)
	voice.waveShape.Connect(0, voice.filter, 0)
	voice.filterEnv.Connect(0, filterScaler, 0)
	filterScaler.Connect(0, voice.filter, 1)
	voice.filter.Connect(0, ampVCA, 0)
	voice.ampEnv.Connect(0, ampVCA, 1)

	ampVCA.Connect(0, voice.panner, 0)

	voice.panner.Connect(0, voice, 0)
	voice.panner.Connect(1, voice, 1)

	return voice
}

func (v *ChaosVoice) IsActive() bool {
	return v.ampEnv.IsActive()
}

func (v *ChaosVoice) Note(duration float64, amplitude float64, msg any, config *muse.Configuration) {
	content := msg.(map[string]any)

	if numCycles, ok := content["numCycles"]; ok {
		v.interpol.SetNumCycles(int(numCycles.(float64)))
	}

	if chaos, ok := content["chaos"]; ok {
		v.verhulst.A = chaos.(float64)
	}

	if pan, ok := content["pan"]; ok {
		v.panner.SetPan(pan.(float64))
	}

	v.iter.SetValues([]float64{rand.Float64()})

	v.ampEnv.TriggerFull(duration, amplitude, v.ampEnvSteps.GetSteps(), adsrc.Absolute, adsrc.Duration)
	v.filterEnv.TriggerFull(duration, 1.0, v.filterEnvSteps.GetSteps(), adsrc.Absolute, adsrc.Duration)
}

func (v *ChaosVoice) NoteOn(amplitude float64, msg any, config *muse.Configuration) {
}

func (v *ChaosVoice) NoteOff() {
}

func randMinMax(min float64, max float64) float64 {
	return rand.Float64()*(max-min) + min
}

func main() {
	env := muse.NewEnvironment(2, 44100.0, 1024)

	ampEnv := adsrc.NewBasicStepProvider()
	ampEnv.Steps[0] = adsrc.Step{Level: 1.0, Duration: 450.0}
	ampEnv.Steps[1] = adsrc.Step{Level: 1.0, Duration: 0.0}
	ampEnv.Steps[3] = adsrc.Step{Duration: 3000.0}

	filterEnv := adsrc.NewBasicStepProvider()
	filterEnv.Steps[0] = adsrc.Step{Level: 1.0, Duration: 435.0}
	filterEnv.Steps[1] = adsrc.Step{Level: 1.0, Duration: 430.0}
	filterEnv.Steps[3] = adsrc.Step{Duration: 3000.0}

	numVoices := 20
	voices := make([]polyphony.Voice, numVoices)

	for i := 0; i < numVoices; i++ {
		voices[i] = NewChaosVoice(env.Config, ampEnv, filterEnv)
	}

	poly := env.AddModule(polyphony.NewPolyphony(2, voices, env.Config, "chaosSynth"))
	timer := env.AddControl(timer.NewControlTimer(500.0, env.Config, ""))
	randomizeTimer := val.NewVal[float64](value.NewFunction(func() float64 {
		return randMinMax(100, 2500.0)
	}), "")

	timer.CtrlConnect(0, randomizeTimer, 0)
	randomizeTimer.CtrlConnect(0, timer, 0)

	trigger := banger.NewControlTemplateGenerator(template.Template{
		"command": "trigger",
		"duration": value.NewFunction(func() any {
			return randMinMax(1600, 5000.0)
		}),
		"amplitude": value.NewFunction(func() any {
			return randMinMax(0.2, 0.6)
		}),
		"message": template.Template{
			"numCycles": value.NewFunction(func() any {
				return randMinMax(60, 160.0)
			}),
			"chaos": value.NewFunction(func() any {
				return randMinMax(0.1, 3.0)
			}),
			"pan": value.NewFunction(func() any {
				return randMinMax(0.0, 1.0)
			}),
		},
	}, "")

	timer.CtrlConnect(0, trigger, 0)
	trigger.CtrlConnect(0, poly, 0)

	chorus1 := env.AddModule(chorus.NewChorus(false, 20.0, 10.0, 0.3, 1.21, 0.3, nil, env.Config, ""))
	chorus2 := env.AddModule(chorus.NewChorus(false, 21.0, 11.0, 0.31, 1.21, 0.3, nil, env.Config, ""))

	poly.Connect(0, chorus1, 0)
	poly.Connect(1, chorus2, 0)

	chorus1.Connect(0, env, 0)
	chorus2.Connect(0, env, 1)

	// chorus2.Connect(0, env, 0)
	// chorus2.Connect(1, env, 1)

	// env.QuickPlayAudio()
	env.SynthesizeToFile("/Users/almerlucke/Desktop/chaos2.aiff", 360.0, 44100.0, true, sndfile.SF_FORMAT_AIFF)
}

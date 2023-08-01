package main

import (
	"math/rand"

	"github.com/almerlucke/muse"
	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	"github.com/almerlucke/muse/components/interpolator"
	"github.com/almerlucke/muse/components/iterator"
	"github.com/almerlucke/muse/components/iterator/chaos"
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
)

type ChaosVoice struct {
	*muse.BasePatch
	verhulst       *chaos.Verhulst
	iter           *iterator.Iterator
	interpol       *interpolator.Interpolator
	ampEnvSteps    adsrc.StepProvider
	filterEnvSteps adsrc.StepProvider
	filter         *korg35.Korg35LPF
	ampEnv         *adsr.ADSR
	filterEnv      *adsr.ADSR
	genMod         *generator.Generator
	waveShape      *waveshaper.WaveShaper
	panner         *pan.Pan
}

func NewChaosVoice(ampEnvSteps adsrc.StepProvider, filterEnvSteps adsrc.StepProvider) *ChaosVoice {
	verhulst := chaos.NewVerhulstWithFunc(3.6951, chaos.Iter1)
	iter := iterator.New([]float64{0.1231}, verhulst)
	interpol := interpolator.New(iter, interpolator.Linear, 1.0/120.0)

	voice := &ChaosVoice{
		BasePatch:      muse.NewPatch(0, 2),
		verhulst:       verhulst,
		iter:           iter,
		interpol:       interpol,
		ampEnvSteps:    ampEnvSteps,
		filterEnvSteps: filterEnvSteps,
		panner:         pan.New(0.5),
		filter:         korg35.New(1500.0, 1.6, 1.0),
		genMod:         generator.New(interpol, nil, nil),
		waveShape:      waveshaper.New(waveshaping.NewSeries(waveshaping.NewMirror(0.0, 1.0), waveshaping.NewBipolar()), 0, nil, nil),
		ampEnv:         adsr.New(ampEnvSteps.GetSteps(), adsrc.Absolute, adsrc.Duration, 1.0),
		filterEnv:      adsr.New(filterEnvSteps.GetSteps(), adsrc.Absolute, adsrc.Duration, 1.0),
	}

	voice.SetSelf(voice)

	voice.AddModule(voice.ampEnv)
	voice.AddModule(voice.filterEnv)
	voice.AddModule(voice.genMod)
	voice.AddModule(voice.waveShape)
	voice.AddModule(voice.filter)
	voice.AddModule(voice.panner)

	filterScaler := voice.AddModule(functor.NewBetween(50.0, 7000.0))

	ampVCA := voice.AddModule(functor.NewMult(2))

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
		v.interpol.SetDelta(1.0 / (numCycles.(float64)))
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
	root := muse.New(2)

	ampEnv := adsrc.NewBasicStepProvider()
	ampEnv.Steps[0] = adsrc.Step{Level: 1.0, Duration: 5.0}
	ampEnv.Steps[1] = adsrc.Step{Level: 0.2, Duration: 5.0}
	ampEnv.Steps[3] = adsrc.Step{Duration: 4000.0}

	filterEnv := adsrc.NewBasicStepProvider()
	filterEnv.Steps[0] = adsrc.Step{Level: 1.0, Duration: 5.0}
	filterEnv.Steps[1] = adsrc.Step{Level: 0.3, Duration: 5.0}
	filterEnv.Steps[3] = adsrc.Step{Duration: 4000.0}

	numVoices := 20
	voices := make([]polyphony.Voice, numVoices)

	for i := 0; i < numVoices; i++ {
		voices[i] = NewChaosVoice(ampEnv, filterEnv)
	}

	poly := polyphony.New(2, voices).Named("chaosSynth").Add(root)
	timer := timer.NewControlTimer(500.0).CtrlAdd(root)
	randomizeTimer := val.New[float64](value.NewFunction(func() float64 {
		return randMinMax(100, 5500.0)
	}))

	timer.CtrlConnect(0, randomizeTimer, 0)
	randomizeTimer.CtrlConnect(0, timer, 0)

	trigger := banger.NewControlTemplateGenerator(template.Template{
		"command": "trigger",
		"duration": value.NewFunction(func() any {
			return randMinMax(100, 1000.0)
		}),
		"amplitude": value.NewFunction(func() any {
			return randMinMax(0.6, 0.9)
		}),
		"message": template.Template{
			"numCycles": value.NewFunction(func() any {
				return randMinMax(20, 90.0)
			}),
			"chaos": value.NewFunction(func() any {
				return randMinMax(3.1, 4.0)
			}),
			"pan": value.NewFunction(func() any {
				return randMinMax(0.0, 1.0)
			}),
		},
	})

	timer.CtrlConnect(0, trigger, 0)
	trigger.CtrlConnect(0, poly, 0)

	chorus1 := root.AddModule(chorus.New(false, 30.0, 20.0, 0.4, 1.21, 0.4, nil))
	chorus2 := root.AddModule(chorus.New(false, 31.0, 21.0, 0.31, 1.31, 0.5, nil))

	poly.Connect(0, chorus1, 0)
	poly.Connect(1, chorus2, 0)

	chorus1.Connect(0, root, 0)
	chorus2.Connect(0, root, 1)

	// chorus2.Connect(0, env, 0)
	// chorus2.Connect(1, env, 1)

	root.RenderAudio()
	// env.SynthesizeToFile("/Users/almerlucke/Desktop/chaosPing1.aiff", 360.0, 44100.0, true, sndfile.SF_FORMAT_AIFF)
}

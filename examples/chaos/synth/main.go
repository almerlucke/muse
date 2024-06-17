package main

import (
	"github.com/almerlucke/muse/modules/effects/chorus"
	"math/rand"

	adsrc "github.com/almerlucke/genny/float/envelopes/adsr"
	"github.com/almerlucke/genny/float/interp"
	"github.com/almerlucke/genny/float/iter"
	"github.com/almerlucke/genny/float/iter/updaters/chaos"
	"github.com/almerlucke/genny/float/shape/shapers/linear"
	"github.com/almerlucke/genny/float/shape/shapers/mirror"
	"github.com/almerlucke/genny/float/shape/shapers/series"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/template"

	"github.com/almerlucke/muse/controls/gen"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/timer"
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/filters/korg35"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/generator"
	"github.com/almerlucke/muse/modules/pan"
	"github.com/almerlucke/muse/modules/polyphony"
	"github.com/almerlucke/muse/modules/waveshaper"
)

type ChaosVoice struct {
	*muse.BasePatch
	verhulst         *chaos.Verhulst
	iter             *iter.Iterator
	interpol         *interp.Interpolator
	ampEnvSetting    *adsrc.Setting
	filterEnvSetting *adsrc.Setting
	filter           *korg35.LPF
	ampEnv           *adsr.ADSR
	filterEnv        *adsr.ADSR
	genMod           *generator.Generator
	waveShape        *waveshaper.WaveShaper
	panner           *pan.Pan
}

func NewChaosVoice(ampEnvSetting *adsrc.Setting, filterEnvSetting *adsrc.Setting) *ChaosVoice {
	verhulst := chaos.NewVerhulstWithFunc(3.6951, chaos.Iter1a)
	it := iter.New([]float64{0.1231}, verhulst)
	interpol := interp.New(it, interp.Cubic, 1.0/120.0)

	voice := &ChaosVoice{
		BasePatch:        muse.NewPatch(0, 2),
		verhulst:         verhulst,
		iter:             it,
		interpol:         interpol,
		ampEnvSetting:    ampEnvSetting,
		filterEnvSetting: filterEnvSetting,
		panner:           pan.New(0.5),
		filter:           korg35.New(1500.0, 1.2, 1.0),
		genMod:           generator.New(interpol, nil, nil),
		waveShape:        waveshaper.New(series.New(mirror.New(0.0, 1.0), linear.NewBipolar()), 0, nil, nil),
		ampEnv:           adsr.New(ampEnvSetting, adsrc.Duration, 1.0),
		filterEnv:        adsr.New(filterEnvSetting, adsrc.Duration, 1.0),
	}

	voice.SetSelf(voice)

	voice.AddModule(voice.ampEnv)
	voice.AddModule(voice.filterEnv)
	voice.AddModule(voice.genMod)
	voice.AddModule(voice.waveShape)
	voice.AddModule(voice.filter)
	voice.AddModule(voice.panner)

	filterScaler := voice.AddModule(functor.NewBetween(50.0, 12000.0))

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
		v.interpol.SetDelta(1.0 / numCycles.(float64))
	}

	if ch, ok := content["chaos"]; ok {
		v.verhulst.A = ch.(float64)
	}

	if p, ok := content["pan"]; ok {
		v.panner.SetPan(p.(float64))
	}

	v.iter.SetValues([]float64{rand.Float64()})

	v.ampEnv.TriggerWithDuration(duration, amplitude)
	v.filterEnv.TriggerWithDuration(duration, 1.0)
}

func (v *ChaosVoice) NoteOn(amplitude float64, msg any, config *muse.Configuration) {
}

func (v *ChaosVoice) NoteOff() {
}

func (v *ChaosVoice) Clear() {
	v.ampEnv.Clear()
	v.filterEnv.Clear()
}

func randMinMax(min float64, max float64) float64 {
	return rand.Float64()*(max-min) + min
}

func main() {
	root := muse.New(2)

	ampEnvSetting := adsrc.NewSetting(1.0, 450.0, 1.0, 0.0, 0.0, 3000.0)
	filterEnvSetting := adsrc.NewSetting(1.0, 435.0, 1.0, 430.0, 0.0, 3000.0)

	numVoices := 20
	voices := make([]polyphony.Voice, numVoices)

	for i := 0; i < numVoices; i++ {
		voices[i] = NewChaosVoice(ampEnvSetting, filterEnvSetting)
	}

	poly := polyphony.New(2, voices).Named("chaosSynth").AddTo(root)
	tim := root.AddControl(timer.NewControl(500.0, nil))
	randomizeTimer := gen.New[float64](function.New(func() float64 {
		return randMinMax(100, 2500.0)
	}), false)

	tim.CtrlConnect(0, randomizeTimer, 0)
	randomizeTimer.CtrlConnect(0, tim, 0)

	trigger := banger.NewControlTemplate(template.Template{
		"command": "trigger",
		"duration": function.New(func() float64 {
			return randMinMax(1600, 5000.0)
		}),
		"amplitude": function.New(func() float64 {
			return randMinMax(0.1, 0.3)
		}),
		"message": template.Template{
			"numCycles": function.New(func() float64 {
				return randMinMax(60, 160.0)
			}),
			"chaos": function.New(func() float64 {
				return randMinMax(0.1, 3.0)
			}),
			"pan": function.New(func() float64 {
				return randMinMax(0.0, 1.0)
			}),
		},
	})

	tim.CtrlConnect(0, trigger, 0)
	trigger.CtrlConnect(0, poly, 0)

	chorus1 := root.AddModule(chorus.New(0.34, 0.4, 0.4, 0.2, 1.0, 0.5, nil))
	chorus2 := root.AddModule(chorus.New(0.43, 0.4, 0.41, 0.21, 1.0, 0.5, nil))

	poly.Connect(0, chorus1, 0)
	poly.Connect(1, chorus2, 0)

	chorus1.Connect(0, root, 0)
	chorus2.Connect(0, root, 1)

	_ = root.RenderAudio()
	// root.RenderToSoundFile("/Users/almerlucke/Desktop/chaos2.aiff", 20.0, 44100.0, true)
}

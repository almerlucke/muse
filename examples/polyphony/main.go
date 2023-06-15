package main

import (
	"github.com/almerlucke/muse"
	"github.com/mkb218/gosndfile/sndfile"

	// Components
	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	shaping "github.com/almerlucke/muse/components/waveshaping"

	// Value
	"github.com/almerlucke/muse/value"

	// Template
	"github.com/almerlucke/muse/value/template"

	// Messengers
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"

	// Modules
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/allpass"
	"github.com/almerlucke/muse/modules/filters/moog"
	"github.com/almerlucke/muse/modules/freeverb"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/phasor"
	"github.com/almerlucke/muse/modules/polyphony"
	"github.com/almerlucke/muse/modules/vartri"
	"github.com/almerlucke/muse/modules/waveshaper"
)

type TestVoice struct {
	*muse.BasePatch
	adsrEnv *adsr.ADSR
	phasor  *phasor.Phasor
	Shaper  *waveshaper.WaveShaper
	Filter  *moog.Moog
}

func paramMapper(param int, value any, shaper shaping.Shaper) {
	if param == 0 {
		shaper.(*shaping.Chain).SetSuperSawM1(value.(float64))
	}
}

func NewTestVoice(config *muse.Configuration) *TestVoice {
	testVoice := &TestVoice{
		BasePatch: muse.NewPatch(0, 1, config, ""),
	}

	steps := []adsrc.Step{
		{Level: 1.0, Duration: 20, Shape: 0.0},
		{Level: 0.3, Duration: 20, Shape: 0.0},
		{Duration: 20},
		{Duration: 350, Shape: 0.1},
	}

	adsrEnv := testVoice.AddModule(adsr.NewADSR(steps, adsrc.Absolute, adsrc.Duration, 1.0, config, "adsr"))
	multiplier := testVoice.AddModule(functor.NewFunctor(2, functor.FunctorMult, config))
	osc := testVoice.AddModule(phasor.NewPhasor(140.0, 0.0, config, "osc"))
	shape := testVoice.AddModule(waveshaper.NewWaveShaper(shaping.NewSuperSaw(), 1, paramMapper, nil, config, "shaper"))
	filter := testVoice.AddModule(moog.NewMoog(1700.0, 0.48, 1.0, config, "filter"))

	osc.Connect(0, shape, 0)
	shape.Connect(0, multiplier, 0)
	adsrEnv.Connect(0, multiplier, 1)
	multiplier.Connect(0, filter, 0)
	filter.Connect(0, testVoice, 0)

	testVoice.adsrEnv = adsrEnv.(*adsr.ADSR)
	testVoice.Shaper = shape.(*waveshaper.WaveShaper)
	testVoice.Filter = filter.(*moog.Moog)
	testVoice.phasor = osc.(*phasor.Phasor)

	return testVoice
}

func (tv *TestVoice) IsActive() bool {
	return tv.adsrEnv.IsActive()
}

func (tv *TestVoice) Note(duration float64, amplitude float64, message any, config *muse.Configuration) {
	msg := message.(map[string]any)

	tv.adsrEnv.TriggerWithDuration(duration, amplitude)
	tv.phasor.ReceiveMessage(msg["osc"])
}

func (tv *TestVoice) NoteOn(amplitude float64, message any, config *muse.Configuration) {
	// STUB
}

func (tv *TestVoice) NoteOff() {
	// STUB
}

func main() {
	env := muse.NewEnvironment(2, 3*44100, 512)

	env.AddMessenger(banger.NewTemplateGenerator([]string{"polyphony"}, template.Template{
		"command":   "trigger",
		"duration":  value.NewSequence([]any{75.0, 125.0, 75.0, 250.0, 75.0, 250.0, 75.0, 75.0, 75.0, 250.0, 125.0}),
		"amplitude": value.NewSequence([]any{1.0, 0.6, 1.0, 0.5, 0.5, 1.0, 0.3, 1.0, 0.7}),
		"message": template.Template{
			"osc": template.Template{
				"frequency": value.NewSequence([]any{50.0, 50.0, 25.0, 50.0, 150.0, 25.0, 150.0, 100.0, 100.0, 200.0, 50.0}),
				"phase":     0.0,
			},
		},
	}, "prototype"))

	bpm := 120

	env.AddMessenger(stepper.NewStepper(
		swing.New(bpm, 4, value.NewSequence([]*swing.Step{
			{}, {Shuffle: 0.2}, {Skip: true}, {Shuffle: 0.4, ShuffleRand: 0.2}, {}, {Shuffle: 0.3}, {Shuffle: 0.1}, {SkipChance: 0.3},
		})),
		[]string{"prototype"}, "",
	))

	milliPerBeat := 60000.0 / float64(bpm) / 4.0

	paramVarTri1 := env.AddModule(vartri.NewVarTri(0.265, 0.0, 0.5, env.Config, "vartri1"))
	paramVarTri2 := env.AddModule(vartri.NewVarTri(0.325, 0.0, 0.5, env.Config, "vartri2"))
	superSawDrive := env.AddModule(functor.NewFunctor(1, func(vec []float64) float64 { return vec[0]*0.84 + 0.15 }, env.Config))
	filterCutOff := env.AddModule(functor.NewFunctor(1, func(vec []float64) float64 { return vec[0]*3200.0 + 40.0 }, env.Config))

	voices := []polyphony.Voice{}
	for i := 0; i < 20; i++ {
		voice := NewTestVoice(env.Config)
		voices = append(voices, voice)
		// connect lfo to voices
		superSawDrive.Connect(0, voice.Shaper, 1)
		filterCutOff.Connect(0, voice.Filter, 1)
	}

	poly := env.AddModule(polyphony.NewPolyphony(1, voices, env.Config, "polyphony"))
	allpass := env.AddModule(allpass.NewAllpass(milliPerBeat*3, milliPerBeat*3, 0.4, env.Config, "allpass"))
	allpassAmp := env.AddModule(functor.NewFunctor(1, func(vec []float64) float64 { return vec[0] * 0.5 }, env.Config))
	reverb := env.AddModule(freeverb.NewFreeVerb(env.Config, "freeverb"))

	reverb.(*freeverb.FreeVerb).SetDamp(0.1)
	reverb.(*freeverb.FreeVerb).SetDry(0.7)
	reverb.(*freeverb.FreeVerb).SetWet(0.1)
	reverb.(*freeverb.FreeVerb).SetRoomSize(0.8)
	reverb.(*freeverb.FreeVerb).SetWidth(0.8)

	// connect external voice inputs to voice player so the external modules
	// are always synthesized even if no voice is active at the moment
	superSawDrive.Connect(0, poly, 0)
	filterCutOff.Connect(0, poly, 0)

	paramVarTri1.Connect(0, superSawDrive, 0)
	paramVarTri2.Connect(0, filterCutOff, 0)
	poly.Connect(0, allpass, 0)
	allpass.Connect(0, allpassAmp, 0)
	poly.Connect(0, reverb, 0)
	allpassAmp.Connect(0, reverb, 1)
	reverb.Connect(0, env, 0)
	reverb.Connect(1, env, 1)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/voices.aiff", 24.0, env.Config.SampleRate, true, sndfile.SF_FORMAT_AIFF)
}

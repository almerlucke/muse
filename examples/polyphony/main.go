package main

import (
	"github.com/almerlucke/muse"

	// Components
	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	shaping "github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/modules"

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
		shaper.(*shaping.Serial).SetSuperSawM1(value.(float64))
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

	adsrEnv := adsr.NewADSR(steps, adsrc.Absolute, adsrc.Duration, 1.0, config).Add(testVoice)
	osc := phasor.NewPhasor(140.0, 0.0, config, "osc").Add(testVoice)
	shape := waveshaper.NewWaveShaper(shaping.NewSuperSaw(1.5, 0.25, 0.88), 1, paramMapper, nil, config, "shaper").
		Add(testVoice).In(osc)
	mult := modules.Mult(shape, adsrEnv).Add(testVoice)
	filter := moog.NewMoog(1700.0, 0.48, 1.0, config, "filter").Add(testVoice).In(mult)

	testVoice.In(filter)

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

	superSawDrive := modules.Scale(vartri.NewVarTri(0.265, 0.0, 0.5, env.Config).Add(env), 0, 0.84, 0.15).Add(env)
	filterCutOff := modules.Scale(vartri.NewVarTri(0.325, 0.0, 0.5, env.Config).Add(env), 0, 3200.0, 40.0).Add(env)

	voices := []polyphony.Voice{}
	for i := 0; i < 20; i++ {
		voice := NewTestVoice(env.Config)
		voices = append(voices, voice)
		// connect lfo to voices
		superSawDrive.Connect(0, voice.Shaper, 1)
		filterCutOff.Connect(0, voice.Filter, 1)
	}

	// connect external voice inputs to voice player so the external modules
	// are always synthesized even if no voice is active at the moment
	poly := polyphony.NewPolyphony(1, voices, env.Config).Named("polyphony").Add(env).In(superSawDrive, filterCutOff, 0, 0)
	allpass := allpass.NewAllpass(milliPerBeat*3, milliPerBeat*3, 0.4, env.Config).Add(env).In(poly)
	allpassAmp := modules.Scale(allpass, 0, 0.5, 0.0).Add(env)
	reverb := freeverb.NewFreeVerb(env.Config).Add(env).In(poly, allpassAmp).(*freeverb.FreeVerb)

	reverb.SetDamp(0.1)
	reverb.SetDry(0.7)
	reverb.SetWet(0.1)
	reverb.SetRoomSize(0.8)
	reverb.SetWidth(0.8)

	env.In(reverb, reverb, 1)

	env.QuickPlayAudio()

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/voices.aiff", 24.0, env.Config.SampleRate, true, sndfile.SF_FORMAT_AIFF)
}

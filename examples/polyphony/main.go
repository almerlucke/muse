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
		shaper.(*shaping.SuperSaw).SetM1(value.(float64))
	}
}

func NewTestVoice() *TestVoice {
	testVoice := &TestVoice{
		BasePatch: muse.NewPatch(0, 1),
	}

	steps := []adsrc.Step{
		{Level: 1.0, Duration: 20, Shape: 0.0},
		{Level: 0.3, Duration: 20, Shape: 0.0},
		{Duration: 20},
		{Duration: 350, Shape: 0.1},
	}

	adsrEnv := adsr.New(steps, adsrc.Absolute, adsrc.Duration, 1.0).Add(testVoice)
	osc := phasor.New(140.0, 0.0).Add(testVoice)
	shape := waveshaper.New(shaping.NewSuperSaw(1.5, 0.25, 0.88), 1, paramMapper, nil).
		Add(testVoice).In(osc)
	mult := modules.Mult(shape, adsrEnv).Add(testVoice)
	filter := moog.New(1700.0, 0.48, 1.0).Add(testVoice).In(mult)

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
	root := muse.New(2)

	root.AddMessenger(banger.NewTemplateGenerator([]string{"polyphony"}, template.Template{
		"command":   "trigger",
		"duration":  value.NewSequence([]any{75.0, 125.0, 75.0, 250.0, 75.0, 250.0, 75.0, 75.0, 75.0, 250.0, 125.0}),
		"amplitude": value.NewSequence([]any{1.0, 0.6, 1.0, 0.5, 0.5, 1.0, 0.3, 1.0, 0.7}),
		"message": template.Template{
			"osc": template.Template{
				"frequency": value.NewSequence([]any{50.0, 50.0, 25.0, 50.0, 150.0, 25.0, 150.0, 100.0, 100.0, 200.0, 50.0}),
				"phase":     0.0,
			},
		},
	}).MsgrNamed("prototype"))

	bpm := 120

	root.AddMessenger(stepper.NewStepper(
		swing.New(bpm, 4, value.NewSequence([]*swing.Step{
			{}, {Shuffle: 0.2}, {Skip: true}, {Shuffle: 0.4, ShuffleRand: 0.2}, {}, {Shuffle: 0.3}, {Shuffle: 0.1}, {SkipChance: 0.3},
		})),
		[]string{"prototype"},
	))

	milliPerBeat := 60000.0 / float64(bpm) / 4.0

	superSawDrive := modules.Scale(vartri.New(0.265, 0.0, 0.5).Add(root), 0, 0.84, 0.15).Add(root)
	filterCutOff := modules.Scale(vartri.New(0.325, 0.0, 0.5).Add(root), 0, 3200.0, 40.0).Add(root)

	voices := []polyphony.Voice{}
	for i := 0; i < 20; i++ {
		voice := NewTestVoice()
		voices = append(voices, voice)
		// connect lfo to voices
		superSawDrive.Connect(0, voice.Shaper, 1)
		filterCutOff.Connect(0, voice.Filter, 1)
	}

	// connect external voice inputs to voice player so the external modules
	// are always synthesized even if no voice is active at the moment
	poly := polyphony.New(1, voices).Named("polyphony").Add(root).In(superSawDrive, filterCutOff, 0, 0)
	allpass := allpass.New(milliPerBeat*3, milliPerBeat*3, 0.4).Add(root).In(poly)
	allpassAmp := modules.Scale(allpass, 0, 0.5, 0.0).Add(root)
	reverb := freeverb.New().Add(root).In(poly, allpassAmp).(*freeverb.FreeVerb)

	reverb.SetDamp(0.1)
	reverb.SetDry(0.7)
	reverb.SetWet(0.1)
	reverb.SetRoomSize(0.8)
	reverb.SetWidth(0.8)

	root.In(reverb, reverb, 1)

	root.RenderAudio()

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/voices.aiff", 24.0, env.Config.SampleRate, true, sndfile.SF_FORMAT_AIFF)
}

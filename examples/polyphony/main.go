package main

import (
	adsrc "github.com/almerlucke/genny/float/envelopes/adsr"
	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/float/shape/shapers/emulations/supersaw"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/template"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/modules"
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/allpass"
	"github.com/almerlucke/muse/modules/effects/freeverb"
	"github.com/almerlucke/muse/modules/filters/moog"
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

func paramMapper(param int, value any, shaper shape.Shaper) {
	if param == 0 {
		shaper.(*supersaw.SuperSaw).SetM1(value.(float64))
	}
}

func NewTestVoice() *TestVoice {
	testVoice := &TestVoice{
		BasePatch: muse.NewPatch(0, 1),
	}

	setting := adsrc.NewSetting(1.0, 20.0, 0.3, 20.0, 20.0, 350.0)

	adsrEnv := adsr.New(setting, adsrc.Duration, 1.0).AddTo(testVoice)
	osc := phasor.New(140.0, 0.0).AddTo(testVoice)
	shaper := waveshaper.New(supersaw.New(1.5, 0.25, 0.88), 1, paramMapper, nil).
		AddTo(testVoice).In(osc)
	mult := modules.Mult(shaper, adsrEnv).AddTo(testVoice)
	filter := moog.New(1700.0, 0.48, 1.0).AddTo(testVoice).In(mult)

	testVoice.In(filter)

	testVoice.adsrEnv = adsrEnv.(*adsr.ADSR)
	testVoice.Shaper = shaper.(*waveshaper.WaveShaper)
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

func (tv *TestVoice) Clear() {
	tv.adsrEnv.Clear()
}

func main() {
	root := muse.New(2)

	root.AddMessenger(banger.NewTemplateBang([]string{"polyphony"}, template.Template{
		"command":   "trigger",
		"duration":  sequence.NewLoop(75.0, 125.0, 75.0, 250.0, 75.0, 250.0, 75.0, 75.0, 75.0, 250.0, 125.0),
		"amplitude": sequence.NewLoop(1.0, 0.6, 1.0, 0.5, 0.5, 1.0, 0.3, 1.0, 0.7),
		"message": template.Template{
			"osc": template.Template{
				"frequency": sequence.NewLoop(50.0, 50.0, 25.0, 50.0, 150.0, 25.0, 150.0, 100.0, 100.0, 200.0, 50.0),
				"phase":     0.0,
			},
		},
	}).MsgrNamed("prototype"))

	bpm := 120

	root.AddMessenger(stepper.NewStepper(
		swing.New(bpm, 4, sequence.NewLoop([]*swing.Step{
			{}, {Shuffle: 0.2}, {Skip: true}, {Shuffle: 0.4, ShuffleRand: 0.2}, {}, {Shuffle: 0.3}, {Shuffle: 0.1}, {SkipChance: 0.3},
		}...)),
		[]string{"prototype"},
	))

	milliPerBeat := 60000.0 / float64(bpm) / 4.0

	superSawDrive := modules.Scale(vartri.New(0.265, 0.0, 0.5).AddTo(root), 0, 0.84, 0.15).AddTo(root)
	filterCutOff := modules.Scale(vartri.New(0.325, 0.0, 0.5).AddTo(root), 0, 3200.0, 40.0).AddTo(root)

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
	poly := polyphony.New(1, voices).Named("polyphony").AddTo(root).In(superSawDrive, filterCutOff, 0, 0)
	allp := allpass.New(milliPerBeat*3, milliPerBeat*3, 0.4).AddTo(root).In(poly)
	allpassAmp := modules.Scale(allp, 0, 0.5, 0.0).AddTo(root)
	reverb := freeverb.New().AddTo(root).In(poly, allpassAmp).(*freeverb.FreeVerb)

	reverb.SetDamp(0.1)
	reverb.SetDry(0.7)
	reverb.SetWet(0.1)
	reverb.SetRoomSize(0.8)
	reverb.SetWidth(0.8)

	root.In(reverb, reverb, 1)

	_ = root.RenderAudio()

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/voices.aiff", 24.0, env.Config.SampleRate, true, sndfile.SF_FORMAT_AIFF)
}

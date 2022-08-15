package main

import (
	"log"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/gordonklaus/portaudio"

	"github.com/almerlucke/muse"

	// Components
	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	shapingc "github.com/almerlucke/muse/components/shaping"

	// Values
	"github.com/almerlucke/muse/values"

	// Messengers
	"github.com/almerlucke/muse/messengers/generators/prototype"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"

	// Modules
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/phasor"
	"github.com/almerlucke/muse/modules/shaper"
)

type TestVoice struct {
	*muse.BasePatch
	adsrEnv *adsr.ADSR
	phasor  *phasor.Phasor
	Shaper  *shaper.Shaper
}

func NewTestVoice(config *muse.Configuration) *TestVoice {
	testVoice := &TestVoice{
		BasePatch: muse.NewPatch(0, 1, config, ""),
	}

	steps := []adsrc.Step{
		{Level: 1.0, Duration: 10, Shape: 0.0},
		{Level: 0.2, Duration: 10, Shape: 0.0},
		{Duration: 30},
		{Duration: 550, Shape: 0.1},
	}

	adsrEnv := testVoice.AddModule(adsr.NewADSR(steps, adsrc.Absolute, adsrc.Duration, 1.0, config, "adsr"))
	multiplier := testVoice.AddModule(functor.NewFunctor(2, functor.FunctorMult, config, ""))
	osc := testVoice.AddModule(phasor.NewPhasor(140.0, 0.0, config, "osc"))
	shape := testVoice.AddModule(shaper.NewShaper(shapingc.NewSineTable(512), 0, nil, nil, config, "shaper"))

	muse.Connect(osc, 0, shape, 0)
	muse.Connect(shape, 0, multiplier, 0)
	muse.Connect(adsrEnv, 0, multiplier, 1)
	muse.Connect(multiplier, 0, testVoice, 0)

	testVoice.adsrEnv = adsrEnv.(*adsr.ADSR)
	testVoice.Shaper = shape.(*shaper.Shaper)
	testVoice.phasor = osc.(*phasor.Phasor)

	return testVoice
}

func (tv *TestVoice) IsActive() bool {
	return tv.adsrEnv.IsActive()
}

func (tv *TestVoice) Activate(duration float64, amplitude float64, message any, config *muse.Configuration) {
	msg := message.(map[string]any)

	tv.adsrEnv.TriggerWithDuration(duration, amplitude)
	tv.phasor.ReceiveMessage(msg["osc"])
}

func main() {
	env := muse.NewEnvironment(1, 44100, 512)

	env.AddMessenger(prototype.NewPrototypeGenerator([]string{"voicePlayer"}, values.MapPrototype{
		"duration":  values.NewSequence([]any{125.0, 125.0, 125.0, 250.0, 125.0, 250.0, 125.0, 125.0, 125.0, 250.0, 125.0}, true),
		"amplitude": values.NewConst[any](1.0),
		"message": values.MapPrototype{
			"osc": values.MapPrototype{
				"frequency": values.NewSequence([]any{
					440.0, 220.0, 110.0, 220.0, 660.0, 440.0, 880.0, 330.0, 880.0, 1320.0, 110.0,
					440.0, 220.0, 110.0, 220.0, 660.0, 440.0, 880.0, 330.0, 880.0, 1100.0, 770.0, 550.0}, true),
				"phase": values.NewConst[any](0.0),
			},
		},
	}, "prototype1"))

	env.AddMessenger(prototype.NewPrototypeGenerator([]string{"voicePlayer"}, values.MapPrototype{
		"duration":  values.NewSequence([]any{250.0, 250.0, 375.0, 375.0, 375.0, 250.0}, true),
		"amplitude": values.NewConst[any](0.3),
		"message": values.MapPrototype{
			"osc": values.MapPrototype{
				"frequency": values.NewSequence([]any{
					110.0, 220.0, 330.0, 110.0, 220.0, 220.0}, true),
				"phase": values.NewConst[any](0.0),
			},
		},
	}, "prototype2"))

	bpm := 100.0

	env.AddMessenger(stepper.NewStepper(
		swing.New(bpm, 4.0, values.NewSequence([]*swing.Step{
			{}, {Shuffle: 0.2}, {Skip: true}, {Shuffle: 0.4, ShuffleRand: 0.2}, {}, {Shuffle: 0.3}, {Shuffle: 0.1}, {SkipFactor: 0.3},
		}, true)),
		[]string{"prototype1", "prototype2"}, "",
	))

	voices := []muse.Voice{}
	for i := 0; i < 20; i++ {
		voice := NewTestVoice(env.Config)
		voices = append(voices, voice)
	}

	voicePlayer := env.AddModule(muse.NewVoicePlayer(1, voices, env.Config, "voicePlayer"))

	// connect external voice inputs to voice player so the external modules
	// are always synthesized even if no voice is active at the moment
	muse.Connect(voicePlayer, 0, env, 0)

	portaudio.Initialize()
	defer portaudio.Terminate()

	stream, err := env.PortaudioStream()
	if err != nil {
		log.Fatalf("error opening portaudio stream, %v", err)
	}

	stream.Start()

	defer stream.Close()

	a := app.New()

	w := a.NewWindow("Hello")

	hello := widget.NewLabel("Hello Fyne!")
	w.SetContent(container.NewVBox(
		hello,
		widget.NewButton("Hi!", func() {
			hello.SetText("Welcome :)")
		}),
	))

	w.ShowAndRun()
}

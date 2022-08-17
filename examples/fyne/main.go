package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
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
	"github.com/almerlucke/muse/modules/allpass"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/phasor"
	"github.com/almerlucke/muse/modules/shaper"
)

type FixedWidthLayout struct {
	Width float32
}

func NewFixedWidthLayout(w float32) *FixedWidthLayout {
	return &FixedWidthLayout{Width: w}
}

func (fwl *FixedWidthLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	for _, object := range objects {
		object.Resize(fyne.NewSize(fwl.Width, objects[0].MinSize().Height))
		object.Move(fyne.NewPos(0, 0))
	}
}

func (fwl *FixedWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	maxH := float32(0.0)

	for _, object := range objects {
		childSize := object.MinSize()
		if childSize.Height > maxH {
			maxH = childSize.Height
		}
	}

	return fyne.NewSize(fwl.Width, maxH)
}

type StepProvider interface {
	Steps() []adsrc.Step
}

type StepSetter struct {
	TheSteps []adsrc.Step
}

func (s *StepSetter) Steps() []adsrc.Step {
	return s.TheSteps
}

type TestVoice struct {
	*muse.BasePatch
	adsrEnv      *adsr.ADSR
	phasor       *phasor.Phasor
	Shaper       *shaper.Shaper
	stepProvider StepProvider
}

func NewTestVoice(config *muse.Configuration, stepProvider StepProvider) *TestVoice {
	testVoice := &TestVoice{
		BasePatch:    muse.NewPatch(0, 1, config, ""),
		stepProvider: stepProvider,
	}

	adsrEnv := testVoice.AddModule(adsr.NewADSR(stepProvider.Steps(), adsrc.Absolute, adsrc.Duration, 1.0, config, "adsr"))
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

	tv.adsrEnv.TriggerFull(duration, amplitude, tv.stepProvider.Steps(), adsrc.Absolute, adsrc.Duration)
	tv.phasor.ReceiveMessage(msg["osc"])
}

func main() {
	env := muse.NewEnvironment(2, 44100, 512)

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
					110.0, 220.0, 660.0, 110.0, 220.0, 440.0, 1540.0, 110.0, 220.0, 660.0, 550.0, 220.0, 440.0, 380.0,
					110.0, 220.0, 660.0, 110.0, 220.0, 440.0, 1110.0, 110.0, 220.0, 660.0, 550.0, 220.0, 440.0, 380.0}, true),
				"phase": values.NewConst[any](0.0),
			},
		},
	}, "prototype2"))

	bpm := 80.0

	stepProvider := &StepSetter{
		TheSteps: []adsrc.Step{
			{Level: 1.0, Duration: 5, Shape: 0.0},
			{Level: 0.3, Duration: 5, Shape: 0.0},
			{Duration: 20},
			{Duration: 5, Shape: 0.0},
		},
	}

	env.AddMessenger(stepper.NewStepper(
		swing.New(bpm, 4.0, values.NewSequence([]*swing.Step{
			{}, {Shuffle: 0.2}, {Skip: true}, {Shuffle: 0.4, ShuffleRand: 0.2}, {}, {Shuffle: 0.3}, {Shuffle: 0.1}, {SkipFactor: 0.3},
		}, true)),
		[]string{"prototype1", "prototype2"}, "",
	))

	voices := []muse.Voice{}
	for i := 0; i < 20; i++ {
		voice := NewTestVoice(env.Config, stepProvider)
		voices = append(voices, voice)
	}

	milliPerBeat := 60000.0 / bpm

	voicePlayer := env.AddModule(muse.NewVoicePlayer(1, voices, env.Config, "voicePlayer"))
	allpass := env.AddModule(allpass.NewAllpass(milliPerBeat*1.5, milliPerBeat*1.5, 0.1, env.Config, "allpass"))

	muse.Connect(voicePlayer, 0, allpass, 0)
	muse.Connect(voicePlayer, 0, env, 0)
	muse.Connect(allpass, 0, env, 1)

	portaudio.Initialize()
	defer portaudio.Terminate()

	stream, err := env.PortaudioStream()
	if err != nil {
		log.Fatalf("error opening portaudio stream, %v", err)
	}

	stream.Start()

	defer stream.Close()

	a := app.New()

	a.Settings().SetTheme(theme.LightTheme())

	w := a.NewWindow("Muse")

	w.Resize(fyne.Size{
		Width:  300,
		Height: 200,
	})

	attackMSLabel := widget.NewLabel("5.0")
	attackMSLabel.Alignment = fyne.TextAlignTrailing
	attackMSSlider := widget.NewSlider(5.0, 500.0)
	attackMSSlider.Value = 5.0
	attackMSSlider.OnChanged = func(f float64) {
		attackMSLabel.SetText(fmt.Sprintf("%.2f", f))
		stepProvider.TheSteps[0].Duration = f
	}

	attackAmpLabel := widget.NewLabel("1.0")
	attackAmpLabel.Alignment = fyne.TextAlignTrailing
	attackAmpSlider := widget.NewSlider(0.0, 1.0)
	attackAmpSlider.Step = 0.01
	attackAmpSlider.Value = 1.0
	attackAmpSlider.OnChanged = func(f float64) {
		attackAmpLabel.SetText(fmt.Sprintf("%.2f", f))
		stepProvider.TheSteps[0].Level = f
	}

	attackShapeLabel := widget.NewLabel("0.0")
	attackShapeLabel.Alignment = fyne.TextAlignTrailing
	attackShapeSlider := widget.NewSlider(-1.0, 1.0)
	attackShapeSlider.Step = 0.01
	attackShapeSlider.Value = 0.0
	attackShapeSlider.OnChanged = func(f float64) {
		attackShapeLabel.SetText(fmt.Sprintf("%.2f", f))
		stepProvider.TheSteps[0].Shape = f
	}

	decayMSLabel := widget.NewLabel("5.0")
	decayMSLabel.Alignment = fyne.TextAlignTrailing
	decayMSSlider := widget.NewSlider(5.0, 500.0)
	decayMSSlider.Value = 5.0
	decayMSSlider.OnChanged = func(f float64) {
		decayMSLabel.SetText(fmt.Sprintf("%.2f", f))
		stepProvider.TheSteps[1].Duration = f
	}

	decayAmpLabel := widget.NewLabel("0.3")
	decayAmpLabel.Alignment = fyne.TextAlignTrailing
	decayAmpSlider := widget.NewSlider(0.0, 1.0)
	decayAmpSlider.Step = 0.01
	decayAmpSlider.Value = 0.3
	decayAmpSlider.OnChanged = func(f float64) {
		decayAmpLabel.SetText(fmt.Sprintf("%.2f", f))
		stepProvider.TheSteps[1].Level = f
	}

	decayShapeLabel := widget.NewLabel("0.0")
	decayShapeLabel.Alignment = fyne.TextAlignTrailing
	decayShapeSlider := widget.NewSlider(-1.0, 1.0)
	decayShapeSlider.Step = 0.01
	decayShapeSlider.Value = 0.0
	decayShapeSlider.OnChanged = func(f float64) {
		decayShapeLabel.SetText(fmt.Sprintf("%.2f", f))
		stepProvider.TheSteps[1].Shape = f
	}

	releaseMSLabel := widget.NewLabel("5.0")
	releaseMSLabel.Alignment = fyne.TextAlignTrailing
	releaseMSSlider := widget.NewSlider(5.0, 500.0)
	releaseMSSlider.Value = 5.0
	releaseMSSlider.OnChanged = func(f float64) {
		releaseMSLabel.SetText(fmt.Sprintf("%.2f", f))
		stepProvider.TheSteps[3].Duration = f
	}

	releaseShapeLabel := widget.NewLabel("0.0")
	releaseShapeLabel.Alignment = fyne.TextAlignTrailing
	releaseShapeSlider := widget.NewSlider(-1.0, 1.0)
	releaseShapeSlider.Step = 0.01
	releaseShapeSlider.Value = 0.0
	releaseShapeSlider.OnChanged = func(f float64) {
		releaseShapeLabel.SetText(fmt.Sprintf("%.2f", f))
		stepProvider.TheSteps[3].Shape = f
	}

	w.SetContent(
		widget.NewCard("ADSR Envelope", "",
			container.NewHBox(
				container.New(NewFixedWidthLayout(250),
					widget.NewCard("Attack", "", container.NewVBox(
						widget.NewLabel("attack duration (ms)"),
						container.NewBorder(nil, nil, nil, container.New(NewFixedWidthLayout(80), attackMSLabel), attackMSSlider),
						widget.NewLabel("attack amplitude (0.0 - 1.0)"),
						container.NewBorder(nil, nil, nil, container.New(NewFixedWidthLayout(80), attackAmpLabel), attackAmpSlider),
						widget.NewLabel("attack shape (-1.0 - 1.0)"),
						container.NewBorder(nil, nil, nil, container.New(NewFixedWidthLayout(80), attackShapeLabel), attackShapeSlider),
					)),
				),
				container.New(NewFixedWidthLayout(250),
					widget.NewCard("Decay", "", container.NewVBox(
						widget.NewLabel("decay duration (ms)"),
						container.NewBorder(nil, nil, nil, container.New(NewFixedWidthLayout(80), decayMSLabel), decayMSSlider),
						widget.NewLabel("decay amplitude (0.0 - 1.0)"),
						container.NewBorder(nil, nil, nil, container.New(NewFixedWidthLayout(80), decayAmpLabel), decayAmpSlider),
						widget.NewLabel("decay shape (-1.0 - 1.0)"),
						container.NewBorder(nil, nil, nil, container.New(NewFixedWidthLayout(80), decayShapeLabel), decayShapeSlider),
					)),
				),
				container.New(NewFixedWidthLayout(250),
					widget.NewCard("Release", "", container.NewVBox(
						widget.NewLabel("release duration (ms)"),
						container.NewBorder(nil, nil, nil, container.New(NewFixedWidthLayout(80), releaseMSLabel), releaseMSSlider),
						widget.NewLabel("release shape (-1.0 - 1.0)"),
						container.NewBorder(nil, nil, nil, container.New(NewFixedWidthLayout(80), releaseShapeLabel), releaseShapeSlider),
					)),
				),
			),
		),
	)

	w.ShowAndRun()
}

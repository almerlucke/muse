package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/gordonklaus/portaudio"

	"github.com/almerlucke/muse"

	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	shapingc "github.com/almerlucke/muse/components/shaping"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/utils"
	"github.com/almerlucke/muse/values"

	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/allpass"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/phasor"
	"github.com/almerlucke/muse/modules/polyphony"
	"github.com/almerlucke/muse/modules/shaper"
)

type ADSRStepProvider interface {
	ADSRSteps() []adsrc.Step
}

type TestVoice struct {
	*muse.BasePatch
	adsrEnv      *adsr.ADSR
	phasor       *phasor.Phasor
	superSaw     *shapingc.Chain
	stepProvider ADSRStepProvider
}

func NewTestVoice(config *muse.Configuration, stepProvider ADSRStepProvider) *TestVoice {
	testVoice := &TestVoice{
		BasePatch:    muse.NewPatch(0, 1, config, ""),
		stepProvider: stepProvider,
		superSaw:     shapingc.NewSuperSaw(),
	}

	adsrEnv := testVoice.AddModule(adsr.NewADSR(stepProvider.ADSRSteps(), adsrc.Absolute, adsrc.Duration, 1.0, config, "adsr"))
	multiplier := testVoice.AddModule(functor.NewFunctor(2, functor.FunctorMult, config, ""))
	osc := testVoice.AddModule(phasor.NewPhasor(140.0, 0.0, config, "osc"))
	shape := testVoice.AddModule(shaper.NewShaper(testVoice.superSaw, 0, nil, nil, config, "shaper"))

	muse.Connect(osc, 0, shape, 0)
	muse.Connect(shape, 0, multiplier, 0)
	muse.Connect(adsrEnv, 0, multiplier, 1)
	muse.Connect(multiplier, 0, testVoice, 0)

	testVoice.adsrEnv = adsrEnv.(*adsr.ADSR)
	testVoice.phasor = osc.(*phasor.Phasor)

	return testVoice
}

func (tv *TestVoice) IsActive() bool {
	return tv.adsrEnv.IsActive()
}

func (tv *TestVoice) Activate(duration float64, amplitude float64, message any, config *muse.Configuration) {
	// STUB
}

func (tv *TestVoice) NoteOn(amplitude float64, message any, config *muse.Configuration) {
	msg := message.(map[string]any)

	tv.adsrEnv.TriggerFull(0, amplitude, tv.stepProvider.ADSRSteps(), adsrc.Absolute, adsrc.NoteOff)
	tv.phasor.ReceiveMessage(msg["osc"])
}

func (tv *TestVoice) NoteOff() {
	tv.adsrEnv.Release()
}

func (tv *TestVoice) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	tv.superSaw.SetSuperSawM1(content["superSawM1"].(float64))

	return nil
}

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

type ADSRControl struct {
	steps                        []adsrc.Step
	attackDurationLabelBinding   binding.String
	attackDurationSliderBinding  binding.Float
	attackLevelLabelBinding      binding.String
	attackLevelSliderBinding     binding.Float
	attackShapeLabelBinding      binding.String
	attackShapeSliderBinding     binding.Float
	decayDurationLabelBinding    binding.String
	decayDurationSliderBinding   binding.Float
	decayLevelLabelBinding       binding.String
	decayLevelSliderBinding      binding.Float
	decayShapeLabelBinding       binding.String
	decayShapeSliderBinding      binding.Float
	releaseDurationLabelBinding  binding.String
	releaseDurationSliderBinding binding.Float
	releaseShapeLabelBinding     binding.String
	releaseShapeSliderBinding    binding.Float
}

func (ctrl *ADSRControl) UI() fyne.CanvasObject {
	// Attack Duration
	ctrl.attackDurationLabelBinding = binding.NewString()
	ctrl.attackDurationLabelBinding.Set(fmt.Sprintf("%.2f", ctrl.steps[0].Duration))
	attackDurationLabel := widget.NewLabelWithData(ctrl.attackDurationLabelBinding)
	attackDurationLabel.Alignment = fyne.TextAlignTrailing

	ctrl.attackDurationSliderBinding = binding.NewFloat()
	ctrl.attackDurationSliderBinding.Set(5.0)
	ctrl.attackDurationSliderBinding.AddListener(binding.NewDataListener(func() {
		v, err := ctrl.attackDurationSliderBinding.Get()
		if err == nil {
			ctrl.attackDurationLabelBinding.Set(fmt.Sprintf("%.2f", v))
			ctrl.steps[0].Duration = v
		}
	}))
	attackDurationSlider := widget.NewSliderWithData(5.0, 500.0, ctrl.attackDurationSliderBinding)

	// Attack Level
	ctrl.attackLevelLabelBinding = binding.NewString()
	ctrl.attackLevelLabelBinding.Set(fmt.Sprintf("%.2f", ctrl.steps[0].Level))
	attackLevelLabel := widget.NewLabelWithData(ctrl.attackLevelLabelBinding)
	attackLevelLabel.Alignment = fyne.TextAlignTrailing

	ctrl.attackLevelSliderBinding = binding.NewFloat()
	ctrl.attackLevelSliderBinding.Set(1.0)
	ctrl.attackLevelSliderBinding.AddListener(binding.NewDataListener(func() {
		v, err := ctrl.attackLevelSliderBinding.Get()
		if err == nil {
			ctrl.attackLevelLabelBinding.Set(fmt.Sprintf("%.2f", v))
			ctrl.steps[0].Level = v
		}
	}))
	attackLevelSlider := widget.NewSliderWithData(0.0, 1.0, ctrl.attackLevelSliderBinding)
	attackLevelSlider.Step = 0.01

	// Attack Shape
	ctrl.attackShapeLabelBinding = binding.NewString()
	ctrl.attackShapeLabelBinding.Set(fmt.Sprintf("%.2f", ctrl.steps[0].Level))
	attackShapeLabel := widget.NewLabelWithData(ctrl.attackShapeLabelBinding)
	attackShapeLabel.Alignment = fyne.TextAlignTrailing

	ctrl.attackShapeSliderBinding = binding.NewFloat()
	ctrl.attackShapeSliderBinding.Set(0.0)
	ctrl.attackShapeSliderBinding.AddListener(binding.NewDataListener(func() {
		v, err := ctrl.attackShapeSliderBinding.Get()
		if err == nil {
			ctrl.attackShapeLabelBinding.Set(fmt.Sprintf("%.2f", v))
			ctrl.steps[0].Shape = v
		}
	}))
	attackShapeSlider := widget.NewSliderWithData(-1.0, 1.0, ctrl.attackShapeSliderBinding)
	attackShapeSlider.Step = 0.01

	// Decay Duration
	ctrl.decayDurationLabelBinding = binding.NewString()
	ctrl.decayDurationLabelBinding.Set(fmt.Sprintf("%.2f", ctrl.steps[1].Duration))
	decayDurationLabel := widget.NewLabelWithData(ctrl.decayDurationLabelBinding)
	decayDurationLabel.Alignment = fyne.TextAlignTrailing

	ctrl.decayDurationSliderBinding = binding.NewFloat()
	ctrl.decayDurationSliderBinding.Set(5.0)
	ctrl.decayDurationSliderBinding.AddListener(binding.NewDataListener(func() {
		v, err := ctrl.decayDurationSliderBinding.Get()
		if err == nil {
			ctrl.decayDurationLabelBinding.Set(fmt.Sprintf("%.2f", v))
			ctrl.steps[1].Duration = v
		}
	}))
	decayDurationSlider := widget.NewSliderWithData(5.0, 500.0, ctrl.decayDurationSliderBinding)

	// Decay Level
	ctrl.decayLevelLabelBinding = binding.NewString()
	ctrl.decayLevelLabelBinding.Set(fmt.Sprintf("%.2f", ctrl.steps[1].Level))
	decayLevelLabel := widget.NewLabelWithData(ctrl.decayLevelLabelBinding)
	decayLevelLabel.Alignment = fyne.TextAlignTrailing

	ctrl.decayLevelSliderBinding = binding.NewFloat()
	ctrl.decayLevelSliderBinding.Set(0.3)
	ctrl.decayLevelSliderBinding.AddListener(binding.NewDataListener(func() {
		v, err := ctrl.decayLevelSliderBinding.Get()
		if err == nil {
			ctrl.decayLevelLabelBinding.Set(fmt.Sprintf("%.2f", v))
			ctrl.steps[1].Level = v
		}
	}))
	decayLevelSlider := widget.NewSliderWithData(0.0, 1.0, ctrl.decayLevelSliderBinding)
	decayLevelSlider.Step = 0.01

	// Decay Shape
	ctrl.decayShapeLabelBinding = binding.NewString()
	ctrl.decayShapeLabelBinding.Set(fmt.Sprintf("%.2f", ctrl.steps[1].Level))
	decayShapeLabel := widget.NewLabelWithData(ctrl.decayShapeLabelBinding)
	decayShapeLabel.Alignment = fyne.TextAlignTrailing

	ctrl.decayShapeSliderBinding = binding.NewFloat()
	ctrl.decayShapeSliderBinding.Set(0.0)
	ctrl.decayShapeSliderBinding.AddListener(binding.NewDataListener(func() {
		v, err := ctrl.decayShapeSliderBinding.Get()
		if err == nil {
			ctrl.decayShapeLabelBinding.Set(fmt.Sprintf("%.2f", v))
			ctrl.steps[1].Shape = v
		}
	}))
	decayShapeSlider := widget.NewSliderWithData(-1.0, 1.0, ctrl.decayShapeSliderBinding)
	decayShapeSlider.Step = 0.01

	// Release Duration
	ctrl.releaseDurationLabelBinding = binding.NewString()
	ctrl.releaseDurationLabelBinding.Set(fmt.Sprintf("%.2f", ctrl.steps[3].Duration))
	releaseDurationLabel := widget.NewLabelWithData(ctrl.releaseDurationLabelBinding)
	releaseDurationLabel.Alignment = fyne.TextAlignTrailing

	ctrl.releaseDurationSliderBinding = binding.NewFloat()
	ctrl.releaseDurationSliderBinding.Set(50.0)
	ctrl.releaseDurationSliderBinding.AddListener(binding.NewDataListener(func() {
		v, err := ctrl.releaseDurationSliderBinding.Get()
		if err == nil {
			ctrl.releaseDurationLabelBinding.Set(fmt.Sprintf("%.2f", v))
			ctrl.steps[3].Duration = v
		}
	}))
	releaseDurationSlider := widget.NewSliderWithData(50.0, 2500.0, ctrl.releaseDurationSliderBinding)
	releaseDurationSlider.Step = 10.0

	// Release Shape
	ctrl.releaseShapeLabelBinding = binding.NewString()
	ctrl.releaseShapeLabelBinding.Set(fmt.Sprintf("%.2f", ctrl.steps[3].Level))
	releaseShapeLabel := widget.NewLabelWithData(ctrl.releaseShapeLabelBinding)
	releaseShapeLabel.Alignment = fyne.TextAlignTrailing

	ctrl.releaseShapeSliderBinding = binding.NewFloat()
	ctrl.releaseShapeSliderBinding.Set(0.0)
	ctrl.releaseShapeSliderBinding.AddListener(binding.NewDataListener(func() {
		v, err := ctrl.releaseShapeSliderBinding.Get()
		if err == nil {
			ctrl.releaseShapeLabelBinding.Set(fmt.Sprintf("%.2f", v))
			ctrl.steps[3].Shape = v
		}
	}))
	releaseShapeSlider := widget.NewSliderWithData(-1.0, 1.0, ctrl.releaseShapeSliderBinding)
	releaseShapeSlider.Step = 0.01

	return widget.NewCard("ADSR Envelope", "",
		container.NewHBox(
			container.New(NewFixedWidthLayout(250),
				widget.NewCard("Attack", "", container.NewVBox(
					widget.NewLabel("attack duration (ms)"),
					container.NewBorder(nil, nil, nil, container.New(NewFixedWidthLayout(80), attackDurationLabel), attackDurationSlider),
					widget.NewLabel("attack amplitude (0.0 - 1.0)"),
					container.NewBorder(nil, nil, nil, container.New(NewFixedWidthLayout(80), attackLevelLabel), attackLevelSlider),
					widget.NewLabel("attack shape (-1.0 - 1.0)"),
					container.NewBorder(nil, nil, nil, container.New(NewFixedWidthLayout(80), attackShapeLabel), attackShapeSlider),
				)),
			),
			container.New(NewFixedWidthLayout(250),
				widget.NewCard("Decay", "", container.NewVBox(
					widget.NewLabel("decay duration (ms)"),
					container.NewBorder(nil, nil, nil, container.New(NewFixedWidthLayout(80), decayDurationLabel), decayDurationSlider),
					widget.NewLabel("decay amplitude (0.0 - 1.0)"),
					container.NewBorder(nil, nil, nil, container.New(NewFixedWidthLayout(80), decayLevelLabel), decayLevelSlider),
					widget.NewLabel("decay shape (-1.0 - 1.0)"),
					container.NewBorder(nil, nil, nil, container.New(NewFixedWidthLayout(80), decayShapeLabel), decayShapeSlider),
				)),
			),
			container.New(NewFixedWidthLayout(250),
				widget.NewCard("Release", "", container.NewVBox(
					widget.NewLabel("release duration (ms)"),
					container.NewBorder(nil, nil, nil, container.New(NewFixedWidthLayout(80), releaseDurationLabel), releaseDurationSlider),
					widget.NewLabel("release shape (-1.0 - 1.0)"),
					container.NewBorder(nil, nil, nil, container.New(NewFixedWidthLayout(80), releaseShapeLabel), releaseShapeSlider),
				)),
			),
		),
	)
}

func (ctrl *ADSRControl) ADSRSteps() []adsrc.Step {
	return ctrl.steps
}

func (ctrl *ADSRControl) GetState() map[string]any {
	stepStates := make([]map[string]any, len(ctrl.steps))

	for index, step := range ctrl.steps {
		stepStates[index] = step.GetState()
	}

	return map[string]any{"steps": stepStates}
}

func (ctrl *ADSRControl) SetState(state map[string]any) {
	stepStates := state["steps"].([]any)

	for index, stepState := range stepStates {
		ctrl.steps[index].SetState(stepState.(map[string]any))
	}

	ctrl.attackDurationSliderBinding.Set(ctrl.steps[0].Duration)
	ctrl.attackLevelSliderBinding.Set(ctrl.steps[0].Level)
	ctrl.attackShapeSliderBinding.Set(ctrl.steps[0].Shape)
	ctrl.decayDurationSliderBinding.Set(ctrl.steps[1].Duration)
	ctrl.decayLevelSliderBinding.Set(ctrl.steps[1].Level)
	ctrl.decayShapeSliderBinding.Set(ctrl.steps[1].Shape)
	ctrl.releaseDurationSliderBinding.Set(ctrl.steps[3].Duration)
	ctrl.releaseShapeSliderBinding.Set(ctrl.steps[3].Shape)
}

func NewADSRControl() *ADSRControl {
	control := &ADSRControl{
		steps: []adsrc.Step{
			{Level: 1.0, Duration: 5, Shape: 0.0},
			{Level: 0.3, Duration: 5, Shape: 0.0},
			{Duration: 20},
			{Duration: 5, Shape: 0.0},
		},
	}

	return control
}

func main() {
	env := muse.NewEnvironment(2, 44100, 512)

	adsrControl := NewADSRControl()

	voices := []polyphony.Voice{}
	for i := 0; i < 20; i++ {
		voice := NewTestVoice(env.Config, adsrControl)
		voices = append(voices, voice)
	}

	poly := env.AddModule(polyphony.NewPolyphony(1, voices, env.Config, "polyphony"))
	allpass := env.AddModule(allpass.NewAllpass(50, 50, 0.3, env.Config, "allpass"))

	target := lfo.NewLFOTarget("polyphony", "superSawM1", values.MapPrototype{
		"command":    "voice",
		"superSawM1": values.NewPlaceholder("superSawM1"),
	})

	env.AddMessenger(lfo.NewLFO(0, 0.13, 0.15, 0.85, shapingc.NewNormalizedSineTable(512), []*lfo.LFOTarget{target}, env.Config, "lfo"))

	muse.Connect(poly, 0, allpass, 0)
	muse.Connect(poly, 0, env, 0)
	muse.Connect(allpass, 0, env, 1)

	portaudio.Initialize()
	defer portaudio.Terminate()

	stream, err := env.PortaudioStream()
	if err != nil {
		log.Fatalf("error opening portaudio stream, %v", err)
	}

	defer stream.Close()

	a := app.New()

	a.Settings().SetTheme(theme.LightTheme())

	w := a.NewWindow("Muse")

	w.Resize(fyne.Size{
		Width:  700,
		Height: 400,
	})

	keyMap := map[string]float64{}

	keyMap["`"] = utils.Mtof(46)
	keyMap["Z"] = utils.Mtof(47)
	keyMap["X"] = utils.Mtof(48)
	keyMap["C"] = utils.Mtof(49)
	keyMap["V"] = utils.Mtof(50)
	keyMap["B"] = utils.Mtof(51)
	keyMap["N"] = utils.Mtof(52)
	keyMap["M"] = utils.Mtof(53)
	keyMap[","] = utils.Mtof(54)
	keyMap["."] = utils.Mtof(55)
	keyMap["/"] = utils.Mtof(56)

	keyMap["A"] = utils.Mtof(57)
	keyMap["S"] = utils.Mtof(58)
	keyMap["D"] = utils.Mtof(59)
	keyMap["F"] = utils.Mtof(60)
	keyMap["G"] = utils.Mtof(61)
	keyMap["H"] = utils.Mtof(62)
	keyMap["J"] = utils.Mtof(63)
	keyMap["K"] = utils.Mtof(64)
	keyMap["L"] = utils.Mtof(65)
	keyMap[";"] = utils.Mtof(66)
	keyMap["'"] = utils.Mtof(67)
	keyMap["\\"] = utils.Mtof(68)

	keyMap["Q"] = utils.Mtof(69)
	keyMap["W"] = utils.Mtof(70)
	keyMap["E"] = utils.Mtof(71)
	keyMap["R"] = utils.Mtof(72)
	keyMap["T"] = utils.Mtof(73)
	keyMap["Y"] = utils.Mtof(74)
	keyMap["U"] = utils.Mtof(75)
	keyMap["I"] = utils.Mtof(76)
	keyMap["O"] = utils.Mtof(77)
	keyMap["P"] = utils.Mtof(78)
	keyMap["["] = utils.Mtof(79)
	keyMap["]"] = utils.Mtof(80)

	keyMap["1"] = utils.Mtof(81)
	keyMap["2"] = utils.Mtof(82)
	keyMap["3"] = utils.Mtof(83)
	keyMap["4"] = utils.Mtof(84)
	keyMap["5"] = utils.Mtof(85)
	keyMap["6"] = utils.Mtof(86)
	keyMap["7"] = utils.Mtof(87)
	keyMap["8"] = utils.Mtof(88)
	keyMap["9"] = utils.Mtof(89)
	keyMap["0"] = utils.Mtof(90)
	keyMap["-"] = utils.Mtof(91)
	keyMap["="] = utils.Mtof(92)

	if deskCanvas, ok := w.Canvas().(desktop.Canvas); ok {
		deskCanvas.SetOnKeyDown(func(k *fyne.KeyEvent) {
			if f, ok := keyMap[string(k.Name)]; ok {
				log.Printf("key down: %v", k.Name)
				poly.ReceiveMessage(map[string]any{
					"command":   "trigger",
					"noteOn":    string(k.Name),
					"amplitude": 1.0,
					"message": map[string]any{
						"osc": map[string]any{
							"frequency": f,
						},
					},
				})
			}
		})

		deskCanvas.SetOnKeyUp(func(k *fyne.KeyEvent) {
			if _, ok := keyMap[string(k.Name)]; ok {
				log.Printf("key up: %v", k.Name)
				poly.ReceiveMessage(map[string]any{
					"command": "trigger",
					"noteOff": string(k.Name),
				})
			}
		})
	}

	w.SetContent(
		container.NewVBox(
			container.NewHBox(
				widget.NewButton("Start", func() {
					stream.Start()
				}),
				widget.NewButton("Stop", func() {
					stream.Stop()
				}),
			),
			adsrControl.UI(),
		),
	)

	w.ShowAndRun()
}

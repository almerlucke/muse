package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/gordonklaus/portaudio"

	"github.com/almerlucke/muse"

	// Components
	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	shapingc "github.com/almerlucke/muse/components/shaping"
	"github.com/almerlucke/muse/ui"

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

type ADSRStepProvider interface {
	ADSRSteps() []adsrc.Step
}

type TestVoice struct {
	*muse.BasePatch
	adsrEnv      *adsr.ADSR
	phasor       *phasor.Phasor
	Shaper       *shaper.Shaper
	stepProvider ADSRStepProvider
}

func NewTestVoice(config *muse.Configuration, stepProvider ADSRStepProvider) *TestVoice {
	testVoice := &TestVoice{
		BasePatch:    muse.NewPatch(0, 1, config, ""),
		stepProvider: stepProvider,
	}

	adsrEnv := testVoice.AddModule(adsr.NewADSR(stepProvider.ADSRSteps(), adsrc.Absolute, adsrc.Duration, 1.0, config, "adsr"))
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

	tv.adsrEnv.TriggerFull(duration, amplitude, tv.stepProvider.ADSRSteps(), adsrc.Absolute, adsrc.Duration)
	tv.phasor.ReceiveMessage(msg["osc"])
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
	ctrl.releaseDurationSliderBinding.Set(5.0)
	ctrl.releaseDurationSliderBinding.AddListener(binding.NewDataListener(func() {
		v, err := ctrl.releaseDurationSliderBinding.Get()
		if err == nil {
			ctrl.releaseDurationLabelBinding.Set(fmt.Sprintf("%.2f", v))
			ctrl.steps[3].Duration = v
		}
	}))
	releaseDurationSlider := widget.NewSliderWithData(5.0, 500.0, ctrl.releaseDurationSliderBinding)

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

func parseFloatWithBounds(s string, min float64, max float64) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	if f < min {
		f = min
	}
	if f > max {
		f = max
	}
	return f
}

type SwingStepControl struct {
	step               *swing.Step
	skipBinding        binding.Bool
	shuffleBinding     binding.String
	shuffleRandBinding binding.String
	skipFactorBinding  binding.String
	currentBinding     binding.Bool
	index              int
	isCurrent          bool
}

func NewSwingStepControl(step *swing.Step, index int, isCurrent bool) *SwingStepControl {
	return &SwingStepControl{
		step:      step,
		index:     index,
		isCurrent: isCurrent,
	}
}

func (ssc *SwingStepControl) SetCurrent(c bool) {
	ssc.isCurrent = c
	ssc.currentBinding.Set(c)
}

func (ssc *SwingStepControl) UI() fyne.CanvasObject {
	ssc.skipBinding = binding.NewBool()
	ssc.skipBinding.Set(!ssc.step.Skip)
	ssc.skipBinding.AddListener(binding.NewDataListener(func() {
		v, err := ssc.skipBinding.Get()
		if err == nil {
			ssc.step.Skip = !v
		}
	}))
	skipCheck := widget.NewCheckWithData("", ssc.skipBinding)

	ssc.shuffleBinding = binding.NewString()
	ssc.shuffleBinding.Set(fmt.Sprintf("%.2f", ssc.step.Shuffle))
	ssc.shuffleBinding.AddListener(ui.NewDelayedListener(1*time.Second, func() {
		v, err := ssc.shuffleBinding.Get()
		if err == nil {
			ssc.step.Shuffle = parseFloatWithBounds(v, 0, 1)
		}
	}))
	shuffleEntry := widget.NewEntryWithData(ssc.shuffleBinding)
	shuffleEntry.OnSubmitted = func(v string) {
		ssc.step.Shuffle = parseFloatWithBounds(v, 0, 1)
	}
	shuffleEntry.Validator = nil

	ssc.shuffleRandBinding = binding.NewString()
	ssc.shuffleRandBinding.Set(fmt.Sprintf("%.2f", ssc.step.ShuffleRand))
	ssc.shuffleRandBinding.AddListener(ui.NewDelayedListener(1*time.Second, func() {
		v, err := ssc.shuffleRandBinding.Get()
		if err == nil {
			ssc.step.ShuffleRand = parseFloatWithBounds(v, 0, 1)
		}
	}))
	shuffleRandEntry := widget.NewEntryWithData(ssc.shuffleRandBinding)
	shuffleRandEntry.Validator = nil

	ssc.skipFactorBinding = binding.NewString()
	ssc.skipFactorBinding.Set(fmt.Sprintf("%.2f", ssc.step.SkipFactor))
	ssc.skipFactorBinding.AddListener(ui.NewDelayedListener(1*time.Second, func() {
		v, err := ssc.skipFactorBinding.Get()
		if err == nil {
			ssc.step.SkipFactor = parseFloatWithBounds(v, 0, 1)
		}
	}))
	skipFactorEntry := widget.NewEntryWithData(ssc.skipFactorBinding)
	skipFactorEntry.Validator = nil

	ssc.currentBinding = binding.NewBool()
	ssc.currentBinding.Set(ssc.isCurrent)
	currentCheck := widget.NewCheckWithData("", ssc.currentBinding)
	currentCheck.Disable()

	return widget.NewForm(
		widget.NewFormItem(fmt.Sprintf("%d", ssc.index), skipCheck),
		widget.NewFormItem("Swing", shuffleEntry),
		widget.NewFormItem("Rand", shuffleRandEntry),
		widget.NewFormItem("Skip", skipFactorEntry),
		widget.NewFormItem("", currentCheck),
	)
}

type SwingControlBank struct {
	Steps []*swing.Step
	N     int
}

func NewSwingControlBank() *SwingControlBank {
	b := &SwingControlBank{
		Steps: make([]*swing.Step, 64),
		N:     8,
	}

	for i := 0; i < 64; i++ {
		b.Steps[i] = &swing.Step{}
	}

	b.Steps[1].Shuffle = 0.2
	b.Steps[2].Skip = true
	b.Steps[3].Shuffle = 0.4
	b.Steps[3].ShuffleRand = 0.2
	b.Steps[5].Shuffle = 0.3
	b.Steps[6].Shuffle = 0.1
	b.Steps[7].SkipFactor = 0.3

	return b
}

type SwingControl struct {
	stepSequence        *values.Sequence[*swing.Step]
	stepControls        []*SwingStepControl
	steps               []*swing.Step
	bpm                 *values.Const[float64]
	noteDivision        *values.Const[float64]
	n                   int
	prevStepIndex       int
	nBinding            binding.String
	bpmBinding          binding.String
	noteDivisionBinding binding.String
}

func (sc *SwingControl) Listen(state map[string]any) {
	if sc.stepControls != nil {
		sc.stepControls[sc.prevStepIndex].SetCurrent(false)
		i := state["steps"].(map[string]any)["index"].(int)
		sc.stepControls[i].SetCurrent(true)
		sc.prevStepIndex = i
	}
}

func (sc *SwingControl) Sequence() *values.Sequence[*swing.Step] {
	return sc.stepSequence
}

func (sc *SwingControl) BPM() *values.Const[float64] {
	return sc.bpm
}

func (sc *SwingControl) NoteDivision() *values.Const[float64] {
	return sc.noteDivision
}

func (sc *SwingControl) UI() fyne.CanvasObject {
	stepCanvasObjects := []fyne.CanvasObject{}
	sc.stepControls = make([]*SwingStepControl, 64)

	for i := 0; i < 64; i++ {
		sc.stepControls[i] = NewSwingStepControl(sc.steps[i], i, i == 0)

		stepCanvasObjects = append(stepCanvasObjects, sc.stepControls[i].UI())

		if (i+1)%8 == 0 {
			stepCanvasObjects = append(stepCanvasObjects, widget.NewSeparator())
		}
	}

	sc.nBinding = binding.NewString()
	sc.nBinding.Set("8")
	sc.nBinding.AddListener(ui.NewDelayedListener(1*time.Second, func() {
		v, err := sc.nBinding.Get()
		if err == nil {
			n, _ := strconv.ParseInt(v, 10, 64)
			if n > 0 && n < 65 {
				sc.n = int(n)
				sc.stepSequence.Set(sc.steps[:sc.n])
			}
		}
	}))

	nEntry := widget.NewEntryWithData(sc.nBinding)
	nEntry.Validator = nil

	sc.bpmBinding = binding.NewString()
	sc.bpmBinding.Set("80")
	sc.bpmBinding.AddListener(ui.NewDelayedListener(1*time.Second, func() {
		v, err := sc.bpmBinding.Get()
		if err == nil {
			bpm, _ := strconv.ParseInt(v, 10, 64)
			if bpm > 0 && bpm < 400 {
				sc.bpm.Set(float64(bpm))
			}
		}
	}))

	bpmEntry := widget.NewEntryWithData(sc.bpmBinding)
	bpmEntry.Validator = nil

	sc.noteDivisionBinding = binding.NewString()
	sc.noteDivisionBinding.Set("4")
	sc.noteDivisionBinding.AddListener(ui.NewDelayedListener(1*time.Second, func() {
		v, err := sc.noteDivisionBinding.Get()
		if err == nil {
			noteDivision, _ := strconv.ParseInt(v, 10, 64)
			if noteDivision > 0 && noteDivision < 257 {
				sc.noteDivision.Set(float64(noteDivision))
			}
		}
	}))

	noteDivisionEntry := widget.NewEntryWithData(sc.noteDivisionBinding)
	noteDivisionEntry.Validator = nil

	radioGroup := widget.NewRadioGroup([]string{"1", "2", "3", "4", "5", "6", "7", "8"}, func(option string) {

	})
	radioGroup.Horizontal = true
	radioGroup.Selected = "1"

	return widget.NewCard("Rhythm", "",
		container.NewVBox(
			container.NewHBox(
				widget.NewCard("", "",
					container.NewHBox(
						widget.NewLabel("BPM"),
						bpmEntry,
						widget.NewLabel("Div"),
						noteDivisionEntry,
					),
				),
				layout.NewSpacer(),
				widget.NewCard("", "",
					container.NewHBox(
						widget.NewLabel("Steps"),
						nEntry,
						widget.NewButton("-", func() {
							if sc.n > 1 {
								sc.nBinding.Set(fmt.Sprintf("%d", sc.n-1))
							}
						}),
						widget.NewButton("+", func() {
							if sc.n < 64 {
								sc.nBinding.Set(fmt.Sprintf("%d", sc.n+1))
							}
						}),
					),
				),
				widget.NewCard("", "",
					container.NewHBox(
						widget.NewLabel("Bank"),
						radioGroup,
					),
				),
			),
			widget.NewCard("", "",
				container.NewHScroll(
					container.NewHBox(
						stepCanvasObjects...,
					),
				),
			),
		))
}

func NewSwingControl(bpm float64, noteDivision float64) *SwingControl {
	steps := make([]*swing.Step, 64)

	for i := 0; i < 64; i++ {
		steps[i] = &swing.Step{}
	}

	steps[1].Shuffle = 0.2
	steps[2].Skip = true
	steps[3].Shuffle = 0.4
	steps[3].ShuffleRand = 0.2
	steps[5].Shuffle = 0.3
	steps[6].Shuffle = 0.1
	steps[7].SkipFactor = 0.3

	return &SwingControl{
		steps:        steps,
		bpm:          values.NewConst(bpm),
		noteDivision: values.NewConst(noteDivision),
		n:            8,
		stepSequence: values.NewSequence(steps[:8], true),
	}
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
					110.0, 220.0, 660.0, 110.0, 220.0, 440.0, 1540.0, 110.0, 220.0, 660.0, 550.0, 220.0, 440.0, 380.0,
					110.0, 220.0, 660.0, 110.0, 220.0, 440.0, 1110.0, 110.0, 220.0, 660.0, 550.0, 220.0, 440.0, 380.0}, true),
				"phase": values.NewConst[any](0.0),
			},
		},
	}, "prototype2"))

	adsrControl := NewADSRControl()
	swingControl := NewSwingControl(80.0, 4.0)

	/*
		values.NewSequence([]*swing.Step{
				{}, {Shuffle: 0.2}, {Skip: true}, {Shuffle: 0.4, ShuffleRand: 0.2}, {}, {Shuffle: 0.3}, {Shuffle: 0.1}, {SkipFactor: 0.3},
			}, true)
	*/

	env.AddMessenger(stepper.NewStepperWithListener(
		swing.New(swingControl.BPM(), swingControl.NoteDivision(), swingControl.Sequence()),
		[]string{"prototype1", "prototype2"},
		swingControl,
		"",
	))

	voices := []muse.Voice{}
	for i := 0; i < 20; i++ {
		voice := NewTestVoice(env.Config, adsrControl)
		voices = append(voices, voice)
	}

	voicePlayer := env.AddModule(muse.NewVoicePlayer(1, voices, env.Config, "voicePlayer"))
	// allpass := env.AddModule(allpass.NewAllpass(milliPerBeat*1.5, milliPerBeat*1.5, 0.1, env.Config, "allpass"))

	// muse.Connect(voicePlayer, 0, allpass, 0)
	muse.Connect(voicePlayer, 0, env, 0)
	// muse.Connect(allpass, 0, env, 1)

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

	w.SetContent(
		container.NewVBox(
			container.NewHBox(
				widget.NewButton("Start", func() {
					stream.Start()
				}),
				widget.NewButton("Stop", func() {
					stream.Stop()
				}),
				widget.NewButton("Save", func() {
					d := dialog.NewFileSave(func(wc fyne.URIWriteCloser, err error) {
						adsrState := adsrControl.GetState()
						jsonData, _ := json.Marshal(adsrState)
						wc.Write(jsonData)
						wc.Close()
					}, w)

					d.Show()
				}),
				widget.NewButton("Load", func() {
					d := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
						data, _ := ioutil.ReadAll(rc)
						rc.Close()
						var state map[string]any
						json.Unmarshal(data, &state)
						adsrControl.SetState(state)
					}, w)

					d.Show()
				}),
			),
			adsrControl.UI(),
			swingControl.UI(),
		),
	)

	w.ShowAndRun()
}

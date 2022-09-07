package adsr

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	"github.com/almerlucke/muse/ui"
)

type ADSRStepProvider interface {
	ADSRSteps() []adsrc.Step
}

type ADSRControl struct {
	title                        string
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

	return widget.NewCard(ctrl.title, "",
		container.NewHBox(
			container.New(ui.NewFixedWidthLayout(250),
				widget.NewCard("Attack", "", container.NewVBox(
					widget.NewLabel("attack duration (ms)"),
					container.NewBorder(nil, nil, nil, container.New(ui.NewFixedWidthLayout(80), attackDurationLabel), attackDurationSlider),
					widget.NewLabel("attack amplitude (0.0 - 1.0)"),
					container.NewBorder(nil, nil, nil, container.New(ui.NewFixedWidthLayout(80), attackLevelLabel), attackLevelSlider),
					widget.NewLabel("attack shape (-1.0 - 1.0)"),
					container.NewBorder(nil, nil, nil, container.New(ui.NewFixedWidthLayout(80), attackShapeLabel), attackShapeSlider),
				)),
			),
			container.New(ui.NewFixedWidthLayout(250),
				widget.NewCard("Decay", "", container.NewVBox(
					widget.NewLabel("decay duration (ms)"),
					container.NewBorder(nil, nil, nil, container.New(ui.NewFixedWidthLayout(80), decayDurationLabel), decayDurationSlider),
					widget.NewLabel("decay amplitude (0.0 - 1.0)"),
					container.NewBorder(nil, nil, nil, container.New(ui.NewFixedWidthLayout(80), decayLevelLabel), decayLevelSlider),
					widget.NewLabel("decay shape (-1.0 - 1.0)"),
					container.NewBorder(nil, nil, nil, container.New(ui.NewFixedWidthLayout(80), decayShapeLabel), decayShapeSlider),
				)),
			),
			container.New(ui.NewFixedWidthLayout(250),
				widget.NewCard("Release", "", container.NewVBox(
					widget.NewLabel("release duration (ms)"),
					container.NewBorder(nil, nil, nil, container.New(ui.NewFixedWidthLayout(80), releaseDurationLabel), releaseDurationSlider),
					widget.NewLabel("release shape (-1.0 - 1.0)"),
					container.NewBorder(nil, nil, nil, container.New(ui.NewFixedWidthLayout(80), releaseShapeLabel), releaseShapeSlider),
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

func (ctrl *ADSRControl) SetAttackDuration(ms float64) {
	ctrl.steps[0].Duration = ms
	ctrl.attackDurationSliderBinding.Set(ctrl.steps[0].Duration)
}

func (ctrl *ADSRControl) SetAttackLevel(level float64) {
	ctrl.steps[0].Level = level
	ctrl.attackLevelSliderBinding.Set(ctrl.steps[0].Level)
}

func (ctrl *ADSRControl) SetAttackShape(shape float64) {
	ctrl.steps[0].Shape = shape
	ctrl.attackShapeSliderBinding.Set(ctrl.steps[0].Shape)
}

func (ctrl *ADSRControl) SetDecayDuration(ms float64) {
	ctrl.steps[1].Duration = ms
	ctrl.decayDurationSliderBinding.Set(ctrl.steps[1].Duration)
}

func (ctrl *ADSRControl) SetDecayLevel(level float64) {
	ctrl.steps[1].Level = level
	ctrl.decayLevelSliderBinding.Set(ctrl.steps[1].Level)
}

func (ctrl *ADSRControl) SetDecayShape(shape float64) {
	ctrl.steps[1].Shape = shape
	ctrl.decayShapeSliderBinding.Set(ctrl.steps[1].Shape)
}

func (ctrl *ADSRControl) SetReleaseDuration(ms float64) {
	ctrl.steps[3].Duration = ms
	ctrl.releaseDurationSliderBinding.Set(ctrl.steps[3].Duration)
}

func (ctrl *ADSRControl) SetReleaseShape(shape float64) {
	ctrl.steps[3].Shape = shape
	ctrl.releaseShapeSliderBinding.Set(ctrl.steps[3].Shape)
}

func NewADSRControl(title string) *ADSRControl {
	control := &ADSRControl{
		title: title,
		steps: []adsrc.Step{
			{Level: 1.0, Duration: 5, Shape: 0.0},
			{Level: 0.3, Duration: 5, Shape: 0.0},
			{Duration: 20},
			{Duration: 5, Shape: 0.0},
		},
	}

	return control
}

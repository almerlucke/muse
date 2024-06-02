package adsr

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

	adsrc "github.com/almerlucke/genny/float/envelopes/adsr"
	"github.com/almerlucke/muse/ui"
)

type Control struct {
	title                        string
	setting                      *adsrc.Setting
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

func (ctrl *Control) UI() fyne.CanvasObject {
	// Attack Duration
	ctrl.attackDurationLabelBinding = binding.NewString()
	attackDurationLabel := widget.NewLabelWithData(ctrl.attackDurationLabelBinding)
	attackDurationLabel.Alignment = fyne.TextAlignTrailing
	_ = ctrl.attackDurationLabelBinding.Set(fmt.Sprintf("%.2f", ctrl.setting.AttackDuration))

	ctrl.attackDurationSliderBinding = binding.NewFloat()
	ctrl.attackDurationSliderBinding.AddListener(binding.NewDataListener(func() {
		v, err := ctrl.attackDurationSliderBinding.Get()
		if err == nil {
			_ = ctrl.attackDurationLabelBinding.Set(fmt.Sprintf("%.2f", v))
			ctrl.setting.AttackDuration = v
		}
	}))
	attackDurationSlider := widget.NewSliderWithData(1.0, 500.0, ctrl.attackDurationSliderBinding)
	_ = ctrl.attackDurationSliderBinding.Set(ctrl.setting.AttackDuration)

	// Attack Level
	ctrl.attackLevelLabelBinding = binding.NewString()
	attackLevelLabel := widget.NewLabelWithData(ctrl.attackLevelLabelBinding)
	attackLevelLabel.Alignment = fyne.TextAlignTrailing
	_ = ctrl.attackLevelLabelBinding.Set(fmt.Sprintf("%.2f", ctrl.setting.AttackLevel))

	ctrl.attackLevelSliderBinding = binding.NewFloat()
	ctrl.attackLevelSliderBinding.AddListener(binding.NewDataListener(func() {
		v, err := ctrl.attackLevelSliderBinding.Get()
		if err == nil {
			_ = ctrl.attackLevelLabelBinding.Set(fmt.Sprintf("%.2f", v))
			ctrl.setting.AttackLevel = v
		}
	}))
	attackLevelSlider := widget.NewSliderWithData(0.0, 1.0, ctrl.attackLevelSliderBinding)
	attackLevelSlider.Step = 0.01
	_ = ctrl.attackLevelSliderBinding.Set(ctrl.setting.AttackLevel)

	// Attack Shape
	ctrl.attackShapeLabelBinding = binding.NewString()
	attackShapeLabel := widget.NewLabelWithData(ctrl.attackShapeLabelBinding)
	attackShapeLabel.Alignment = fyne.TextAlignTrailing
	_ = ctrl.attackShapeLabelBinding.Set(fmt.Sprintf("%.2f", ctrl.setting.AttackShape))

	ctrl.attackShapeSliderBinding = binding.NewFloat()
	ctrl.attackShapeSliderBinding.AddListener(binding.NewDataListener(func() {
		v, err := ctrl.attackShapeSliderBinding.Get()
		if err == nil {
			_ = ctrl.attackShapeLabelBinding.Set(fmt.Sprintf("%.2f", v))
			ctrl.setting.AttackShape = v
		}
	}))
	attackShapeSlider := widget.NewSliderWithData(-1.0, 1.0, ctrl.attackShapeSliderBinding)
	attackShapeSlider.Step = 0.01
	attackShapeSlider.Value = ctrl.setting.AttackShape
	_ = ctrl.attackShapeSliderBinding.Set(ctrl.setting.AttackShape)

	// Decay Duration
	ctrl.decayDurationLabelBinding = binding.NewString()
	decayDurationLabel := widget.NewLabelWithData(ctrl.decayDurationLabelBinding)
	decayDurationLabel.Alignment = fyne.TextAlignTrailing
	_ = ctrl.decayDurationLabelBinding.Set(fmt.Sprintf("%.2f", ctrl.setting.DecayDuration))

	ctrl.decayDurationSliderBinding = binding.NewFloat()
	ctrl.decayDurationSliderBinding.AddListener(binding.NewDataListener(func() {
		v, err := ctrl.decayDurationSliderBinding.Get()
		if err == nil {
			_ = ctrl.decayDurationLabelBinding.Set(fmt.Sprintf("%.2f", v))
			ctrl.setting.DecayDuration = v
		}
	}))
	decayDurationSlider := widget.NewSliderWithData(1.0, 500.0, ctrl.decayDurationSliderBinding)
	_ = ctrl.decayDurationSliderBinding.Set(ctrl.setting.DecayDuration)

	// Decay Level
	ctrl.decayLevelLabelBinding = binding.NewString()
	decayLevelLabel := widget.NewLabelWithData(ctrl.decayLevelLabelBinding)
	decayLevelLabel.Alignment = fyne.TextAlignTrailing
	_ = ctrl.decayLevelLabelBinding.Set(fmt.Sprintf("%.2f", ctrl.setting.DecayLevel))

	ctrl.decayLevelSliderBinding = binding.NewFloat()
	ctrl.decayLevelSliderBinding.AddListener(binding.NewDataListener(func() {
		v, err := ctrl.decayLevelSliderBinding.Get()
		if err == nil {
			_ = ctrl.decayLevelLabelBinding.Set(fmt.Sprintf("%.2f", v))
			ctrl.setting.DecayLevel = v
		}
	}))
	decayLevelSlider := widget.NewSliderWithData(0.0, 1.0, ctrl.decayLevelSliderBinding)
	decayLevelSlider.Step = 0.01
	_ = ctrl.decayLevelSliderBinding.Set(ctrl.setting.DecayLevel)

	// Decay Shape
	ctrl.decayShapeLabelBinding = binding.NewString()
	decayShapeLabel := widget.NewLabelWithData(ctrl.decayShapeLabelBinding)
	decayShapeLabel.Alignment = fyne.TextAlignTrailing
	_ = ctrl.decayShapeLabelBinding.Set(fmt.Sprintf("%.2f", ctrl.setting.DecayShape))

	ctrl.decayShapeSliderBinding = binding.NewFloat()
	ctrl.decayShapeSliderBinding.AddListener(binding.NewDataListener(func() {
		v, err := ctrl.decayShapeSliderBinding.Get()
		if err == nil {
			_ = ctrl.decayShapeLabelBinding.Set(fmt.Sprintf("%.2f", v))
			ctrl.setting.DecayShape = v
		}
	}))
	decayShapeSlider := widget.NewSliderWithData(-1.0, 1.0, ctrl.decayShapeSliderBinding)
	decayShapeSlider.Step = 0.01
	_ = ctrl.decayShapeSliderBinding.Set(ctrl.setting.DecayShape)

	// Release Duration
	ctrl.releaseDurationLabelBinding = binding.NewString()
	releaseDurationLabel := widget.NewLabelWithData(ctrl.releaseDurationLabelBinding)
	releaseDurationLabel.Alignment = fyne.TextAlignTrailing
	_ = ctrl.releaseDurationLabelBinding.Set(fmt.Sprintf("%.2f", ctrl.setting.ReleaseDuration))

	ctrl.releaseDurationSliderBinding = binding.NewFloat()
	ctrl.releaseDurationSliderBinding.AddListener(binding.NewDataListener(func() {
		v, err := ctrl.releaseDurationSliderBinding.Get()
		if err == nil {
			_ = ctrl.releaseDurationLabelBinding.Set(fmt.Sprintf("%.2f", v))
			ctrl.setting.ReleaseDuration = v
		}
	}))
	releaseDurationSlider := widget.NewSliderWithData(50.0, 2500.0, ctrl.releaseDurationSliderBinding)
	releaseDurationSlider.Step = 10.0
	_ = ctrl.releaseDurationSliderBinding.Set(ctrl.setting.ReleaseDuration)

	// Release Shape
	ctrl.releaseShapeLabelBinding = binding.NewString()
	releaseShapeLabel := widget.NewLabelWithData(ctrl.releaseShapeLabelBinding)
	releaseShapeLabel.Alignment = fyne.TextAlignTrailing
	_ = ctrl.releaseShapeLabelBinding.Set(fmt.Sprintf("%.2f", ctrl.setting.ReleaseShape))

	ctrl.releaseShapeSliderBinding = binding.NewFloat()
	ctrl.releaseShapeSliderBinding.AddListener(binding.NewDataListener(func() {
		v, err := ctrl.releaseShapeSliderBinding.Get()
		if err == nil {
			_ = ctrl.releaseShapeLabelBinding.Set(fmt.Sprintf("%.2f", v))
			ctrl.setting.ReleaseShape = v
		}
	}))
	releaseShapeSlider := widget.NewSliderWithData(-1.0, 1.0, ctrl.releaseShapeSliderBinding)
	releaseShapeSlider.Step = 0.01
	_ = ctrl.releaseShapeSliderBinding.Set(ctrl.setting.ReleaseShape)

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

func (ctrl *Control) Setting() *adsrc.Setting {
	return ctrl.setting
}

func (ctrl *Control) SetAttackDuration(ms float64) {
	ctrl.setting.AttackDuration = ms
	if ctrl.attackDurationSliderBinding != nil {
		_ = ctrl.attackDurationSliderBinding.Set(ms)
	}
}

func (ctrl *Control) SetAttackLevel(level float64) {
	ctrl.setting.AttackLevel = level
	if ctrl.attackLevelSliderBinding != nil {
		_ = ctrl.attackLevelSliderBinding.Set(level)
	}
}

func (ctrl *Control) SetAttackShape(shape float64) {
	ctrl.setting.AttackShape = shape
	if ctrl.attackShapeSliderBinding != nil {
		_ = ctrl.attackShapeSliderBinding.Set(shape)
	}
}

func (ctrl *Control) SetDecayDuration(ms float64) {
	ctrl.setting.DecayDuration = ms
	if ctrl.decayDurationSliderBinding != nil {
		_ = ctrl.decayDurationSliderBinding.Set(ms)
	}
}

func (ctrl *Control) SetDecayLevel(level float64) {
	ctrl.setting.DecayLevel = level
	if ctrl.decayLevelSliderBinding != nil {
		_ = ctrl.decayLevelSliderBinding.Set(level)
	}
}

func (ctrl *Control) SetDecayShape(shape float64) {
	ctrl.setting.DecayShape = shape
	if ctrl.decayShapeSliderBinding != nil {
		_ = ctrl.decayShapeSliderBinding.Set(shape)
	}
}

func (ctrl *Control) SetReleaseDuration(ms float64) {
	ctrl.setting.ReleaseDuration = ms
	if ctrl.releaseDurationSliderBinding != nil {
		_ = ctrl.releaseDurationSliderBinding.Set(ms)
	}
}

func (ctrl *Control) SetReleaseShape(shape float64) {
	ctrl.setting.ReleaseShape = shape
	if ctrl.releaseShapeSliderBinding != nil {
		_ = ctrl.releaseShapeSliderBinding.Set(shape)
	}
}

func NewADSRControl(title string, setting *adsrc.Setting) *Control {
	control := &Control{
		title:   title,
		setting: setting,
	}

	return control
}

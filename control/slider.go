package control

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/almerlucke/muse/ui"
)

type Slider struct {
	*Control
	min   float64
	max   float64
	step  float64
	value float64
}

func NewSlider(id string, name string, min float64, max float64, step float64, value float64) *Slider {
	return &Slider{
		Control: NewControl(id, name, SliderType),
		min:     min,
		max:     max,
		step:    step,
		value:   value,
	}
}

func (sc *Slider) Min() float64 {
	return sc.min
}

func (sc *Slider) Max() float64 {
	return sc.max
}

func (sc *Slider) Step() float64 {
	return sc.step
}

func (sc *Slider) Get() float64 {
	return sc.value
}

func (sc *Slider) Set(newValue float64, setter any) {
	if sc.value != newValue && newValue >= sc.min && newValue <= sc.max {
		oldValue := sc.value
		sc.value = newValue
		sc.SendChangeToListeners(sc, oldValue, newValue, setter)
	}
}

func (sc *Slider) AddListener(listener Listener) {
	sc.Control.AddListener(listener)
	listener.ControlChanged(sc, sc.value, sc.value, sc)
}

func (sc *Slider) UI() fyne.CanvasObject {
	floatValueLabelBinding := binding.NewString()
	floatValueLabelBinding.Set(fmt.Sprintf("%.2f", sc.value))

	floatValueLabel := widget.NewLabelWithData(floatValueLabelBinding)
	floatValueLabel.Alignment = fyne.TextAlignTrailing

	valueBinding := binding.NewFloat()
	valueBinding.Set(sc.value)

	valueBinding.AddListener(binding.NewDataListener(func() {
		v, err := valueBinding.Get()
		if err == nil {
			floatValueLabelBinding.Set(fmt.Sprintf("%.2f", v))
			sc.Set(v, valueBinding)
		}
	}))

	sc.AddListener(NewChangeCallback(func(ctrl ControlProtocol, oldValue any, newValue any, setter any) {
		if setter != valueBinding {
			valueBinding.Set(newValue.(float64))
		}
	}))

	valueSlider := widget.NewSliderWithData(sc.min, sc.max, valueBinding)
	valueSlider.Step = sc.step

	return container.NewVBox(
		widget.NewLabel(sc.DisplayName()),
		container.NewBorder(nil, nil, nil,
			ui.NewFixedWidthContainer(70, floatValueLabel),
			ui.NewFixedWidthContainer(140, valueSlider),
		),
	)
}

package control

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Select struct {
	*BaseControl
	selected string
	options  []string
}

func NewSelect(id string, name string, options []string, selected string) *Select {
	return &Select{
		BaseControl: NewBaseControl(id, name, SelectType),
		selected:    selected,
		options:     options,
	}
}

func (sel *Select) Set(newValue string, setter any) {
	if sel.selected != newValue {
		oldValue := sel.selected
		sel.selected = newValue
		sel.SendChangeToListeners(sel, oldValue, newValue, setter)
	}
}

func (sel *Select) AddListener(listener Listener) {
	sel.BaseControl.AddListener(listener)
	listener.ControlChanged(sel, sel.selected, sel.selected, sel)
}

func (sel *Select) UI() fyne.CanvasObject {
	selectWidget := widget.NewSelect(sel.options, func(newValue string) {
		sel.Set(newValue, nil)
	})

	selectWidget.Alignment = fyne.TextAlignLeading
	selectWidget.SetSelected(sel.selected)

	return container.NewVBox(
		widget.NewLabel(sel.DisplayName()),
		selectWidget,
	)

	// floatValueLabelBinding := binding.NewString()
	// floatValueLabelBinding.Set(fmt.Sprintf("%.2f", sel.value))

	// floatValueLabel := widget.NewLabelWithData(floatValueLabelBinding)
	// floatValueLabel.Alignment = fyne.TextAlignTrailing

	// valueBinding := binding.NewFloat()
	// valueBinding.Set(sel.value)

	// valueBinding.AddListener(binding.NewDataListener(func() {
	// 	v, err := valueBinding.Get()
	// 	if err == nil {
	// 		floatValueLabelBinding.Set(fmt.Sprintf("%.2f", v))
	// 		sc.Set(v, valueBinding)
	// 	}
	// }))

	// sc.AddListener(NewChangeCallback(func(ctrl Control, oldValue any, newValue any, setter any) {
	// 	if setter != valueBinding {
	// 		valueBinding.Set(newValue.(float64))
	// 	}
	// }))

	// valueSlider := widget.NewSliderWithData(sc.min, sc.max, valueBinding)
	// valueSlider.Step = sc.step

	// return container.NewVBox(
	// 	widget.NewLabel(sc.DisplayName()),
	// 	container.NewBorder(nil, nil, nil,
	// 		ui.NewFixedWidthContainer(70, floatValueLabel),
	// 		ui.NewFixedWidthContainer(140, valueSlider),
	// 	),
	// )
}

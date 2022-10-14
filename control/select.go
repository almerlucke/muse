package control

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Select struct {
	*Control
	selected string
	options  []string
}

func NewSelect(id string, name string, options []string, selected string) *Select {
	return &Select{
		Control:  NewControl(id, name, SelectType),
		selected: selected,
		options:  options,
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
	sel.Control.AddListener(listener)
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
}

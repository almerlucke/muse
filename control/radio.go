package control

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Radio struct {
	*BaseControl
	selected string
	options  []string
}

func NewRadio(id string, name string, options []string, selected string) *Radio {
	return &Radio{
		BaseControl: NewBaseControl(id, name, RadioType),
		selected:    selected,
		options:     options,
	}
}

func (r *Radio) Set(newValue string, setter any) {
	if r.selected != newValue {
		oldValue := r.selected
		r.selected = newValue
		r.SendChangeToListeners(r, oldValue, newValue, setter)
	}
}

func (r *Radio) AddListener(listener Listener) {
	r.BaseControl.AddListener(listener)
	listener.ControlChanged(r, r.selected, r.selected, r)
}

func (r *Radio) UI() fyne.CanvasObject {
	radioWidget := widget.NewRadioGroup(r.options, func(newValue string) {
		r.Set(newValue, nil)
	})

	radioWidget.SetSelected(r.selected)

	return container.NewVBox(
		widget.NewLabel(r.DisplayName()),
		radioWidget,
	)
}

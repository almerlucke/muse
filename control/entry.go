package control

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/almerlucke/muse/ui"
)

type EntryTranformer func(string) any

type Entry struct {
	*BaseControl
	text        string
	transformer EntryTranformer
}

func NewEntry(id string, name string, text string, transformer EntryTranformer) *Entry {
	return &Entry{
		BaseControl: NewBaseControl(id, name, EntryType),
		text:        text,
		transformer: transformer,
	}
}

func (e *Entry) Set(newValue any, setter any) {
	if e.text != newValue {
		oldValue := e.Get()
		e.text = fmt.Sprintf("%v", newValue)
		e.SendChangeToListeners(e, oldValue, e.Get(), setter)
	}
}

func (e *Entry) Get() any {
	if e.transformer != nil {
		return e.transformer(e.text)
	}

	return e.text
}

func (e *Entry) AddListener(listener Listener) {
	v := e.Get()
	e.BaseControl.AddListener(listener)
	listener.ControlChanged(e, v, v, e)
}

func (e *Entry) UI() fyne.CanvasObject {
	entry := widget.NewEntry()
	entry.SetText(e.text)

	saveButton := widget.NewButton("Save", nil)
	saveButton.OnTapped = func() {
		saveButton.Disable()
		e.Set(entry.Text, entry)
	}

	saveButton.Disable()

	entry.OnChanged = func(v string) {
		if entry.Text != e.text {
			saveButton.Enable()
		} else {
			saveButton.Disable()
		}
	}

	e.AddListener(NewChangeCallback(func(ctrl Control, oldValue any, newValue any, setter any) {
		if setter != entry {
			entry.SetText(e.text)
			saveButton.Disable()
		}
	}))

	return container.NewVBox(
		widget.NewLabel(e.DisplayName()),
		container.NewHBox(ui.NewFixedWidthContainer(130, entry), saveButton),
	)
}

package controls

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/almerlucke/muse/ui"
)

type FilePicker struct {
	*BaseControl
	file fyne.URI
	w    fyne.Window
}

func NewFilePicker(id string, name string, w fyne.Window) *FilePicker {
	return &FilePicker{
		BaseControl: NewBaseControl(id, name, FilePickerType),
		w:           w,
	}
}

func (fp *FilePicker) Set(newValue fyne.URI, setter any) {
	if fp.file == nil || (fp.file.String() != newValue.String()) {
		oldValue := fp.file
		fp.file = newValue
		fp.SendChangeToListeners(fp, oldValue, newValue, setter)
	}
}

func (fp *FilePicker) UI() fyne.CanvasObject {
	fileNameLabel := widget.NewLabel("Select a file")

	selectButton := widget.NewButton("File", nil)
	selectButton.OnTapped = func() {
		selectButton.Disable()
		dialog.ShowFileOpen(func(rc fyne.URIReadCloser, err error) {
			if err == nil && rc != nil {
				fp.Set(rc.URI(), nil)
				fileNameLabel.Text = rc.URI().Name()
				fileNameLabel.Refresh()
			}
			selectButton.Enable()
		}, fp.w)
	}

	return container.NewVBox(
		widget.NewLabel(fp.DisplayName()),
		container.NewHBox(ui.NewFixedWidthContainer(130, container.NewHScroll(fileNameLabel)), selectButton),
	)
}

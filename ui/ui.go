package ui

import (
	"fyne.io/fyne/v2"
)

type Listener interface {
	Listen(map[string]any)
}

type Controller interface {
	UI() fyne.CanvasObject
}

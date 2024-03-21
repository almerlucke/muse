package ui

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
)

type Listener interface {
	Listen(map[string]any)
}

type Controller interface {
	UI() fyne.CanvasObject
}

// DelayedExecution hold execution of callback until cnt reaches zero
type DelayedExecution struct {
	mu   sync.Mutex
	cnt  int
	exec func()
}

func NewDelayedExecution(exec func()) *DelayedExecution {
	return &DelayedExecution{
		exec: exec,
	}
}

func (d *DelayedExecution) Inc() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cnt++
}

func (d *DelayedExecution) Dec() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cnt--
	if d.cnt == 0 {
		if d.exec != nil {
			d.exec()
		}
	}
}

// NewDelayedListener only call listener when there hasn't been a change in given duration
func NewDelayedListener(duration time.Duration, listener func()) binding.DataListener {
	delayer := NewDelayedExecution(listener)
	return binding.NewDataListener(func() {
		delayer.Inc()
		time.AfterFunc(duration, func() {
			delayer.Dec()
		})
	})
}

type FixedWidthLayout struct {
	Width float32
}

func NewFixedWidthLayout(w float32) *FixedWidthLayout {
	return &FixedWidthLayout{Width: w}
}

func NewFixedWidthContainer(w float32, object fyne.CanvasObject) *fyne.Container {
	return container.New(NewFixedWidthLayout(w), object)
}

func (fwl *FixedWidthLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	for _, object := range objects {
		object.Resize(fyne.NewSize(containerSize.Width, object.MinSize().Height))
		object.Move(fyne.NewPos(0, (containerSize.Height-object.MinSize().Height)/2.0))
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

type FixedSizeLayout struct {
	size fyne.Size
}

func NewFixedSizeLayout(size fyne.Size) *FixedSizeLayout {
	return &FixedSizeLayout{size: size}
}

func (fsl *FixedSizeLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	for _, object := range objects {
		object.Resize(fsl.size)
		object.Move(fyne.NewPos(0, 0))
	}
}

func (fsl *FixedSizeLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fsl.size
}

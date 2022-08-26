package ui

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
)

type Listener interface {
	Listen(map[string]any)
}

type Controller interface {
	UI() fyne.CanvasObject
}

// DelayedExecution, hold execution of callback until cnt reaches zero
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

// NewDelayedListener, only call listener when there hasn't been a change in given duration
func NewDelayedListener(duration time.Duration, listener func()) binding.DataListener {
	delayer := NewDelayedExecution(listener)
	return binding.NewDataListener(func() {
		delayer.Inc()
		time.AfterFunc(duration, func() {
			delayer.Dec()
		})
	})
}

package control

import "github.com/almerlucke/muse"

type ControlType int

const (
	Float ControlType = iota
	// Int
	// List
	// Switch
	// Radio
)

type Listener interface {
	// ControlChanged is called with the control that changed, the old value,
	// the new value and the setter that changed the value
	ControlChanged(Control, any, any, any)
}

/**
** Callback listener function
**/
type ListenerFunc func(Control, any, any, any)

type Callback struct {
	f ListenerFunc
}

func NewCallback(f ListenerFunc) *Callback {
	return &Callback{f: f}
}

func (c *Callback) ControlChanged(control Control, oldValue any, newValue any, setter any) {
	c.f(control, oldValue, newValue, setter)
}

type Control interface {
	muse.Identifiable
	DisplayName() string
	Type() ControlType
	AddListener(Listener)
	RemoveListener(Listener)
}

type FloatControl interface {
	Control
	Min() float64
	Max() float64
	Step() float64
	Get() float64
	// Set a new value, also pass the setter so listeners can decide
	// to do something with the value or not (if they themselves where the setter)
	Set(float64, any)
}

// type Control struct {
// 	ID          string
// 	DisplayName string
// }

type Controllable interface {
	Controls()
}

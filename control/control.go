package control

import (
	"container/list"

	"github.com/almerlucke/muse"
)

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

type ListenerFunc func(Control, any, any, any)

type ChangeCallback struct {
	f ListenerFunc
}

func NewChangeCallback(f ListenerFunc) *ChangeCallback {
	return &ChangeCallback{f: f}
}

func (c *ChangeCallback) ControlChanged(control Control, oldValue any, newValue any, setter any) {
	c.f(control, oldValue, newValue, setter)
}

type Control interface {
	muse.Identifiable
	Group() string
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

type BaseControl struct {
	identifier  string
	group       string
	displayName string
	listeners   *list.List
	controlType ControlType
}

func NewBaseControl(id string, group string, name string, controlType ControlType) *BaseControl {
	bc := &BaseControl{
		identifier:  id,
		group:       group,
		displayName: name,
		listeners:   list.New(),
		controlType: controlType,
	}

	return bc
}

func (bc *BaseControl) Identifier() string {
	return bc.identifier
}

func (bc *BaseControl) SetIdentifier(identifier string) {
	bc.identifier = identifier
}

func (bc *BaseControl) Group() string {
	return bc.group
}

func (bc *BaseControl) DisplayName() string {
	return bc.displayName
}

func (bc *BaseControl) AddListener(listener Listener) {
	bc.listeners.PushBack(listener)
}

func (bc *BaseControl) RemoveListener(listener Listener) {
	elem := bc.listeners.Front()
	for elem != nil {
		if elem.Value == listener {
			break
		}
		elem = elem.Next()
	}

	if elem != nil {
		bc.listeners.Remove(elem)
	}
}

func (bc *BaseControl) Type() ControlType {
	return bc.controlType
}

func (bc *BaseControl) SendChangeToListeners(control Control, oldValue any, newValue any, setter any) {
	elem := bc.listeners.Front()
	for elem != nil {
		elem.Value.(Listener).ControlChanged(control, oldValue, newValue, setter)
		elem = elem.Next()
	}
}

type BaseFloatControl struct {
	*BaseControl
	min   float64
	max   float64
	step  float64
	value float64
}

func NewBaseFloatControl(id string, group string, name string, min float64, max float64, step float64, value float64) *BaseFloatControl {
	return &BaseFloatControl{
		BaseControl: NewBaseControl(id, group, name, Float),
		min:         min,
		max:         max,
		step:        step,
		value:       value,
	}
}

func (fc *BaseFloatControl) Min() float64 {
	return fc.min
}

func (fc *BaseFloatControl) Max() float64 {
	return fc.max
}

func (fc *BaseFloatControl) Step() float64 {
	return fc.step
}

func (fc *BaseFloatControl) Get() float64 {
	return fc.value
}

func (fc *BaseFloatControl) Set(newValue float64, setter any) {
	if fc.value != newValue && newValue >= fc.min && newValue <= fc.max {
		oldValue := fc.value
		fc.value = newValue
		fc.SendChangeToListeners(fc, oldValue, newValue, setter)
	}
}

func (fc *BaseFloatControl) AddListener(listener Listener) {
	fc.BaseControl.AddListener(listener)
	listener.ControlChanged(fc, fc.value, fc.value, fc)
}

type Collection struct {
	controls []Control
}

func NewCollection() *Collection {
	return &Collection{
		controls: []Control{},
	}
}

func (c *Collection) AddControl(ctrl Control) {
	c.controls = append(c.controls, ctrl)
}

func (c *Collection) Controls() []Control {
	return c.controls
}

func (c *Collection) ControlById(id string) Control {
	for _, ctrl := range c.controls {
		if ctrl.Identifier() == id {
			return ctrl
		}
	}

	return nil
}

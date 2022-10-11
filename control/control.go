package control

import (
	"container/list"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/ui"
)

type ControlType int

const (
	Slider ControlType = iota
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

type UserInterfacer interface {
	DisplayName() string
	UI() fyne.CanvasObject
}

type Control interface {
	muse.Identifiable
	UserInterfacer
	Group() *Group
	SetGroup(*Group)
	Type() ControlType
	AddListener(Listener)
	RemoveListener(Listener)
}

type Group struct {
	id          string
	displayName string
	Controls    []Control
	Parent      *Group
	Children    []*Group
}

func NewGroup(id string, displayName string) *Group {
	return &Group{
		id:          id,
		displayName: displayName,
		Controls:    []Control{},
		Children:    []*Group{},
	}
}

func (g *Group) DisplayName() string {
	return g.displayName
}

func (g *Group) UI() fyne.CanvasObject {
	controlUis := make([]fyne.CanvasObject, len(g.Controls))

	for i, c := range g.Controls {
		controlUis[i] = c.UI()
	}

	childrenUis := make([]fyne.CanvasObject, len(g.Children)+1)
	childrenUis[0] = container.NewVBox(controlUis...)

	for i, child := range g.Children {
		childrenUis[i+1] = child.UI()
	}

	return widget.NewCard(g.displayName, "", container.NewHBox(childrenUis...))

	// 				widget.NewLabel("attack duration (ms)"),
	// 				container.NewBorder(nil, nil, nil, container.New(ui.NewFixedWidthLayout(80), attackDurationLabel), attackDurationSlider),
	// 				widget.NewLabel("attack amplitude (0.0 - 1.0)"),
	// 				container.NewBorder(nil, nil, nil, container.New(ui.NewFixedWidthLayout(80), attackLevelLabel), attackLevelSlider),
	// 				widget.NewLabel("attack shape (-1.0 - 1.0)"),
	// 				container.NewBorder(nil, nil, nil, container.New(ui.NewFixedWidthLayout(80), attackShapeLabel), attackShapeSlider),
	// 			)),
	// 		),
}

func (g *Group) Identifier() string {
	return g.id
}

func (g *Group) SetIdentifier(id string) {
	g.id = id
}

func (g *Group) AddControl(c Control) Control {
	g.Controls = append(g.Controls, c)
	c.SetGroup(g)
	return c
}

func (g *Group) AddChild(child *Group) *Group {
	g.Children = append(g.Children, child)
	child.Parent = g
	return child
}

func (g *Group) ControlById(id string) Control {
	for _, c := range g.Controls {
		if c.Identifier() == id {
			return c
		}
	}

	for _, child := range g.Children {
		c := child.ControlById(id)
		if c != nil {
			return c
		}
	}

	return nil
}

func (g *Group) ChildById(id string) *Group {
	for _, child := range g.Children {
		if child.Identifier() == id {
			return child
		}
	}

	return nil
}

func (g *Group) AddListenerShallow(l Listener) {
	for _, c := range g.Controls {
		c.AddListener(l)
	}
}

func (g *Group) AddListenerDeep(l Listener) {
	for _, c := range g.Controls {
		c.AddListener(l)
	}

	for _, child := range g.Children {
		child.AddListenerDeep(l)
	}
}

type BaseControl struct {
	id          string
	name        string
	group       *Group
	listeners   *list.List
	controlType ControlType
}

func NewBaseControl(id string, name string, controlType ControlType) *BaseControl {
	c := &BaseControl{
		id:          id,
		name:        name,
		listeners:   list.New(),
		controlType: controlType,
	}

	return c
}

func (c *BaseControl) Identifier() string {
	return c.id
}

func (c *BaseControl) SetIdentifier(id string) {
	c.id = id
}

func (c *BaseControl) DisplayName() string {
	return c.name
}

func (c *BaseControl) UI() fyne.CanvasObject {
	return nil // STUB
}

func (c *BaseControl) Group() *Group {
	return c.group
}

func (c *BaseControl) SetGroup(g *Group) {
	c.group = g
}

func (c *BaseControl) AddListener(listener Listener) {
	c.listeners.PushBack(listener)
}

func (c *BaseControl) RemoveListener(listener Listener) {
	elem := c.listeners.Front()
	for elem != nil {
		if elem.Value == listener {
			break
		}
		elem = elem.Next()
	}

	if elem != nil {
		c.listeners.Remove(elem)
	}
}

func (c *BaseControl) Type() ControlType {
	return c.controlType
}

func (c *BaseControl) SendChangeToListeners(control Control, oldValue any, newValue any, setter any) {
	elem := c.listeners.Front()
	for elem != nil {
		elem.Value.(Listener).ControlChanged(control, oldValue, newValue, setter)
		elem = elem.Next()
	}
}

type SliderControl struct {
	*BaseControl
	min   float64
	max   float64
	step  float64
	value float64
}

func NewSliderControl(id string, name string, min float64, max float64, step float64, value float64) *SliderControl {
	return &SliderControl{
		BaseControl: NewBaseControl(id, name, Slider),
		min:         min,
		max:         max,
		step:        step,
		value:       value,
	}
}

func (sc *SliderControl) Min() float64 {
	return sc.min
}

func (sc *SliderControl) Max() float64 {
	return sc.max
}

func (sc *SliderControl) Step() float64 {
	return sc.step
}

func (sc *SliderControl) Get() float64 {
	return sc.value
}

func (sc *SliderControl) Set(newValue float64, setter any) {
	if sc.value != newValue && newValue >= sc.min && newValue <= sc.max {
		oldValue := sc.value
		sc.value = newValue
		sc.SendChangeToListeners(sc, oldValue, newValue, setter)
	}
}

func (sc *SliderControl) AddListener(listener Listener) {
	sc.BaseControl.AddListener(listener)
	listener.ControlChanged(sc, sc.value, sc.value, sc)
}

func (sc *SliderControl) UI() fyne.CanvasObject {
	floatValueLabelBinding := binding.NewString()
	floatValueLabelBinding.Set(fmt.Sprintf("%.2f", sc.value))

	floatValueLabel := widget.NewLabelWithData(floatValueLabelBinding)
	floatValueLabel.Alignment = fyne.TextAlignTrailing

	valueBinding := binding.NewFloat()
	valueBinding.Set(sc.value)

	valueBinding.AddListener(binding.NewDataListener(func() {
		v, err := valueBinding.Get()
		if err == nil {
			floatValueLabelBinding.Set(fmt.Sprintf("%.2f", v))
			sc.Set(v, valueBinding)
		}
	}))

	sc.AddListener(NewChangeCallback(func(ctrl Control, oldValue any, newValue any, setter any) {
		if setter != valueBinding {
			valueBinding.Set(newValue.(float64))
		}
	}))

	valueSlider := widget.NewSliderWithData(sc.min, sc.max, valueBinding)
	valueSlider.Step = sc.step

	return container.NewVBox(
		widget.NewLabel(sc.DisplayName()),
		container.NewBorder(nil, nil, nil,
			ui.NewFixedWidthContainer(70, floatValueLabel),
			ui.NewFixedWidthContainer(140, valueSlider),
		),
	)
}

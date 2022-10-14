package control

import (
	"container/list"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/almerlucke/muse"
)

type Type int

const (
	SliderType Type = iota
	SelectType
	RadioType
	EntryType
	FilePickerType
	// Int
	// List
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
	Type() Type
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
	controlType Type
}

func NewBaseControl(id string, name string, controlType Type) *BaseControl {
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

func (c *BaseControl) Type() Type {
	return c.controlType
}

func (c *BaseControl) SendChangeToListeners(control Control, oldValue any, newValue any, setter any) {
	elem := c.listeners.Front()
	for elem != nil {
		elem.Value.(Listener).ControlChanged(control, oldValue, newValue, setter)
		elem = elem.Next()
	}
}

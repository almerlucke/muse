package fgen

import (
	"github.com/almerlucke/genny/float"
	"github.com/almerlucke/muse"
)

type ControlFunction func(any, int)

type MessageFunction func(any) []*muse.Message

type Gen struct {
	*muse.BaseControl
	gen             float.FrameGenerator
	controlFunction ControlFunction
	messageFunction MessageFunction
}

func NewX(gen float.FrameGenerator, controlFunction ControlFunction, messageFunction MessageFunction) *Gen {
	g := &Gen{
		BaseControl:     muse.NewBaseControl(),
		gen:             gen,
		controlFunction: controlFunction,
		messageFunction: messageFunction,
	}

	g.SetSelf(g)

	return g
}

func New(gen float.FrameGenerator) *Gen {
	return NewX(gen, nil, nil)
}

func (g *Gen) bang() {
	for i, v := range g.gen.Generate() {
		g.SendControlValue(v, i)
	}
}

func (g *Gen) ReceiveControlValue(value any, index int) {
	if index == 0 && muse.IsBang(value) {
		g.bang()
	}

	if g.controlFunction != nil {
		g.controlFunction(value, index)
	}
}

func (g *Gen) ReceiveMessage(msg any) []*muse.Message {
	if muse.IsBang(msg) {
		g.bang()
	}

	if g.messageFunction != nil {
		return g.messageFunction(msg)
	}

	return nil
}

func (g *Gen) Tick(_ int64, _ *muse.Configuration) {
	for i, v := range g.gen.Generate() {
		g.SendControlValue(v, i)
	}
}

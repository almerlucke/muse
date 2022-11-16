package gen

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/generator"
)

type ControlFunction func(any, int)

type MessageFunction func(any) []*muse.Message

type Gen struct {
	*muse.BaseControl
	generator       generator.Generator
	controlFunction ControlFunction
	messageFunction MessageFunction
}

func NewGen(generator generator.Generator, controlFunction ControlFunction, messageFunction MessageFunction, id string) *Gen {
	g := &Gen{
		BaseControl:     muse.NewBaseControl(id),
		generator:       generator,
		controlFunction: controlFunction,
		messageFunction: messageFunction,
	}

	g.SetSelf(g)

	return g
}

func NewBasicGen(generator generator.Generator, id string) *Gen {
	return NewGen(generator, nil, nil, id)
}

func (g *Gen) bang() {
	for i, v := range g.generator.Generate() {
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

func (g *Gen) Tick(timestamp int64, config *muse.Configuration) {
	for i, v := range g.generator.Generate() {
		g.SendControlValue(v, i)
	}
}

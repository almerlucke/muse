package gen

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/muse"
)

type ControlFunction func(any, int)

type MessageFunction func(any) []*muse.Message

type Gen[T any] struct {
	*muse.BaseControl
	gen             genny.Generator[T]
	controlFunction ControlFunction
	messageFunction MessageFunction
}

func NewX[T any](gen genny.Generator[T], controlFunction ControlFunction, messageFunction MessageFunction) *Gen[T] {
	g := &Gen[T]{
		BaseControl:     muse.NewBaseControl(),
		gen:             gen,
		controlFunction: controlFunction,
		messageFunction: messageFunction,
	}

	g.SetSelf(g)

	return g
}

func New[T any](gen genny.Generator[T]) *Gen[T] {
	return NewX(gen, nil, nil)
}

func (g *Gen[T]) bang() {
	g.SendControlValue(g.gen.Generate(), 0)
}

func (g *Gen[T]) ReceiveControlValue(value any, index int) {
	if index == 0 && muse.IsBang(value) {
		g.bang()
	}

	if g.controlFunction != nil {
		g.controlFunction(value, index)
	}
}

func (g *Gen[T]) ReceiveMessage(msg any) []*muse.Message {
	if muse.IsBang(msg) {
		g.bang()
	}

	if g.messageFunction != nil {
		return g.messageFunction(msg)
	}

	return nil
}

func (g *Gen[T]) Tick(_ int64, _ *muse.Configuration) {
	g.SendControlValue(g.gen.Generate(), 0)
}

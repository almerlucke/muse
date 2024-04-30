package gen

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/muse"
)

type ControlFunction[T any] func(genny.Generator[T], any, int)

type MessageFunction[T any] func(genny.Generator[T], any) []*muse.Message

type Gen[T any] struct {
	*muse.BaseControl
	gen             genny.Generator[T]
	controlFunction ControlFunction[T]
	messageFunction MessageFunction[T]
	useTick         bool
}

func NewWithFunctions[T any](gen genny.Generator[T], useTick bool, controlFunction ControlFunction[T], messageFunction MessageFunction[T]) *Gen[T] {
	g := &Gen[T]{
		BaseControl:     muse.NewBaseControl(),
		gen:             gen,
		controlFunction: controlFunction,
		messageFunction: messageFunction,
		useTick:         useTick,
	}

	g.SetSelf(g)

	return g
}

func New[T any](gen genny.Generator[T], useTick bool) *Gen[T] {
	return NewWithFunctions(gen, useTick, nil, nil)
}

func (g *Gen[T]) bang() {
	g.SendControlValue(g.gen.Generate(), 0)
}

func (g *Gen[T]) ReceiveControlValue(value any, index int) {
	if muse.IsBang(value) {
		switch index {
		case 0:
			v := g.gen.Generate()

			if g.gen.Done() {
				g.SendControlValue(muse.Bang, 1)
			}

			g.SendControlValue(v, 0)
		case 1:
			g.gen.Reset()
		default:
			break
		}
	}

	if g.controlFunction != nil {
		g.controlFunction(g.gen, value, index)
	}
}

func (g *Gen[T]) ReceiveMessage(msg any) []*muse.Message {
	if muse.IsBang(msg) {
		g.bang()
	}

	if g.messageFunction != nil {
		return g.messageFunction(g.gen, msg)
	}

	return nil
}

func (g *Gen[T]) Tick(_ int64, _ *muse.Configuration) {
	if g.useTick {
		g.SendControlValue(g.gen.Generate(), 0)
	}
}

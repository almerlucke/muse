package generator

import (
	"github.com/almerlucke/genny/float"
	"github.com/almerlucke/muse"
)

type ControlFunction func(any, int)

type MessageFunction func(any) []*muse.Message

type Generator struct {
	*muse.BaseModule
	gen             float.FrameGenerator
	controlFunction ControlFunction
	messageFunction MessageFunction
}

func NewBasic(gen float.FrameGenerator) *Generator {
	return New(gen, nil, nil)
}

func New(gen float.FrameGenerator, controlFunction ControlFunction, messageFunction MessageFunction) *Generator {
	gg := &Generator{
		BaseModule:      muse.NewBaseModule(0, gen.Dimensions()),
		gen:             gen,
		controlFunction: controlFunction,
		messageFunction: messageFunction,
	}

	gg.SetSelf(gg)

	return gg
}

func (g *Generator) ReceiveControlValue(value any, index int) {
	if g.controlFunction != nil {
		g.controlFunction(value, index)
	}
}

func (g *Generator) ReceiveMessage(msg any) []*muse.Message {
	if g.messageFunction != nil {
		return g.messageFunction(msg)
	}

	return nil
}

func (g *Generator) Synthesize() bool {
	if !g.BaseModule.Synthesize() {
		return false
	}

	numDim := g.gen.Dimensions()

	for i := 0; i < g.Config.BufferSize; i++ {
		vs := g.gen.Generate()
		for dim := 0; dim < numDim; dim++ {
			g.Outputs[dim].Buffer[i] = vs[dim]
		}
	}

	return true
}

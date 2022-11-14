package generator

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/generator"
)

type ControlFunction func(any, int)

type MessageFunction func(any) []*muse.Message

type Generator struct {
	*muse.BaseModule
	generator       generator.Generator
	controlFunction ControlFunction
	messageFunction MessageFunction
}

func NewBasicGenerator(generator generator.Generator, config *muse.Configuration, id string) *Generator {
	return NewGenerator(generator, nil, nil, config, id)
}

func NewGenerator(generator generator.Generator, controlFunction ControlFunction, messageFunction MessageFunction, config *muse.Configuration, id string) *Generator {
	gen := &Generator{
		BaseModule:      muse.NewBaseModule(0, 1, config, id),
		generator:       generator,
		controlFunction: controlFunction,
		messageFunction: messageFunction,
	}

	gen.SetSelf(gen)

	return gen
}

func (gen *Generator) ReceiveControlValue(value any, index int) {
	if gen.controlFunction != nil {
		gen.controlFunction(value, index)
	}
}

func (gen *Generator) ReceiveMessage(msg any) []*muse.Message {
	if gen.messageFunction != nil {
		return gen.messageFunction(msg)
	}

	return nil
}

func (gen *Generator) Synthesize() bool {
	if !gen.BaseModule.Synthesize() {
		return false
	}

	out := gen.Outputs[0].Buffer

	for i := 0; i < gen.Config.BufferSize; i++ {
		out[i] = gen.generator.Tick()[0]
	}

	return true
}

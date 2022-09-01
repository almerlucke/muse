package banger

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/values"
)

func IsBang(msg any) bool {
	bang, ok := msg.(string)

	return ok && bang == "bang"
}

type Banger interface {
	Bang() []*muse.Message
}

type Generator struct {
	*muse.BaseMessenger
	banger Banger
}

func NewGenerator(banger Banger, identifier string) *Generator {
	return &Generator{
		BaseMessenger: muse.NewBaseMessenger(identifier),
		banger:        banger,
	}
}

func (g *Generator) ReceiveMessage(msg any) []*muse.Message {
	if IsBang(msg) {
		return g.banger.Bang()
	}

	return nil
}

type ValueBang struct {
	value values.Valuer[[]*muse.Message]
}

func NewValueBang(value values.Valuer[[]*muse.Message]) *ValueBang {
	return &ValueBang{
		value: value,
	}
}

func NewValueGenerator(value values.Valuer[[]*muse.Message], identifier string) *Generator {
	return NewGenerator(NewValueBang(value), identifier)
}

func (vb *ValueBang) Bang() []*muse.Message {
	msgs := vb.value.Value()

	if vb.value.Finished() {
		vb.value.Reset()
	}

	return msgs
}

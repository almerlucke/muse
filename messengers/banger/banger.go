package banger

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/value"
	"github.com/almerlucke/muse/value/template"
)

type bang struct{ Bang string }

var Bang = &bang{Bang: "bang"}

func IsBang(msg any) bool {
	return msg == Bang
}

type Banger interface {
	MessageBang() []*muse.Message
	ControlBang() []any
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

func (g *Generator) ReceiveControlValue(value any, index int) {
	if index == 0 && value == Bang {
		contents := g.banger.ControlBang()
		for _, content := range contents {
			g.SendControlValue(content, 0)
		}
	}
}

func (g *Generator) ReceiveMessage(msg any) []*muse.Message {
	if msg == Bang {
		return g.banger.MessageBang()
	}

	return nil
}

type ValueBang struct {
	value value.Valuer[[]*muse.Message]
}

func NewValueBang(val value.Valuer[[]*muse.Message]) *ValueBang {
	return &ValueBang{
		value: val,
	}
}

func NewValueGenerator(val value.Valuer[[]*muse.Message], identifier string) *Generator {
	return NewGenerator(NewValueBang(val), identifier)
}

func (vb *ValueBang) MessageBang() []*muse.Message {
	msgs := vb.value.Value()

	if vb.value.Finished() {
		vb.value.Reset()
	}

	return msgs
}

func (vb *ValueBang) ControlBang() []any {
	msgs := vb.value.Value()

	if vb.value.Finished() {
		vb.value.Reset()
	}

	contents := make([]any, len(msgs))
	for i, msg := range msgs {
		contents[i] = msg.Content
	}

	return contents
}

type templateDestination struct {
	addresses []string
	template  template.Template
}

func newTemplateDestination(addresses []string, template template.Template) *templateDestination {
	return &templateDestination{
		addresses: addresses,
		template:  template,
	}
}

func NewTemplateGenerator(addresses []string, template template.Template, identifier string) *Generator {
	return NewGenerator(newTemplateDestination(addresses, template), identifier)
}

func (d *templateDestination) MessageBang() []*muse.Message {
	allMessages := []*muse.Message{}
	protoMessages := d.template.Value()

	for _, address := range d.addresses {
		for _, message := range protoMessages {
			allMessages = append(allMessages, muse.NewMessage(address, message))
		}
	}

	return allMessages
}

func (d *templateDestination) ControlBang() []any {
	protoMessages := d.template.Value()

	contents := make([]any, len(protoMessages))
	for i, msg := range protoMessages {
		contents[i] = msg
	}

	return contents
}

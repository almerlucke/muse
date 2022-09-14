package banger

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/values"
	"github.com/almerlucke/muse/values/template"
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

func (d *templateDestination) Bang() []*muse.Message {
	allMessages := []*muse.Message{}
	protoMessages := d.template.Value()

	for _, address := range d.addresses {
		for _, message := range protoMessages {
			allMessages = append(allMessages, muse.NewMessage(address, message))
		}
	}

	return allMessages
}

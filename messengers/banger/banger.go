package banger

import (
	"fmt"

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
	muse.MessageReceiver
	muse.ControlReceiver
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
	if index == 0 {
		if value == Bang {
			contents := g.banger.ControlBang()
			for _, content := range contents {
				g.SendControlValue(content, 0)
			}
		}
	} else {
		g.banger.ReceiveControlValue(value, index)
	}
}

func (g *Generator) ReceiveMessage(msg any) []*muse.Message {
	if msg == Bang {
		msgsToSend := g.banger.MessageBang()
		for _, msgToSend := range msgsToSend {
			g.SendControlValue(msgToSend.Content, 0)
		}

		return msgsToSend
	}

	return g.banger.ReceiveMessage(msg)
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

func (vb *ValueBang) ReceiveControlValue(value any, index int) {
	// STUB
}

func (vb *ValueBang) ReceiveMessage(msg any) []*muse.Message {
	// STUB
	return nil
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
	paramMap  map[string]any
}

func newTemplateDestination(addresses []string, template template.Template) *templateDestination {
	return &templateDestination{
		addresses: addresses,
		template:  template,
		paramMap:  map[string]any{},
	}
}

func NewTemplateGenerator(addresses []string, template template.Template, identifier string) *Generator {
	return NewGenerator(newTemplateDestination(addresses, template), identifier)
}

func NewControlTemplateGenerator(template template.Template, identifier string) *Generator {
	return NewGenerator(newTemplateDestination(nil, template), identifier)
}

func (d *templateDestination) ReceiveControlValue(value any, index int) {
	d.paramMap[fmt.Sprintf("controlInput%d", index)] = value
}

func (d *templateDestination) ReceiveMessage(msg any) []*muse.Message {
	d.paramMap = msg.(map[string]any)
	return nil
}

func (d *templateDestination) resolveParams() {
	params := []*template.Parameter{}
	for k, v := range d.paramMap {
		params = append(params, template.NewParameter(k, v))
	}

	d.template.SetParameters(params)
}

func (d *templateDestination) MessageBang() []*muse.Message {
	allMessages := []*muse.Message{}

	d.resolveParams()

	protoMessages := d.template.Value()

	for _, address := range d.addresses {
		for _, message := range protoMessages {
			allMessages = append(allMessages, muse.NewMessage(address, message))
		}
	}

	return allMessages
}

func (d *templateDestination) ControlBang() []any {
	d.resolveParams()

	protoMessages := d.template.Value()

	contents := make([]any, len(protoMessages))
	for i, msg := range protoMessages {
		contents[i] = msg
	}

	return contents
}

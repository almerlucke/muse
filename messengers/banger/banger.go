package banger

import (
	"fmt"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/value"
	"github.com/almerlucke/muse/value/template"
)

type Banger interface {
	muse.MessageReceiver
	muse.ControlReceiver
	MessageBang(msg any) []*muse.Message
	ControlBang() []any
}

type Generator struct {
	*muse.BaseMessenger
	banger Banger
}

func NewGenerator(banger Banger) *Generator {
	g := &Generator{
		BaseMessenger: muse.NewBaseMessenger(),
		banger:        banger,
	}

	g.SetSelf(g)

	return g
}

func (g *Generator) ReceiveControlValue(value any, index int) {
	if index == 0 {
		if value == muse.Bang {
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
	if muse.IsBang(msg) {
		msgsToSend := g.banger.MessageBang(msg)
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

func NewValueGenerator(val value.Valuer[[]*muse.Message]) *Generator {
	return NewGenerator(NewValueBang(val))
}

func (vb *ValueBang) ReceiveControlValue(value any, index int) {
	// STUB
}

func (vb *ValueBang) ReceiveMessage(msg any) []*muse.Message {
	// STUB
	return nil
}

func (vb *ValueBang) MessageBang(msg any) []*muse.Message {
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

func NewTemplateGenerator(addresses []string, template template.Template) *Generator {
	return NewGenerator(newTemplateDestination(addresses, template))
}

func NewControlTemplateGenerator(template template.Template) *Generator {
	return NewGenerator(newTemplateDestination(nil, template))
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

func (d *templateDestination) MessageBang(msg any) []*muse.Message {
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

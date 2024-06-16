package banger

import (
	"fmt"
	"github.com/almerlucke/genny"

	"github.com/almerlucke/genny/template"
	"github.com/almerlucke/muse"
)

type Banger interface {
	muse.MessageReceiver
	muse.ControlReceiver
	MessageBang(msg any) []*muse.Message
	ControlBang() []any
}

type Bang struct {
	*muse.BaseMessenger
	banger Banger
}

func NewBang(banger Banger) *Bang {
	b := &Bang{
		BaseMessenger: muse.NewBaseMessenger(),
		banger:        banger,
	}

	b.SetSelf(b)

	return b
}

func (b *Bang) ReceiveControlValue(value any, index int) {
	if index == 0 {
		if value == muse.Bang {
			contents := b.banger.ControlBang()
			for _, content := range contents {
				b.SendControlValue(content, 0)
			}
		}
	} else {
		b.banger.ReceiveControlValue(value, index)
	}
}

func (b *Bang) ReceiveMessage(msg any) []*muse.Message {
	if muse.IsBang(msg) {
		msgsToSend := b.banger.MessageBang(msg)
		for _, msgToSend := range msgsToSend {
			b.SendControlValue(msgToSend.Content, 0)
		}

		return msgsToSend
	}

	return b.banger.ReceiveMessage(msg)
}

type genBanger struct {
	gen genny.Generator[[]*muse.Message]
}

func newGenBanger(gen genny.Generator[[]*muse.Message]) *genBanger {
	return &genBanger{
		gen: gen,
	}
}

func NewGenBang(gen genny.Generator[[]*muse.Message]) *Bang {
	return NewBang(newGenBanger(gen))
}

func (gb *genBanger) ReceiveControlValue(value any, index int) {
	// STUB
}

func (gb *genBanger) ReceiveMessage(msg any) []*muse.Message {
	// STUB
	return nil
}

func (gb *genBanger) MessageBang(msg any) []*muse.Message {
	msgs := gb.gen.Generate()

	if gb.gen.Done() {
		gb.gen.Reset()
	}

	return msgs
}

func (gb *genBanger) ControlBang() []any {
	msgs := gb.gen.Generate()

	if gb.gen.Done() {
		gb.gen.Reset()
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

func NewTemplateBang(addresses []string, template template.Template) *Bang {
	return NewBang(newTemplateDestination(addresses, template))
}

func NewControlTemplate(template template.Template) *Bang {
	return NewBang(newTemplateDestination(nil, template))
}

func (d *templateDestination) ReceiveControlValue(value any, index int) {
	d.paramMap[fmt.Sprintf("controlInput%d", index)] = value
}

func (d *templateDestination) ReceiveMessage(msg any) []*muse.Message {
	d.paramMap = msg.(map[string]any)
	return nil
}

func (d *templateDestination) resolveParams() {
	var params []*template.Parameter
	for k, v := range d.paramMap {
		params = append(params, template.NewParameter(k, v))
	}

	d.template.SetParameters(params)
}

func (d *templateDestination) MessageBang(msg any) []*muse.Message {
	var allMessages []*muse.Message

	d.resolveParams()

	if d.template.Done() {
		d.template.Reset()
	}

	protoMessages := d.template.Generate()

	for _, address := range d.addresses {
		for _, message := range protoMessages {
			allMessages = append(allMessages, muse.NewMessage(address, message))
		}
	}

	return allMessages
}

func (d *templateDestination) ControlBang() []any {
	d.resolveParams()

	if d.template.Done() {
		d.template.Reset()
	}

	protoMessages := d.template.Generate()

	contents := make([]any, len(protoMessages))
	for i, msg := range protoMessages {
		contents[i] = msg
	}

	return contents
}

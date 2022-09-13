package prototype

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/banger"

	proto "github.com/almerlucke/muse/values/prototype"
)

type Prototype struct {
	addresses []string
	proto     proto.Prototype
}

func NewPrototype(addresses []string, proto proto.Prototype) *Prototype {
	return &Prototype{
		addresses: addresses,
		proto:     proto,
	}
}

func NewPrototypeGenerator(addresses []string, proto proto.Prototype, identifier string) *banger.Generator {
	return banger.NewGenerator(NewPrototype(addresses, proto), identifier)
}

func (p *Prototype) Bang() []*muse.Message {
	allMessages := []*muse.Message{}
	protoMessages := p.proto.Value()

	for _, address := range p.addresses {
		for _, message := range protoMessages {
			allMessages = append(allMessages, muse.NewMessage(address, message))
		}
	}

	return allMessages
}

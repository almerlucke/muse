package prototype

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers"
	"github.com/almerlucke/muse/values"
)

type Prototype struct {
	addresses []string
	proto     values.MapPrototype
}

func NewPrototype(addresses []string, proto values.MapPrototype) *Prototype {
	return &Prototype{
		addresses: addresses,
		proto:     proto,
	}
}

func NewPrototypeGenerator(addresses []string, proto values.MapPrototype, identifier string) *messengers.Generator {
	return messengers.NewGenerator(NewPrototype(addresses, proto), identifier)
}

func (p *Prototype) Bang() []*muse.Message {
	messages := make([]*muse.Message, len(p.addresses))
	message := p.proto.MapRaw()

	for index, address := range p.addresses {
		messages[index] = muse.NewMessage(address, message)
	}

	return messages
}

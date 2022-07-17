package prototype

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers"
	params "github.com/almerlucke/muse/parameters"
)

type Prototype struct {
	addresses []string
	proto     params.Prototype
}

func NewPrototype(addresses []string, proto params.Prototype) *Prototype {
	return &Prototype{
		addresses: addresses,
		proto:     proto,
	}
}

func NewPrototypeGenerator(addresses []string, proto params.Prototype, identifier string) *messengers.Generator {
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

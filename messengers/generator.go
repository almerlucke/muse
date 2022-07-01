package messengers

import "github.com/almerlucke/muse"

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

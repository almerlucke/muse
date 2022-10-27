package once

import "github.com/almerlucke/muse"

type Once struct {
	*muse.BaseMessenger
	bang      bool
	addresses []string
}

func NewOnce(addresses []string) *Once {
	return &Once{BaseMessenger: muse.NewBaseMessenger(""), addresses: addresses}
}

func NewControlOnce() *Once {
	return &Once{BaseMessenger: muse.NewBaseMessenger("")}
}

func (o *Once) Tick(int64, *muse.Configuration) {
	if !o.bang {
		o.SendControlValue(muse.Bang, 0)
		o.bang = true
	}
}

func (o *Once) Messages(timestamp int64, config *muse.Configuration) []*muse.Message {
	messages := []*muse.Message{}

	if !o.bang {
		o.SendControlValue(muse.Bang, 0)
		o.bang = true

		for _, address := range o.addresses {
			messages = append(messages, &muse.Message{
				Address: address,
				Content: muse.Bang,
			})
		}
	}

	return messages
}

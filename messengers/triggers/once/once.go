package once

import "github.com/almerlucke/muse"

type Once struct {
	*muse.BaseMessenger
	bang      bool
	addresses []string
}

func NewOnce(addresses []string) *Once {
	o := &Once{BaseMessenger: muse.NewBaseMessenger(""), addresses: addresses}
	o.SetSelf(o)
	return o
}

func NewControlOnce() *Once {
	return NewOnce(nil)
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

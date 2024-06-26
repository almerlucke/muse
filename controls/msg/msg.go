package msg

import "github.com/almerlucke/muse"

type Msg struct {
	*muse.BaseControl
	sender    muse.Patch
	addresses []string
}

func New(sender muse.Patch, addresses []string) *Msg {
	m := &Msg{BaseControl: muse.NewBaseControl(), sender: sender, addresses: addresses}
	m.SetSelf(m)
	return m
}

func (msg *Msg) ReceiveControlValue(value any, index int) {
	if index == 0 {
		for _, address := range msg.addresses {
			msg.sender.SendMessage(muse.NewMessage(address, value))
		}
	}
}

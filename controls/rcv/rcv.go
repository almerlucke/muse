package rcv

import "github.com/almerlucke/muse"

type Rcv struct {
	*muse.BaseControl
}

func NewRcv(id string) *Rcv {
	r := &Rcv{BaseControl: muse.NewBaseControl(id)}
	r.SetSelf(r)
	return r
}

func (r *Rcv) ReceiveMessage(msg any) []*muse.Message {
	r.SendControlValue(msg, 0)
	return nil
}

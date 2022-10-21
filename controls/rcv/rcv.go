package rcv

import "github.com/almerlucke/muse"

type Rcv struct {
	*muse.BaseControl
}

func NewRcv(id string) *Rcv {
	return &Rcv{BaseControl: muse.NewBaseControl(id)}
}

func (r *Rcv) ReceiveMessage(msg any) []*muse.Message {
	r.SendControlValue(msg, 0)
	return nil
}

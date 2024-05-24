package divider

import (
	"github.com/almerlucke/muse"
)

type Divider struct {
	*muse.BaseControl
	cnt int
	n   int
}

func New(n int) *Divider {
	d := &Divider{
		BaseControl: muse.NewBaseControl(),
		n:           n,
	}

	d.SetSelf(d)

	return d
}

func (d *Divider) ReceiveControlValue(value any, index int) {
	if index == 0 && muse.IsBang(value) {
		d.cnt++
		if d.cnt == d.n {
			d.SendControlValue(muse.Bang, 0)
			d.cnt = 0
		}
	}
}

package perform

import (
	"github.com/almerlucke/muse"
)

type Perform struct {
	*muse.BaseControl
	f func()
}

func New(f func()) *Perform {
	p := &Perform{
		BaseControl: muse.NewBaseControl(),
		f:           f,
	}

	p.SetSelf(p)

	return p
}

func (p *Perform) ReceiveControlValue(v any, i int) {
	if muse.IsBang(v) {
		p.f()
	}
}

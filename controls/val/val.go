package val

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/value"
)

type Val[T any] struct {
	*muse.BaseControl
	v value.Valuer[T]
}

func New[T any](v value.Valuer[T]) *Val[T] {
	val := &Val[T]{
		BaseControl: muse.NewBaseControl(),
		v:           v,
	}

	val.SetSelf(val)

	return val
}

func (val *Val[T]) ReceiveControlValue(value any, index int) {
	if index == 0 && value == muse.Bang {
		v := val.v.Value()

		if val.v.Finished() {
			val.SendControlValue(muse.Bang, 1)
		}

		val.SendControlValue(v, 0)
	} else if index == 1 && value == muse.Bang {
		val.v.Reset()
	}
}

package seq

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/value"
)

type Seq[T any] struct {
	*muse.BaseControl
	sequence *value.Sequence[T]
}

func New[T any](s *value.Sequence[T], id string) *Seq[T] {
	sq := &Seq[T]{
		BaseControl: muse.NewBaseControl(id),
		sequence:    s,
	}

	sq.SetSelf(sq)

	return sq
}

func (s *Seq[T]) ReceiveControlValue(value any, index int) {
	if index == 0 && value == muse.Bang {
		v := s.sequence.Value()

		if s.sequence.Finished() {
			s.SendControlValue(muse.Bang, 1)
		}

		s.SendControlValue(v, 0)
	} else if index == 1 && value == muse.Bang {
		s.sequence.Reset()
	} else if index == 2 && value == muse.Bang {
		s.sequence.Randomize()
	} else if index == 3 {
		if newSequence, ok := value.([]T); ok {
			s.sequence.Set(newSequence)
			s.sequence.Reset()
		}
	}
}

package sequencer

import (
	"github.com/almerlucke/muse"
)

type Sequencer struct {
	*muse.BaseMessenger
	sequence [][]*muse.Message
	index    int
}

func NewSequencer(sequence [][]*muse.Message, identifier string) *Sequencer {
	return &Sequencer{
		BaseMessenger: muse.NewBaseMessenger(identifier),
		sequence:      sequence,
	}
}

func (s *Sequencer) ReceiveMessage(msg any) []*muse.Message {
	bang, ok := msg.(string)
	if ok && bang == "bang" {
		msgs := s.sequence[s.index]

		s.index++
		if s.index >= len(s.sequence) {
			s.index = 0
		}

		return msgs
	}

	return nil
}

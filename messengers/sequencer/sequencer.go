package sequencer

import (
	"encoding/json"
	"os"

	"github.com/almerlucke/muse"
)

type Sequence [][]*muse.Message

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

func ReadSequence(file string) (Sequence, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var sequence Sequence

	err = json.Unmarshal(data, &sequence)
	if err != nil {
		return nil, err
	}

	return sequence, nil
}

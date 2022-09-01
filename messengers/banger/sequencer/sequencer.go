package sequencer

import (
	"encoding/json"
	"os"

	"github.com/almerlucke/muse"
)

type Sequence [][]*muse.Message

type Sequencer struct {
	sequence [][]*muse.Message
	index    int
}

func NewSequencer(sequence Sequence) *Sequencer {
	return &Sequencer{
		sequence: sequence,
	}
}

func NewSequencerWithFile(file string) (*Sequencer, error) {
	sequence, err := ReadSequence(file)
	if err != nil {
		return nil, err
	}

	return &Sequencer{
		sequence: sequence,
	}, nil
}

func (s *Sequencer) Bang() []*muse.Message {
	msgs := s.sequence[s.index]

	s.index++
	if s.index >= len(s.sequence) {
		s.index = 0
	}

	return msgs
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

package stepper

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/muse"
)

type Stepper struct {
	*muse.BaseMessenger
	addresses   []string
	accum       float64
	durationGen genny.Generator[float64]
}

func NewStepper(durationGen genny.Generator[float64], addresses []string) *Stepper {
	s := &Stepper{
		BaseMessenger: muse.NewBaseMessenger(),
		addresses:     addresses,
		durationGen:   durationGen,
	}

	s.SetSelf(s)

	return s
}

func (s *Stepper) Tick(timestamp int64, config *muse.Configuration) {
	_ = s.Messages(timestamp, config)
}

func (s *Stepper) Messages(timestamp int64, config *muse.Configuration) []*muse.Message {
	var (
		messages   []*muse.Message
		bang       bool
		durationMs float64
	)

	for {
		if float64(timestamp) < s.accum {
			break
		}

		if s.durationGen.Done() {
			s.SendControlValue(muse.Bang, 2)
			s.durationGen.Reset()
		}

		durationMs = s.durationGen.Generate()

		wait := config.MilliToSampsf(durationMs)
		if wait > 0 {
			bang = true
			s.accum += wait
		} else {
			s.accum -= wait
		}
	}

	if bang {
		s.SendControlValue(durationMs, 1)
		s.SendControlValue(muse.Bang, 0)

		for _, address := range s.addresses {
			messages = append(messages, &muse.Message{
				Address: address,
				Content: muse.Bang,
			})
		}
	}

	return messages
}

package stepper

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/muse"
)

type Stepper struct {
	*muse.BaseMessenger
	addresses []string
	accum     float64
	provider  genny.Generator[float64]
}

func NewStepper(provider genny.Generator[float64], addresses []string) *Stepper {
	s := &Stepper{
		BaseMessenger: muse.NewBaseMessenger(),
		addresses:     addresses,
		provider:      provider,
	}

	s.SetSelf(s)

	return s
}

func NewControlStepper(provider genny.Generator[float64]) *Stepper {
	return NewStepper(provider, nil)
}

func (s *Stepper) tick(timestamp int64, config *muse.Configuration) (bool, float64) {
	bang := false
	durationMs := 0.0

	for {
		if float64(timestamp) < s.accum {
			break
		}

		if s.provider.Done() {
			s.SendControlValue(muse.Bang, 2)
			s.provider.Reset()
		}

		durationMs = s.provider.Generate()

		wait := durationMs * 0.001 * config.SampleRate
		if wait > 0 {
			bang = true
			s.accum += wait
		} else {
			s.accum += -wait
		}
	}

	return bang, durationMs
}

func (s *Stepper) Tick(timestamp int64, config *muse.Configuration) {
	bang, duration := s.tick(timestamp, config)

	if bang {
		s.SendControlValue(duration, 1)
		s.SendControlValue(muse.Bang, 0)
	}
}

func (s *Stepper) Messages(timestamp int64, config *muse.Configuration) []*muse.Message {
	var messages []*muse.Message
	bang, duration := s.tick(timestamp, config)

	if bang {
		s.SendControlValue(duration, 1)
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

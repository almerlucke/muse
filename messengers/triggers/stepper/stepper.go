package stepper

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/ui"
)

type StepProvider interface {
	muse.Stater
	NextStep() float64
}

type Stepper struct {
	*muse.BaseMessenger
	addresses []string
	accum     int64
	provider  StepProvider
	listener  ui.Listener
}

func NewStepper(provider StepProvider, addresses []string, identifier string) *Stepper {
	return NewStepperWithListener(provider, addresses, nil, identifier)
}

func NewStepperWithListener(provider StepProvider, addresses []string, listener ui.Listener, identifier string) *Stepper {
	return &Stepper{
		BaseMessenger: muse.NewBaseMessenger(identifier),
		addresses:     addresses,
		provider:      provider,
		listener:      listener,
	}
}

func (s *Stepper) Messages(timestamp int64, config *muse.Configuration) []*muse.Message {
	messages := []*muse.Message{}
	bang := false

	for true {
		if timestamp < s.accum {
			break
		}

		if s.listener != nil {
			s.listener.Listen(s.provider.GetState())
		}

		wait := int64(s.provider.NextStep() * 0.001 * config.SampleRate)
		if wait > 0 {
			bang = true
			s.accum += wait
		} else {
			s.accum += -wait
		}
	}

	if bang {
		for _, address := range s.addresses {
			messages = append(messages, &muse.Message{
				Address: address,
				Content: "bang",
			})
		}
	}

	return messages
}

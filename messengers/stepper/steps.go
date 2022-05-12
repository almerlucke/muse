package stepper

import "github.com/almerlucke/muse"

type StepProvider interface {
	NextStep() float64
}

type SliceProvider struct {
	steps     []float64
	stepIndex int
}

func NewSliceProvider(steps []float64) *SliceProvider {
	return &SliceProvider{
		steps: steps,
	}
}

func (sp *SliceProvider) NextStep() float64 {
	step := sp.steps[sp.stepIndex]

	sp.stepIndex++

	if sp.stepIndex >= len(sp.steps) {
		sp.stepIndex = 0
	}

	return step
}

type Stepper struct {
	*muse.BaseMessenger
	addresses []string
	accum     int64
	provider  StepProvider
}

func NewStepper(provider StepProvider, addresses []string, identifier string) *Stepper {
	return &Stepper{
		BaseMessenger: muse.NewBaseMessenger(identifier),
		addresses:     addresses,
		provider:      provider,
	}
}

func (s *Stepper) Messages(timestamp int64, config *muse.Configuration) []*muse.Message {
	messages := []*muse.Message{}
	bang := false

	for true {
		if timestamp < s.accum {
			break
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

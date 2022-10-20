package stepper

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/value"
)

type StepProvider interface {
	muse.Stater
	NextStep() float64
}

type Stepper struct {
	*muse.BaseMessenger
	addresses []string
	accum     float64
	provider  StepProvider
}

func NewStepper(provider StepProvider, addresses []string, identifier string) *Stepper {
	return &Stepper{
		BaseMessenger: muse.NewBaseMessenger(identifier),
		addresses:     addresses,
		provider:      provider,
	}
}

func NewControlStepper(provider StepProvider, identifier string) *Stepper {
	return &Stepper{
		BaseMessenger: muse.NewBaseMessenger(identifier),
		addresses:     nil,
		provider:      provider,
	}
}

func (s *Stepper) tick(timestamp int64, config *muse.Configuration) (bool, float64) {
	bang := false
	durationMs := 0.0

	for true {
		if float64(timestamp) < s.accum {
			break
		}

		durationMs = s.provider.NextStep()

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
		s.SendControlValue(banger.Bang, 0)
	}
}

func (s *Stepper) Messages(timestamp int64, config *muse.Configuration) []*muse.Message {
	messages := []*muse.Message{}
	bang, duration := s.tick(timestamp, config)

	if bang {
		s.SendControlValue(duration, 1)
		s.SendControlValue(banger.Bang, 0)

		for _, address := range s.addresses {
			messages = append(messages, &muse.Message{
				Address: address,
				Content: banger.Bang,
			})
		}
	}

	return messages
}

type ValueStepper struct {
	value value.Valuer[float64]
}

func NewValueStepper(val value.Valuer[float64]) *ValueStepper {
	return &ValueStepper{
		value: val,
	}
}

func (vs *ValueStepper) NextStep() float64 {
	v := vs.value.Value()
	if vs.value.Finished() {
		vs.value.Reset()
	}
	return v
}

func (vs *ValueStepper) GetState() map[string]any {
	return vs.value.GetState()
}

func (vs *ValueStepper) SetState(state map[string]any) {
	vs.value.SetState(state)
}

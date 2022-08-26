package sequence

import (
	"github.com/almerlucke/muse/values"
)

type Sequence struct {
	sequence *values.Sequence[float64]
}

func New(sequence *values.Sequence[float64]) *Sequence {
	return &Sequence{
		sequence: sequence,
	}
}

func (sp *Sequence) NextStep() float64 {
	return sp.sequence.Value()
}

func (sp *Sequence) GetState() map[string]any {
	return sp.sequence.GetState()
}

func (sp *Sequence) SetState(state map[string]any) {
	sp.sequence.SetState(state)
}

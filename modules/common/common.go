package common

import "github.com/almerlucke/muse"

type Mult struct {
	*muse.BaseModule
}

func NewMult(numInputs int, config *muse.Configuration, identifier string) *Mult {
	return &Mult{
		BaseModule: muse.NewBaseModule(numInputs, 1, config, identifier),
	}
}

func (m *Mult) Synthesize() bool {
	if !m.BaseModule.Synthesize() {
		return false
	}

	out := m.Outputs[0].Buffer

	hasConnections := false
	for _, input := range m.Inputs {
		if input.IsConnected() {
			hasConnections = true
			break
		}
	}

	if hasConnections {
		for i := 0; i < m.Config.BufferSize; i++ {
			mult := 1.0

			for _, input := range m.Inputs {
				mult *= input.Buffer[i]
			}

			out[i] = mult
		}
	}

	return true
}

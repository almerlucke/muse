package mixer

import "github.com/almerlucke/muse"

type Mixer struct {
	*muse.BaseModule
	mix []float64
}

func NewMixer(numInputs int, config *muse.Configuration, identifier string) *Mixer {
	m := &Mixer{
		BaseModule: muse.NewBaseModule(numInputs, 1, config, identifier),
		mix:        make([]float64, numInputs),
	}

	for i := 0; i < numInputs; i++ {
		m.mix[i] = 0.0
	}

	return m
}

func (m *Mixer) MixAt(i int) float64 {
	return m.mix[i]
}

func (m *Mixer) Mix() []float64 {
	return m.mix
}

func (m *Mixer) SetMixAt(i int, mix float64) {
	m.mix[i] = mix
}

func (m *Mixer) SetMix(mix []float64) {
	m.mix = mix
}

func (m *Mixer) ReceiveControlValue(value any, index int) {
	m.SetMixAt(index, value.(float64))
}

func (m *Mixer) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	index := -1
	mix := 0.0

	if rawIndex, ok := content["index"]; ok {
		index = rawIndex.(int)
	}

	if rawMix, ok := content["mix"]; ok {
		mix = rawMix.(float64)
	}

	m.SetMixAt(index, mix)

	return nil
}

func (m *Mixer) Synthesize() bool {
	if !m.BaseModule.Synthesize() {
		return false
	}

	outBuf := m.Outputs[0].Buffer

	for i := 0; i < m.Config.BufferSize; i++ {
		outSamp := 0.0

		for j, in := range m.Inputs {
			if in.IsConnected() {
				outSamp += in.Buffer[i] * m.mix[j]
			}
		}

		outBuf[i] = outSamp
	}

	return true
}

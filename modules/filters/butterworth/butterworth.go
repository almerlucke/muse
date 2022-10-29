package butterworth

import (
	"github.com/almerlucke/muse"
	butterworthc "github.com/almerlucke/muse/components/filters/butterworth"
)

type Butterworth struct {
	*muse.BaseModule
	filter butterworthc.Butterworth
	fc     float64
	q      float64
}

func NewButterworth(fc float64, q float64, config *muse.Configuration, identifier string) *Butterworth {
	b := &Butterworth{
		BaseModule: muse.NewBaseModule(3, 1, config, identifier),
		fc:         fc,
		q:          q,
	}

	b.filter.Set(fc, q, config.SampleRate)

	b.SetSelf(b)

	return b
}

func (b *Butterworth) Frequency() float64 {
	return b.fc
}

func (b *Butterworth) SetFrequency(fc float64) {
	b.fc = fc
	b.filter.Set(fc, b.q, b.Config.SampleRate)
}

func (b *Butterworth) Resonance() float64 {
	return b.q
}

func (b *Butterworth) SetResonance(q float64) {
	b.q = q
	b.filter.Set(b.fc, b.q, b.Config.SampleRate)
}

func (b *Butterworth) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Cutoff Frequency
		b.SetFrequency(value.(float64))
	case 1: // Resonance
		b.SetResonance(value.(float64))
	}
}

func (b *Butterworth) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	if fc, ok := content["frequency"]; ok {
		b.SetFrequency(fc.(float64))
	}

	if res, ok := content["resonance"]; ok {
		b.SetResonance(res.(float64))
	}

	return nil
}

func (b *Butterworth) Synthesize() bool {
	if !b.BaseModule.Synthesize() {
		return false
	}

	in := b.Inputs[0].Buffer
	out := b.Outputs[0].Buffer

	for i := 0; i < b.Config.BufferSize; i++ {
		needUpdate := false
		fc := b.fc
		q := b.q

		if b.Inputs[1].IsConnected() {
			needUpdate = true
			fc = b.Inputs[1].Buffer[i]
		}

		if b.Inputs[2].IsConnected() {
			needUpdate = true
			q = b.Inputs[2].Buffer[i]
		}

		if needUpdate {
			b.filter.Set(fc, q, b.Config.SampleRate)
		}

		out[i] = b.filter.Process(in[i])
	}

	return true
}

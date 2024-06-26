package butterworth

import (
	"github.com/almerlucke/muse"
	butterworthc "github.com/almerlucke/muse/components/filters/butterworth"
	"github.com/almerlucke/muse/modules/filters"
)

type Butterworth struct {
	*muse.BaseModule
	filter butterworthc.Butterworth
	fc     float64
	q      float64
}

type Factory struct{}

func (f *Factory) New(cfg any) filters.Filter {
	fCfg := cfg.(*filters.FilterConfig)
	return New(fCfg.Frequency, fCfg.Resonance)
}

func DefaultConfig() *filters.FilterConfig {
	return filters.NewFilterConfig(1500.0, 0.06, 0.0, 0)
}

func New(fc float64, q float64) *Butterworth {
	b := &Butterworth{
		BaseModule: muse.NewBaseModule(3, 1),
		fc:         fc,
		q:          q,
	}

	b.filter.Set(fc, q, muse.SampleRate())

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

func (b *Butterworth) Drive() float64 { return 0.0 }

func (b *Butterworth) SetDrive(_ float64) {}

func (b *Butterworth) Type() int { return 0 }

func (b *Butterworth) SetType(_ int) {}

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

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

	return b
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

		out[i] = b.filter.Filter(in[i])
	}

	return true
}

package moog1

import (
	"github.com/almerlucke/muse"
	moog1c "github.com/almerlucke/muse/components/filters/moog/moog1"
)

type Moog1 struct {
	*muse.BaseModule
	moog       *moog1c.Moog1c
	fc         float64
	res        float64
	saturation float64
}

func NewMoog1(fc float64, res float64, saturation float64, config *muse.Configuration, identifier string) *Moog1 {
	return &Moog1{
		BaseModule: muse.NewBaseModule(4, 1, config, identifier),
		moog:       moog1c.NewMoog1c(fc, res, saturation, config.SampleRate),
		fc:         fc,
		res:        res,
		saturation: saturation,
	}
}

func (m *Moog1) Synthesize() bool {
	if !m.BaseModule.Synthesize() {
		return false
	}

	in := m.Inputs[0].Buffer
	out := m.Outputs[0].Buffer

	for i := 0; i < m.Config.BufferSize; i++ {
		needsUpdate := false
		fc := m.fc
		res := m.res
		saturation := m.saturation

		if m.Inputs[1].IsConnected() {
			fc = m.Inputs[1].Buffer[i]
			needsUpdate = true
		}

		if m.Inputs[2].IsConnected() {
			res = m.Inputs[2].Buffer[i]
			needsUpdate = true
		}

		if m.Inputs[3].IsConnected() {
			saturation = m.Inputs[3].Buffer[i]
			needsUpdate = true
		}

		if needsUpdate {
			m.moog.Set(fc, res, saturation, m.Config.SampleRate)
		}

		out[i] = m.moog.Process(in[i])
	}

	return true
}

package vartri

import "github.com/almerlucke/muse"

type VarTri struct {
	*muse.BaseModule
	phase float64
	delta float64
	w     float64
}

func NewVarTri(freq float64, phase float64, w float64, config *muse.Configuration, identifier string) *VarTri {
	return &VarTri{
		BaseModule: muse.NewBaseModule(3, 1, config, identifier),
		phase:      phase,
		delta:      freq / config.SampleRate,
		w:          w,
	}
}

func (vt *VarTri) ReceiveMessage(msg any) []*muse.Message {
	params, ok := msg.(map[string]any)
	if ok {
		f, ok := params["frequency"]
		if ok {
			vt.delta = f.(float64) / vt.Config.SampleRate
		}
		w, ok := params["dutyWidth"]
		if ok {
			vt.w = w.(float64)
			if vt.w > 1.0 {
				vt.w = 1.0
			}
			if vt.w < 0.0 {
				vt.w = 0.0
			}
		}
	}

	return nil
}

func (vt *VarTri) Synthesize() bool {
	if !vt.BaseModule.Synthesize() {
		return false
	}

	out := vt.Outputs[0].Buffer

	for i := 0; i < vt.Config.BufferSize; i++ {
		phase := vt.phase

		if vt.Inputs[1].IsConnected() {
			phase = vt.phase + vt.Inputs[1].Buffer[i]

			for phase >= 1.0 {
				phase -= 1.0
			}

			for phase < 0.0 {
				phase += 1.0
			}
		}

		if vt.Inputs[2].IsConnected() {
			vt.w = vt.Inputs[2].Buffer[i]
			if vt.w > 1.0 {
				vt.w = 1.0
			}

			if vt.w < 0.0 {
				vt.w = 0.0
			}
		}

		g1 := 2.0*phase - 1.0
		g2 := phase - vt.w

		if g2 < 0.0 {
			g2 += 1.0
		}

		g2 = 2.0*g2 - 1.0

		if vt.Inputs[0].IsConnected() {
			vt.delta = vt.Inputs[0].Buffer[i] / vt.Config.SampleRate
		}

		vt.phase += vt.delta

		for vt.phase >= 1.0 {
			vt.phase -= 1.0
		}

		for vt.phase < 0.0 {
			vt.phase += 1.0
		}

		out[i] = 1.0/(8.0*(vt.w-vt.w*vt.w))*(g1*g1-g2*g2) + 0.5
	}

	return true
}

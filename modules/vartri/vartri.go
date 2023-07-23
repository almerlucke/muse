package vartri

import "github.com/almerlucke/muse"

type VarTri struct {
	*muse.BaseModule
	phase float64
	delta float64
	w     float64
	fc    float64
}

func New(freq float64, phase float64, w float64) *VarTri {
	v := &VarTri{
		BaseModule: muse.NewBaseModule(3, 1),
		phase:      phase,
		delta:      freq / muse.SampleRate(),
		w:          w,
	}

	v.SetSelf(v)

	return v
}

func (vt *VarTri) DutyWidth() float64 {
	return vt.w
}

func (vt *VarTri) SetDutyWidth(w float64) {
	vt.w = w
}

func (vt *VarTri) Phase() float64 {
	return vt.phase
}

func (vt *VarTri) SetPhase(ph float64) {
	vt.phase = ph
}

func (vt *VarTri) Frequency() float64 {
	return vt.fc
}

func (vt *VarTri) SetFrequency(fc float64) {
	vt.delta = fc / vt.Config.SampleRate
	vt.fc = fc
}

func (vt *VarTri) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Frequency
		vt.SetFrequency(value.(float64))
	case 1: // Phase
		vt.SetPhase(value.(float64))
	case 2: // Duty Width
		vt.SetDutyWidth(value.(float64))
	}
}

func (vt *VarTri) ReceiveMessage(msg any) []*muse.Message {
	params, ok := msg.(map[string]any)
	if ok {
		p, ok := params["phase"]
		if ok {
			vt.SetPhase(p.(float64))
		}
		fc, ok := params["frequency"]
		if ok {
			vt.SetFrequency(fc.(float64))
		}
		w, ok := params["dutyWidth"]
		if ok {
			vt.SetDutyWidth(w.(float64))
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

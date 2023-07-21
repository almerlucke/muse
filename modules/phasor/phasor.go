package phasor

import "github.com/almerlucke/muse"

type Phasor struct {
	*muse.BaseModule
	phase float64
	delta float64
	fc    float64
}

func NewPhasor(freq float64, phase float64, config *muse.Configuration) *Phasor {
	p := &Phasor{
		BaseModule: muse.NewBaseModule(2, 1, config, ""),
		phase:      phase,
		delta:      freq / config.SampleRate,
		fc:         freq,
	}

	p.SetSelf(p)

	return p
}

func (p *Phasor) Phase() float64 {
	return p.phase
}

func (p *Phasor) SetPhase(ph float64) {
	p.phase = ph
}

func (p *Phasor) Frequency() float64 {
	return p.fc
}

func (p *Phasor) SetFrequency(fc float64) {
	p.delta = fc / p.Config.SampleRate
	p.fc = fc
}

func (p *Phasor) ReceiveControlValue(value any, index int) {
	switch index {
	case 0:
		p.SetFrequency(value.(float64))
	case 1:
		p.SetPhase(value.(float64))
	}
}

func (p *Phasor) ReceiveMessage(msg any) []*muse.Message {
	params, ok := msg.(map[string]any)
	if ok {
		fc, ok := params["frequency"]
		if ok {
			p.SetFrequency(fc.(float64))
		}
		phase, ok := params["phase"]
		if ok {
			p.SetPhase(phase.(float64))
		}
	}

	return nil
}

func (p *Phasor) Synthesize() bool {
	if !p.BaseModule.Synthesize() {
		return false
	}

	out := p.Outputs[0].Buffer

	for i := 0; i < p.Config.BufferSize; i++ {
		phase := p.phase

		if p.Inputs[1].IsConnected() {
			phase = p.phase + p.Inputs[1].Buffer[i]

			for phase >= 1.0 {
				phase -= 1.0
			}

			for phase < 0.0 {
				phase += 1.0
			}
		}

		out[i] = phase

		if p.Inputs[0].IsConnected() {
			p.delta = p.Inputs[0].Buffer[i] / p.Config.SampleRate
		}

		p.phase += p.delta

		for p.phase >= 1.0 {
			p.phase -= 1.0
		}

		for p.phase < 0.0 {
			p.phase += 1.0
		}
	}

	return true
}

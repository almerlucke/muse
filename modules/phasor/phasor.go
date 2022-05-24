package phasor

import "github.com/almerlucke/muse"

type Phasor struct {
	*muse.BaseModule
	phase float64
	delta float64
}

func NewPhasor(freq float64, phase float64, config *muse.Configuration, identifier string) *Phasor {
	return &Phasor{
		BaseModule: muse.NewBaseModule(2, 1, config, identifier),
		phase:      phase,
		delta:      freq / config.SampleRate,
	}
}

func (p *Phasor) ReceiveMessage(msg any) []*muse.Message {
	params, ok := msg.(map[string]any)
	if ok {
		f, ok := params["frequency"]
		if ok {
			p.delta = f.(float64) / p.Config.SampleRate
		}
		phase, ok := params["phase"]
		if ok {
			p.phase = phase.(float64)

			for p.phase >= 1.0 {
				p.phase -= 1.0
			}

			for p.phase < 0.0 {
				p.phase += 1.0
			}
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

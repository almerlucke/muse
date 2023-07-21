package pan

import (
	"math"

	"github.com/almerlucke/muse"
)

var halfPi = math.Pi / 2.0

type StereoPan struct {
	*muse.BaseModule
	pan float64
}

func NewStereoPan(pan float64, config *muse.Configuration) *StereoPan {
	p := &StereoPan{
		BaseModule: muse.NewBaseModule(2, 2, config, ""),
		pan:        pan,
	}

	p.SetSelf(p)

	return p
}

func (p *StereoPan) Pan() float64 {
	return p.pan
}

func (p *StereoPan) SetPan(pan float64) {
	p.pan = pan
}

func (p *StereoPan) ReceiveControlValue(value any, index int) {
	if index == 0 {
		p.SetPan(value.(float64))
	}
}

func (p *StereoPan) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	if pan, ok := content["pan"]; ok {
		p.SetPan(pan.(float64))
	}

	return nil
}

func (p *StereoPan) Synthesize() bool {
	if !p.BaseModule.Synthesize() {
		return false
	}

	pan := p.pan

	for i := 0; i < p.Config.BufferSize; i++ {
		if p.Inputs[1].IsConnected() {
			pan = p.Inputs[1].Buffer[i]
		}

		inSamp := p.Inputs[0].Buffer[i]
		panLookup := pan * halfPi

		p.Outputs[0].Buffer[i] = inSamp * math.Cos(panLookup)
		p.Outputs[1].Buffer[i] = inSamp * math.Sin(panLookup)
	}

	return true
}

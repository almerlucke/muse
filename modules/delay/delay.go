package delay

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/delayc"
)

type Delay struct {
	*muse.BaseModule
	delay        *delayc.Delay
	readLocation float64
}

func NewDelay(length float64, location float64, config *muse.Configuration, identifier string) *Delay {
	return &Delay{
		BaseModule:   muse.NewBaseModule(2, 1, config, identifier),
		delay:        delayc.NewDelay(int(length * config.SampleRate * 0.001)),
		readLocation: location * config.SampleRate * 0.001,
	}
}

func (d *Delay) Synthesize() bool {
	in := d.Inputs[0].Buffer
	out := d.Outputs[0].Buffer

	for i := 0; i < d.Config.BufferSize; i++ {
		readLocation := d.readLocation
		if d.Inputs[1].IsConnected() {
			readLocation = d.Inputs[1].Buffer[i] * d.Config.SampleRate * 0.001
		}
		out[i] = d.delay.Read(readLocation)
		d.delay.Write(in[i])
	}

	return true
}

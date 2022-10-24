package delay

import (
	"github.com/almerlucke/muse"
	delayc "github.com/almerlucke/muse/components/delay"
)

type Delay struct {
	*muse.BaseModule
	delay        *delayc.Delay
	readLocation float64
	readLocMS    float64
}

func NewDelay(length float64, location float64, config *muse.Configuration, identifier string) *Delay {
	return &Delay{
		BaseModule:   muse.NewBaseModule(2, 1, config, identifier),
		delay:        delayc.NewDelay(int(length * config.SampleRate * 0.001)),
		readLocation: location * config.SampleRate * 0.001,
		readLocMS:    location,
	}
}

func (d *Delay) ReadLocationMS() float64 {
	return d.readLocMS
}

func (d *Delay) SetReadLocationMS(readLocMS float64) {
	d.readLocMS = readLocMS
	d.readLocation = readLocMS * d.Config.SampleRate * 0.001
}

func (d *Delay) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Read Location
		d.SetReadLocationMS(value.(float64))
	}
}

func (d *Delay) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	if readLocMS, ok := content["location"]; ok {
		d.SetReadLocationMS(readLocMS.(float64))
	}

	return nil
}

func (d *Delay) Synthesize() bool {
	if !d.BaseModule.Synthesize() {
		return false
	}
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

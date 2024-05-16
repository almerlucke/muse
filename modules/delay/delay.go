package delay

import (
	"github.com/almerlucke/muse"
	delayc "github.com/almerlucke/muse/components/delay"
)

type Delay struct {
	*muse.BaseModule
	delay          *delayc.Delay
	readLocation   float64
	readLocationMS float64
}

func New(length float64, location float64) *Delay {
	sr := muse.SampleRate()
	d := &Delay{
		BaseModule:     muse.NewBaseModule(2, 1),
		delay:          delayc.New(int(length * sr * 0.001)),
		readLocation:   location * sr * 0.001,
		readLocationMS: location,
	}

	d.SetSelf(d)

	return d
}

func (d *Delay) ReadLocation() float64 {
	return d.readLocationMS
}

func (d *Delay) SetReadLocation(readLocation float64) {
	d.readLocationMS = readLocation
	d.readLocation = readLocation * d.Config.SampleRate * 0.001
}

func (d *Delay) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Read Location
		d.SetReadLocation(value.(float64))
	}
}

func (d *Delay) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	if readLocation, ok := content["location"]; ok {
		d.SetReadLocation(readLocation.(float64))
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
		if d.Inputs[1].IsConnected() {
			d.SetReadLocation(d.Inputs[1].Buffer[i])
		}

		out[i] = d.delay.ReadHermite(d.readLocation)

		d.delay.Write(in[i])
	}

	return true
}

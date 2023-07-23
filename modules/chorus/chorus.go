package chorus

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/buffer"
	"github.com/almerlucke/muse/components/delay"
	"github.com/almerlucke/muse/components/phasor"
	"github.com/almerlucke/muse/components/waveshaping"
)

/*
flanger : delay from 0.01 to 5 ms
chorus : delay from 5 to 25 ms
doubler : delay from 25 to 75 ms
echo : delay from 75 to 1000 ms (and beyond)

speed : 0 - 20
*/

var defaultModTable = waveshaping.NewSineTable(512)

const (
	mod2SpeedDiv = 2.0
	mod3SpeedDiv = 3.0
	mod4SpeedDiv = 5.0
	mod2Phase    = 0.5
	mod3Phase    = 0.0
	mod4Phase    = 0.5
)

type Chorus struct {
	muse.BaseModule
	delayLine   *delay.Delay
	delayCenter float64
	delayRange  float64
	modShaper   waveshaping.Shaper
	mods        [4]*phasor.Phasor
	modDepth    float64
	modSpeed    float64
	modRange    float64
	mix         float64
}

func New(stereo bool, delayCenter float64, delayRange float64, modDepth float64, modSpeed float64, mix float64, modShaper waveshaping.Shaper) *Chorus {
	numOutputs := 1
	if stereo {
		numOutputs = 2
	}

	c := &Chorus{
		BaseModule:  *muse.NewBaseModule(4, numOutputs),
		delayLine:   delay.New(int((delayCenter + delayRange*0.5 + 1) * muse.SampleRate() * 0.001)),
		delayCenter: delayCenter,
		delayRange:  delayRange,
		modShaper:   modShaper,
		modDepth:    modDepth,
		modSpeed:    modSpeed,
		modRange:    modDepth * delayRange * 0.5,
		mix:         mix,
	}

	if modShaper == nil {
		c.modShaper = defaultModTable
	}

	speed := [4]float64{modSpeed, modSpeed / mod2SpeedDiv, modSpeed / mod3SpeedDiv, modSpeed / mod4SpeedDiv}
	phase := [4]float64{0, mod2Phase, mod3Phase, mod4Phase}

	for i := 0; i < 4; i++ {
		c.mods[i] = phasor.New(speed[i], muse.SampleRate(), phase[i])
	}

	c.SetSelf(c)

	return c
}

func (c *Chorus) ModSpeed() float64 {
	return c.modSpeed
}

func (c *Chorus) SetModSpeed(modSpeed float64) {
	c.modSpeed = modSpeed
	c.mods[0].SetFrequency(c.modSpeed, c.Config.SampleRate)
	c.mods[1].SetFrequency(c.modSpeed/mod2SpeedDiv, c.Config.SampleRate)
	c.mods[2].SetFrequency(c.modSpeed/mod3SpeedDiv, c.Config.SampleRate)
	c.mods[3].SetFrequency(c.modSpeed/mod4SpeedDiv, c.Config.SampleRate)
}

func (c *Chorus) ModDepth() float64 {
	return c.modDepth
}

func (c *Chorus) SetModDepth(modDepth float64) {
	c.modDepth = modDepth
	c.modRange = modDepth * c.delayRange * 0.5
}

func (c *Chorus) Mix() float64 {
	return c.mix
}

func (c *Chorus) SetMix(mix float64) {
	c.mix = mix
}

func (c *Chorus) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // ModSpeed
		c.SetModSpeed(value.(float64))
	case 1: // ModDepth
		c.SetModDepth(value.(float64))
	case 2: // Mix
		c.SetMix(value.(float64))
	}
}

func (c *Chorus) ReceiveMessage(msg any) []*muse.Message {
	m := msg.(map[string]any)

	if modSpeed, ok := m["modSpeed"].(float64); ok {
		c.SetModSpeed(modSpeed)
	}

	if modDepth, ok := m["modDepth"].(float64); ok {
		c.SetModDepth(modDepth)
	}

	if mix, ok := m["mix"].(float64); ok {
		c.SetMix(mix)
	}

	return nil
}

func (c *Chorus) Synthesize() bool {
	if !c.BaseModule.Synthesize() {
		return false
	}

	in := c.Inputs[0].Buffer
	outLeft := c.Outputs[0].Buffer

	var outRight buffer.Buffer

	stereo := len(c.Outputs) == 2

	if stereo {
		outRight = c.Outputs[1].Buffer
	}

	msSamps := c.Config.SampleRate * 0.001

	for i := 0; i < c.Config.BufferSize; i++ {
		if c.Inputs[1].IsConnected() {
			c.SetModDepth(c.Inputs[1].Buffer[i])
		}

		if c.Inputs[2].IsConnected() {
			c.SetModSpeed(c.Inputs[2].Buffer[i])
		}

		if c.Inputs[3].IsConnected() {
			c.SetMix(c.Inputs[3].Buffer[i])
		}

		d1 := c.delayLine.Read(msSamps * (c.delayCenter + c.modRange*c.modShaper.Shape(c.mods[0].Tick()[0])))
		d2 := c.delayLine.Read(msSamps * (c.delayCenter + c.modRange*c.modShaper.Shape(c.mods[1].Tick()[0])))
		d3 := c.delayLine.Read(msSamps * (c.delayCenter + c.modRange*c.modShaper.Shape(c.mods[2].Tick()[0])))
		d4 := c.delayLine.Read(msSamps * (c.delayCenter + c.modRange*c.modShaper.Shape(c.mods[3].Tick()[0])))

		c.delayLine.Write(in[i])

		if stereo {
			outLeft[i] = in[i]*(1.0-c.mix) + c.mix*(d1+d3)
			outRight[i] = in[i]*(1.0-c.mix) + c.mix*(d2+d4)
		} else {
			outLeft[i] = in[i]*(1.0-c.mix) + c.mix*(d1+d2+d3+d4)
		}
	}

	return true
}

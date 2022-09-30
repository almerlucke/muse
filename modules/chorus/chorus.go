package chorus

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/delay"
	"github.com/almerlucke/muse/components/phasor"
	"github.com/almerlucke/muse/components/waveshaping"
)

/*
flanger : delay from 0.01 to 5 ms
chorus : delay from 5 to 25 ms
doubler : delay from 25 to 75 ms
echo : delay from 75 to 1000 ms (and beyond)

speed : 0 -20
*/

type Chorus struct {
	muse.BaseModule
	modShaper   waveshaping.Shaper
	delayLine   *delay.Delay
	mods        [4]*phasor.Phasor
	delayCenter float64
	delayRange  float64
	modDepth    float64
	modSpeed    float64
	mix         float64
}

func NewChorus(stereo bool, delayCenter float64, delayRange float64, modDepth float64, modSpeed float64, mix float64, modShaper waveshaping.Shaper, config *muse.Configuration, identifier string) *Chorus {
	numOutputs := 1
	if stereo {
		numOutputs = 2
	}
	c := &Chorus{
		BaseModule:  *muse.NewBaseModule(1, numOutputs, config, identifier),
		modShaper:   modShaper,
		delayLine:   delay.NewDelay(int((delayCenter + delayRange*0.5 + 1) * config.SampleRate * 0.001)),
		delayCenter: delayCenter,
		delayRange:  delayRange,
		modDepth:    modDepth,
		modSpeed:    modSpeed,
		mix:         mix,
	}

	speed := [4]float64{modSpeed, modSpeed / 2.0, modSpeed / 3.0, modSpeed / 5.0}
	phase := [4]float64{0, 0.5, 0, 0.5}

	for i := 0; i < 4; i++ {
		c.mods[i] = phasor.NewPhasor(speed[i], config.SampleRate, phase[i])
	}

	return c
}

func (c *Chorus) Synthesize() bool {
	if !c.BaseModule.Synthesize() {
		return false
	}

	msSamps := c.Config.SampleRate * 0.001

	in := c.Inputs[0].Buffer
	stereo := len(c.Outputs) == 2

	if stereo {
		outLeft := c.Outputs[0].Buffer
		outRight := c.Outputs[1].Buffer

		for i := 0; i < c.Config.BufferSize; i++ {
			d1 := c.delayLine.Read(msSamps * (c.delayCenter + c.delayRange*0.5*c.modDepth*c.modShaper.Shape(c.mods[0].Tick())))
			d2 := c.delayLine.Read(msSamps * (c.delayCenter + c.delayRange*0.5*c.modDepth*c.modShaper.Shape(c.mods[1].Tick())))
			d3 := c.delayLine.Read(msSamps * (c.delayCenter + c.delayRange*0.5*c.modDepth*c.modShaper.Shape(c.mods[2].Tick())))
			d4 := c.delayLine.Read(msSamps * (c.delayCenter + c.delayRange*0.5*c.modDepth*c.modShaper.Shape(c.mods[3].Tick())))

			c.delayLine.Write(in[i])

			outLeft[i] = in[i]*(1.0-c.mix) + c.mix*(d1+d3)
			outRight[i] = in[i]*(1.0-c.mix) + c.mix*(d2+d4)
		}
	} else {
		out := c.Outputs[0].Buffer

		for i := 0; i < c.Config.BufferSize; i++ {
			d1 := c.delayLine.Read(msSamps * (c.delayCenter + c.delayRange*0.5*c.modDepth*c.modShaper.Shape(c.mods[0].Tick())))
			d2 := c.delayLine.Read(msSamps * (c.delayCenter + c.delayRange*0.5*c.modDepth*c.modShaper.Shape(c.mods[1].Tick())))
			d3 := c.delayLine.Read(msSamps * (c.delayCenter + c.delayRange*0.5*c.modDepth*c.modShaper.Shape(c.mods[2].Tick())))
			d4 := c.delayLine.Read(msSamps * (c.delayCenter + c.delayRange*0.5*c.modDepth*c.modShaper.Shape(c.mods[3].Tick())))

			c.delayLine.Write(in[i])

			out[i] = in[i]*(1.0-c.mix) + c.mix*(d1+d2+d3+d4)
		}
	}

	return true
}

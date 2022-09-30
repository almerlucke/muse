package chorus

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/delay"
	"github.com/almerlucke/muse/components/phasor"
	"github.com/almerlucke/muse/components/waveshaping"
)

/*

baseDelay * 1.414 + 2


Chorus delay: milliseconds 20 - 50
Chorus speed: frequency

Fc, at 1/2 Fc, 1/4 Fc and at 1/8 F

Delay between 20 / 50 milliseconds


msSamps = fs / 1000.0
d1 = delay.Read(msSamps * (35 + 15 * modDepth * mod1.Tick()))
d2 = delay.Read(msSamps * (35 + 15 * modDepth * mod2.Tick()))
d3 = delay.Read(msSamps * (35 + 15 * modDepth * mod3.Tick()))
d4 = delay.Read(msSamps * (35 + 15 * modDepth * mod4.Tick()))



phase1 = 0
phase2 = 0.2497
phase3 = 0.5001
phase4 = 0.7493

fc1 = fc
fc2 = fc * 0.5001
fc3 = fc * 0.2499
fc4 = fc * 0.1241

35 + 15 * mod(-1 / 1) * modDepth(0 / 1)


*/

type Chorus struct {
	muse.BaseModule
	modShaper waveshaping.Shaper
	delayLine *delay.Delay
	mods      [4]*phasor.Phasor
	modDepth  float64
	modSpeed  float64
	mix       float64
}

func NewChorus(modShaper waveshaping.Shaper, modDepth float64, modSpeed float64, mix float64, config *muse.Configuration, identifier string) *Chorus {
	c := &Chorus{
		BaseModule: *muse.NewBaseModule(1, 1, config, identifier),
		modShaper:  modShaper,
		delayLine:  delay.NewDelay(int(51 * config.SampleRate * 0.001)),
		modDepth:   modDepth,
		modSpeed:   modSpeed,
		mix:        mix,
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
	out := c.Outputs[0].Buffer

	for i := 0; i < c.Config.BufferSize; i++ {
		d1 := c.delayLine.Read(msSamps * (35 + 15*c.modDepth*c.modShaper.Shape(c.mods[0].Tick())))
		d2 := c.delayLine.Read(msSamps * (35 + 15*c.modDepth*c.modShaper.Shape(c.mods[1].Tick())))
		d3 := c.delayLine.Read(msSamps * (35 + 15*c.modDepth*c.modShaper.Shape(c.mods[2].Tick())))
		d4 := c.delayLine.Read(msSamps * (35 + 15*c.modDepth*c.modShaper.Shape(c.mods[3].Tick())))

		c.delayLine.Write(in[i])

		out[i] = in[i]*(1.0-c.mix) + c.mix*(d1+d2+d3+d4)
	}

	return true
}

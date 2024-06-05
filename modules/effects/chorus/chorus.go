package chorus

import (
	"github.com/almerlucke/genny/float/phasor"
	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/float/shape/shapers/lookup"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/delay"
	"github.com/almerlucke/muse/utils/mmath"
	"github.com/almerlucke/muse/utils/timing"
	"math"
)

/*
flanger : delay from 0.01 to 5 ms
chorus : delay from 5 to 25 ms
doubler : delay from 25 to 75 ms
echo : delay from 75 to 1000 ms (and beyond)

speed : 0 - 20
*/

var defaultModTable = lookup.NewSineTable(512)

const (
	_maxDelay     = 50.0
	_minDelay     = 5.01
	_amountRange  = 20.0
	_mod2SpeedDiv = 2.0
	_mod3SpeedDiv = 3.0
	_mod4SpeedDiv = 5.0
	_mod2Phase    = 0.5
	_mod3Phase    = 0.0
	_mod4Phase    = 0.5
)

type lpFilter struct {
	cf float64
	x1 float64
	y1 float64
}

func (f *lpFilter) set(fc float64, sr float64) {
	f.cf = math.Tan(math.Pi * fc / muse.SampleRate())
}

func (f *lpFilter) filter(x float64) float64 {
	out := f.cf*x + f.cf*f.x1 - (f.cf-1.0)*f.y1
	out /= f.cf + 1.0
	f.y1 = out
	f.x1 = x
	return out
}

type Chorus struct {
	muse.BaseModule
	delayLineLeft  *delay.Delay
	delayLineRight *delay.Delay
	delayCenter    float64
	delayRange     float64
	modShaper      shape.Shaper
	mods           [4]*phasor.Phasor
	amount         float64 // 0 - 1
	delay          float64 // 0 - 1
	rate           float64
	width          float64
	mix            float64
	fb             float64
	lp1            lpFilter
	lp2            lpFilter
}

func New(rate float64, amount float64, delayAmount float64, feedback float64, width float64, mix float64, modShaper shape.Shaper) *Chorus {
	delayLengthSamps := int(timing.MilliToSamps(_maxDelay+_amountRange, muse.SampleRate()) + 1)

	c := &Chorus{
		BaseModule:     *muse.NewBaseModule(2, 2),
		delayLineLeft:  delay.New(delayLengthSamps),
		delayLineRight: delay.New(delayLengthSamps),
		modShaper:      modShaper,
		amount:         amount,
		delay:          delayAmount,
		rate:           rate,
		mix:            mix,
		width:          width,
		fb:             feedback,
	}

	c.lp1.set(2000.0, muse.SampleRate())
	c.lp2.set(2000.0, muse.SampleRate())
	c.updateCalculations()

	if modShaper == nil {
		c.modShaper = defaultModTable
	}

	speed := [4]float64{rate, rate / _mod2SpeedDiv, rate / _mod3SpeedDiv, rate / _mod4SpeedDiv}
	phase := [4]float64{0, _mod2Phase, _mod3Phase, _mod4Phase}

	for i := 0; i < 4; i++ {
		c.mods[i] = phasor.New(speed[i], muse.SampleRate(), phase[i])
	}

	c.SetSelf(c)

	return c
}

func (c *Chorus) updateCalculations() {
	c.delayCenter, c.delayRange = c.delay*(_maxDelay-_minDelay)+_minDelay, c.amount*_amountRange
}

func (c *Chorus) Rate() float64 {
	return c.rate
}

func (c *Chorus) SetRate(rate float64) {
	rate = mmath.Limit(rate, 0, 1)
	c.rate = rate
	c.mods[0].SetFrequency(rate, c.Config.SampleRate)
	c.mods[1].SetFrequency(rate/_mod2SpeedDiv, c.Config.SampleRate)
	c.mods[2].SetFrequency(rate/_mod3SpeedDiv, c.Config.SampleRate)
	c.mods[3].SetFrequency(rate/_mod4SpeedDiv, c.Config.SampleRate)
}

func (c *Chorus) Amount() float64 {
	return c.amount
}

func (c *Chorus) SetAmount(amount float64) {
	amount = mmath.Limit(amount, 0, 1)
	c.amount = amount
	c.updateCalculations()
}

func (c *Chorus) Delay() float64 {
	return c.delay
}

func (c *Chorus) SetDelay(delayAmount float64) {
	delayAmount = mmath.Limit(delayAmount, 0, 1)
	c.delay = delayAmount
	c.updateCalculations()
}

func (c *Chorus) Mix() float64 {
	return c.mix
}

func (c *Chorus) SetMix(mix float64) {
	mix = mmath.Limit(mix, 0, 1)
	c.mix = mix
}

func (c *Chorus) Feedback() float64 {
	return c.fb
}

func (c *Chorus) SetFeedback(fb float64) {
	fb = mmath.Limit(fb, 0, 1)
	c.fb = fb
}

func (c *Chorus) Width() float64 {
	return c.width
}

func (c *Chorus) SetWidth(w float64) {
	w = mmath.Limit(w, 0, 1)
	c.width = w
}

func (c *Chorus) ReceiveControlValue(value any, index int) {
	switch index {
	case 0:
		c.SetRate(value.(float64))
	case 1:
		c.SetAmount(value.(float64))
	case 2:
		c.SetDelay(value.(float64))
	case 3:
		c.SetFeedback(value.(float64))
	case 4:
		c.SetWidth(value.(float64))
	case 5:
		c.SetMix(value.(float64))
	}
}

func (c *Chorus) ReceiveMessage(msg any) []*muse.Message {
	m := msg.(map[string]any)

	if rate, ok := m["rate"].(float64); ok {
		c.SetRate(rate)
	}

	if amount, ok := m["amount"].(float64); ok {
		c.SetAmount(amount)
	}

	if d, ok := m["delay"].(float64); ok {
		c.SetDelay(d)
	}

	if fb, ok := m["feedback"].(float64); ok {
		c.SetFeedback(fb)
	}

	if w, ok := m["width"].(float64); ok {
		c.SetWidth(w)
	}

	if mix, ok := m["mix"].(float64); ok {
		c.SetMix(mix)
	}

	return nil
}

func (c *Chorus) synthesizeStereoInput() {
	inLeft := c.Inputs[0].Buffer
	inRight := c.Inputs[1].Buffer
	outLeft := c.Outputs[0].Buffer
	outRight := c.Outputs[1].Buffer

	msSamps := c.Config.SampleRate * 0.001

	for i := 0; i < c.Config.BufferSize; i++ {
		d1Loc := math.Abs(msSamps * (c.delayCenter + c.delayRange*c.modShaper.Shape(c.mods[0].Generate())))
		d2Loc := math.Abs(msSamps * (c.delayCenter + c.delayRange*c.modShaper.Shape(c.mods[1].Generate())))
		d3Loc := math.Abs(msSamps * (c.delayCenter + c.delayRange*c.modShaper.Shape(c.mods[2].Generate())))
		d4Loc := math.Abs(msSamps * (c.delayCenter + c.delayRange*c.modShaper.Shape(c.mods[3].Generate())))

		d1Left := c.delayLineLeft.ReadLinear(d1Loc)
		d2Left := c.delayLineLeft.ReadLinear(d2Loc)
		d3Left := c.delayLineLeft.ReadLinear(d3Loc)
		d4Left := c.delayLineLeft.ReadLinear(d4Loc)

		d1Right := c.delayLineRight.ReadLinear(d1Loc)
		d2Right := c.delayLineRight.ReadLinear(d2Loc)
		d3Right := c.delayLineRight.ReadLinear(d3Loc)
		d4Right := c.delayLineRight.ReadLinear(d4Loc)

		im1Left := d1Left + d3Left
		im2Left := d2Left + d4Left

		im1Right := d1Right + d3Right
		im2Right := d2Right + d4Right

		c.delayLineLeft.Write(inLeft[i] + c.fb*c.lp1.filter((im1Left+im2Left)*0.25))
		c.delayLineRight.Write(inRight[i] + c.fb*c.lp2.filter((im1Right+im2Right)*0.25))

		out1 := im1Left + im2Left*(1.0-c.width)
		out2 := im1Left*(1.0-c.width) + im2Left
		out1 += im1Right + im2Right*(1.0-c.width)
		out2 += im1Right*(1.0-c.width) + im2Right

		outLeft[i] = inLeft[i]*(1.0-c.mix) + c.mix*out1
		outRight[i] = inRight[i]*(1.0-c.mix) + c.mix*out2
	}
}

func (c *Chorus) Synthesize() bool {
	if !c.BaseModule.Synthesize() {
		return false
	}

	if c.Inputs[1].IsConnected() {
		c.synthesizeStereoInput()
		return true
	}

	inLeft := c.Inputs[0].Buffer
	outLeft := c.Outputs[0].Buffer
	outRight := c.Outputs[1].Buffer

	msSamps := c.Config.SampleRate * 0.001

	for i := 0; i < c.Config.BufferSize; i++ {
		d1Loc := math.Abs(msSamps * (c.delayCenter + c.delayRange*c.modShaper.Shape(c.mods[0].Generate())))
		d2Loc := math.Abs(msSamps * (c.delayCenter + c.delayRange*c.modShaper.Shape(c.mods[1].Generate())))
		d3Loc := math.Abs(msSamps * (c.delayCenter + c.delayRange*c.modShaper.Shape(c.mods[2].Generate())))
		d4Loc := math.Abs(msSamps * (c.delayCenter + c.delayRange*c.modShaper.Shape(c.mods[3].Generate())))

		d1 := c.delayLineLeft.ReadLinear(d1Loc)
		d2 := c.delayLineLeft.ReadLinear(d2Loc)
		d3 := c.delayLineLeft.ReadLinear(d3Loc)
		d4 := c.delayLineLeft.ReadLinear(d4Loc)

		im1 := d1 + d3
		im2 := d2 + d4

		out1 := im1 + im2*(1.0-c.width)
		out2 := im1*(1.0-c.width) + im2

		c.delayLineLeft.Write(inLeft[i] + c.fb*c.lp1.filter((im1+im2)*0.25))

		outLeft[i] = inLeft[i]*(1.0-c.mix) + c.mix*out1
		outRight[i] = inLeft[i]*(1.0-c.mix) + c.mix*out2
	}

	return true
}

package blosc

import "math"

// Band limited oscillator
type Osc struct {
	lastOutput   float64
	sr           float64
	phase        float64
	frequency    float64
	dt           float64
	pw           float64
	mix          [4]float64
	mixScale     float64
	outputVector [5]float64
}

func NewOsc(frequency float64, phase float64, sr float64) *Osc {
	return NewOscX(frequency, phase, 0.5, [4]float64{1.0, 0.0, 0.0, 0.0}, sr)
}

func NewOscX(frequency float64, phase float64, pw float64, mix [4]float64, sr float64) *Osc {
	o := &Osc{
		frequency: frequency,
		phase:     phase,
		pw:        pw,
		mix:       mix,
		sr:        sr,
	}

	o.updateMixScale()

	return o
}

func (o *Osc) MixAt(i int) float64 {
	return o.mix[i]
}

func (o *Osc) Mix() [4]float64 {
	return o.mix
}

func (o *Osc) SetMixAt(i int, mix float64) {
	o.mix[i] = mix
	o.updateMixScale()
}

func (o *Osc) SetMix(mix [4]float64) {
	o.mix = mix
	o.updateMixScale()
}

func (o *Osc) Phase() float64 {
	return o.phase
}

func (o *Osc) SetPhase(ph float64) {
	o.phase = ph
}

func (o *Osc) SetFrequency(fc float64) {
	o.frequency = fc
	o.dt = o.frequency / o.sr
}

func (o *Osc) Frequency() float64 {
	return o.frequency
}

func (o *Osc) SetPulseWidth(pw float64) {
	o.pw = pw
}

func (o *Osc) PulseWidth() float64 {
	return o.pw
}

func (o *Osc) updateMixScale() {
	mixScale := o.mix[0] + o.mix[1] + o.mix[2] + o.mix[3]

	if mixScale > 0 {
		mixScale = 1.0 / mixScale
	}

	o.mixScale = mixScale
}

func polyBlep(t float64, dt float64) float64 {
	if t < dt {
		t /= dt
		return t + t - t*t - 1.0
	} else if t > 1.0-dt {
		t = (t - 1.0) / dt
		return t*t + t + t + 1.0
	}

	return 0.0
}

func (o *Osc) NumDimensions() int {
	return 5
}

func (o *Osc) Tick() []float64 {
	pw := o.pw
	dt := o.dt
	t := o.phase

	o.phase += dt

	for o.phase >= 1.0 {
		o.phase -= 1.0
	}

	sinSamp := math.Sin(t * math.Pi * 2.0)

	sawSamp := 2.0*t - 1.0
	sawSamp -= polyBlep(t, dt)

	pwSamp := 0.0
	sqrSamp := 0.0

	if t < pw {
		pwSamp = 1.0
	} else {
		pwSamp = -1.0
	}

	if t < 0.5 {
		sqrSamp = 1.0
	} else {
		sqrSamp = -1.0
	}

	sqrSamp += polyBlep(t, dt)
	pwSamp += polyBlep(t, dt)

	sqrSamp -= polyBlep(math.Mod(t+0.5, 1.0), dt)
	pwSamp -= polyBlep(math.Mod(t+(1.0-pw), 1.0), dt)

	// Use square wave as input, leaky integration
	triSamp := dt*sqrSamp + (1.0-dt)*o.lastOutput
	o.lastOutput = triSamp

	triSamp *= 4.0
	if triSamp > 1.0 {
		triSamp = 1.0
	}
	if triSamp < -1.0 {
		triSamp = -1.0
	}

	o.outputVector[0] = sinSamp
	o.outputVector[1] = sawSamp
	o.outputVector[2] = pwSamp
	o.outputVector[3] = triSamp
	o.outputVector[4] = o.mixScale * (sinSamp*o.mix[0] + sawSamp*o.mix[1] + pwSamp*o.mix[2] + triSamp*o.mix[3])

	return o.outputVector[:]
}

package blosc

import (
	"math"

	"github.com/almerlucke/muse"
)

// Band limited oscillator
type Osc struct {
	*muse.BaseModule
	lastOutput float64
	phase      float64
	frequency  float64
	pw         float64
	mix        [4]float64
}

func NewOsc(frequency float64, phase float64, config *muse.Configuration, identifier string) *Osc {
	return NewOscX(frequency, phase, 0.5, [4]float64{0.5, 0.01, 0.1, 0.5}, config, identifier)
}

func NewOscX(frequency float64, phase float64, pw float64, mix [4]float64, config *muse.Configuration, identifier string) *Osc {
	osc := &Osc{
		BaseModule: muse.NewBaseModule(3, 5, config, identifier),
		frequency:  frequency,
		phase:      phase,
		pw:         pw,
		mix:        mix,
	}

	osc.SetSelf(osc)

	return osc
}

func (o *Osc) MixAt(i int) float64 {
	return o.mix[i]
}

func (o *Osc) Mix() [4]float64 {
	return o.mix
}

func (o *Osc) SetMixAt(i int, mix float64) {
	o.mix[i] = mix
}

func (o *Osc) SetMix(mix [4]float64) {
	o.mix = mix
}

func (o *Osc) Phase() float64 {
	return o.phase
}

func (o *Osc) SetPhase(ph float64) {
	o.phase = ph
}

func (o *Osc) SetFrequency(fc float64) {
	o.frequency = fc
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

func (o *Osc) ReceiveControlValue(value any, index int) {
	switch index {
	case 0: // Frequency
		o.SetFrequency(value.(float64))
	case 1: // Phase
		o.SetPhase(value.(float64))
	case 2: // Pulse Width
		o.SetPulseWidth(value.(float64))
	case 3: // Mix Sine
		o.SetMixAt(0, value.(float64))
	case 4: // Mix Saw
		o.SetMixAt(1, value.(float64))
	case 5: // Mix Pulse
		o.SetMixAt(2, value.(float64))
	case 6: // Mix Tri
		o.SetMixAt(3, value.(float64))
	}
}

func (o *Osc) ReceiveMessage(msg any) []*muse.Message {
	params, ok := msg.(map[string]any)
	if ok {
		f, ok := params["frequency"]
		if ok {
			o.SetFrequency(f.(float64))
		}

		ph, ok := params["phase"]
		if ok {
			o.SetPhase(ph.(float64))
		}

		pw, ok := params["pulseWidth"]
		if ok {
			o.SetPulseWidth(pw.(float64))
		}

		mix1, ok := params["mix1"]
		if ok {
			o.SetMixAt(0, mix1.(float64))
		}

		mix2, ok := params["mix2"]
		if ok {
			o.SetMixAt(1, mix2.(float64))
		}

		mix3, ok := params["mix3"]
		if ok {
			o.SetMixAt(2, mix3.(float64))
		}

		mix4, ok := params["mix4"]
		if ok {
			o.SetMixAt(3, mix4.(float64))
		}
	}

	return nil
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

func (o *Osc) Synthesize() bool {
	if !o.BaseModule.Synthesize() {
		return false
	}

	frequency := o.frequency
	phaseOffset := 0.0
	pw := o.pw

	freqInput := o.InputAtIndex(0)
	phaseOffsetInput := o.InputAtIndex(1)
	pwInput := o.InputAtIndex(2)

	sinOut := o.OutputAtIndex(0).Buffer
	sawOut := o.OutputAtIndex(1).Buffer
	pwOut := o.OutputAtIndex(2).Buffer
	triOut := o.OutputAtIndex(3).Buffer
	mixOut := o.OutputAtIndex(4).Buffer

	mixScale := o.mix[0] + o.mix[1] + o.mix[2] + o.mix[3]

	if mixScale > 0 {
		mixScale = 1.0 / mixScale
	}

	for i := 0; i < o.Config.BufferSize; i++ {
		var sinSamp, sawSamp, pwSamp, sqrSamp, triSamp float64

		if freqInput.IsConnected() {
			frequency = float64(freqInput.Buffer[i])
		}
		if phaseOffsetInput.IsConnected() {
			phaseOffset = float64(phaseOffsetInput.Buffer[i])
		}
		if pwInput.IsConnected() {
			pw = float64(pwInput.Buffer[i])
		}

		dt := frequency / o.Config.SampleRate

		t := o.phase + phaseOffset

		// Fold phase back because of possible overshoot by adding phase offset
		for t >= 1.0 {
			t -= 1.0
		}

		for t < 0.0 {
			t += 1.0
		}

		sinSamp = math.Sin(t * math.Pi * 2.0)

		sawSamp = 2.0*t - 1.0
		sawSamp -= polyBlep(t, dt)

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
		triSamp = dt*sqrSamp + (1.0-dt)*o.lastOutput
		o.lastOutput = triSamp

		triSamp *= 4.0
		if triSamp > 1.0 {
			triSamp = 1.0
		}
		if triSamp < -1.0 {
			triSamp = -1.0
		}

		// Separate outputs
		sinOut[i] = sinSamp
		sawOut[i] = sawSamp
		pwOut[i] = pwSamp
		triOut[i] = triSamp

		// Mixed output
		mixOut[i] = mixScale * (sinSamp*o.mix[0] + sawSamp*o.mix[1] + pwSamp*o.mix[2] + triSamp*o.mix[3])

		// Update phase
		o.phase += dt

		for o.phase >= 1.0 {
			o.phase -= 1.0
		}
	}

	return true
}

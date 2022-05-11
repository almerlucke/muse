package blosc

import (
	"math"

	"github.com/almerlucke/muse"
)

// Band limited oscillator
type BloscModule struct {
	*muse.BaseModule
	lastOutput float64
	phase      float64
	frequency  float64
	amplitude  float64
}

func NewBloscModule(frequency float64, phase float64, amplitude float64, config *muse.Configuration, identifier string) *BloscModule {
	return &BloscModule{
		BaseModule: muse.NewBaseModule(3, 4, config, identifier),
		phase:      phase,
		frequency:  frequency,
		amplitude:  amplitude,
	}
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

func (b *BloscModule) ReceiveMessage(msg any) []*muse.Message {
	params, ok := msg.(map[string]any)
	if ok {
		f, ok := params["frequency"]
		if ok {
			b.frequency = f.(float64)
		}
	}

	return nil
}

func (b *BloscModule) Synthesize() bool {
	if !b.BaseModule.Synthesize() {
		return false
	}

	frequency := b.frequency
	amplitude := b.amplitude
	phaseOffset := 0.0

	freqInput := b.InputAtIndex(0)
	phaseOffsetInput := b.InputAtIndex(1)
	ampInput := b.InputAtIndex(2)

	sinOut := b.OutputAtIndex(0).Buffer
	sawOut := b.OutputAtIndex(1).Buffer
	sqrOut := b.OutputAtIndex(2).Buffer
	triOut := b.OutputAtIndex(3).Buffer

	for i := 0; i < b.Config.BufferSize; i++ {
		if freqInput.IsConnected() {
			frequency = float64(freqInput.Buffer[i])
		}
		if phaseOffsetInput.IsConnected() {
			phaseOffset = float64(phaseOffsetInput.Buffer[i])
		}
		if ampInput.IsConnected() {
			amplitude = float64(ampInput.Buffer[i])
		}

		dt := frequency / b.Config.SampleRate
		t := b.phase + phaseOffset

		var sinSamp, sawSamp, sqrSamp, triSamp float64

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

		if t < 0.5 {
			sqrSamp = 1.0
		} else {
			sqrSamp = -1.0
		}

		sqrSamp += polyBlep(t, dt)
		sqrSamp -= polyBlep(math.Mod(t+0.5, 1.0), dt)

		// Use square wave as input, leaky integration
		triSamp = dt*sqrSamp + (1.0-dt)*b.lastOutput
		b.lastOutput = triSamp
		// Boost signal with triangle
		triSamp *= 2.0

		b.phase += dt

		// Keep phase within 0-1 bounds
		for b.phase >= 1.0 {
			b.phase -= 1.0
		}

		for b.phase < 0.0 {
			b.phase += 1.0
		}

		sinOut[i] = sinSamp * amplitude
		sawOut[i] = sawSamp * amplitude
		sqrOut[i] = sqrSamp * amplitude
		triOut[i] = triSamp * amplitude
	}

	return true
}

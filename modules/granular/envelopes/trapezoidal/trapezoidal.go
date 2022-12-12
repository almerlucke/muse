package trapezoidal

import (
	"math"

	"github.com/almerlucke/muse/modules/granular"
)

type Stage int

const (
	Attack Stage = iota
	Sustain
	Release
)

type Parameter interface {
	Attack() float64
	Release() float64
	Smoothness() float64
}

type Envelope struct {
	stage           Stage
	stageDuration   int64
	sustainDuration int64
	releaseDuration int64
	increment       float64
	phase           float64
	maxAmplitude    float64
	smoothness      float64
	cnt             int64
}

type Factory struct{}

func (f *Factory) New() granular.Envelope {
	return &Envelope{}
}

func (e *Envelope) Activate(amplitude float64, durationSamples int64, parameter granular.Parameter) {
	envParam := any(parameter).(Parameter)

	e.smoothness = envParam.Smoothness()
	e.stage = Attack
	e.stageDuration = int64(float64(durationSamples) * envParam.Attack())
	e.releaseDuration = int64(float64(durationSamples) * envParam.Release())
	e.sustainDuration = durationSamples - (e.stageDuration + e.releaseDuration)
	e.increment = 1.0 / float64(e.stageDuration)
	e.phase = 0
	e.maxAmplitude = amplitude
	e.cnt = 0
}

func (e *Envelope) Synthesize(buf []float64, bufSize int) {
	for i := 0; i < bufSize; i++ {
		var out float64

		// Interpolate with smoothness between trapezoidal and raised cosine bell
		if e.stage == Attack {
			outSmooth := (1.0 + math.Cos(math.Pi+math.Pi*e.phase)) * (e.maxAmplitude / 2.0)
			outRaw := e.phase * e.maxAmplitude
			out = outRaw + (outSmooth-outRaw)*e.smoothness
		} else if e.stage == Sustain {
			out = e.maxAmplitude
		} else {
			outSmooth := (1.0 + math.Cos(math.Pi*e.phase)) * (e.maxAmplitude / 2.0)
			outRaw := (1.0 - e.phase) * e.maxAmplitude
			out = outRaw + (outSmooth-outRaw)*e.smoothness
		}

		e.phase += e.increment
		e.cnt++

		if e.cnt >= e.stageDuration {
			e.cnt = 0
			e.phase = 0
			if e.stage == Attack {
				e.stage = Sustain
				e.stageDuration = e.sustainDuration
			} else if e.stage == Sustain {
				e.stage = Release
				e.stageDuration = e.releaseDuration
				e.increment = 1.0 / float64(e.releaseDuration)
			}
		}

		buf[i] = out
	}
}

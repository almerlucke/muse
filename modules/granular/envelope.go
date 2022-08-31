package granular

import "math"

type DefaultEnvelopeType int

const (
	Parabolic DefaultEnvelopeType = iota
	Trapezoidal
	RaisedCosineBell
)

type DefaultEnvelopeStage int

const (
	Attack DefaultEnvelopeStage = iota
	Sustain
	Release
)

type DefaultEnvelopeParameter interface {
	Attack() float64
	Release() float64
	EnvType() DefaultEnvelopeType
}

type DefaultEnvelope struct {
	envType DefaultEnvelopeType

	// Parabolic
	currentAmp float64
	slope      float64
	curve      float64

	// Trapezoidal and Raised Cosine Bell
	stage           DefaultEnvelopeStage
	stageDuration   int64
	sustainDuration int64
	releaseDuration int64
	increment       float64
	phase           float64
	maxAmplitude    float64
	cnt             int64
}

type DefaultEnvelopeFactory struct{}

func (f *DefaultEnvelopeFactory) New() Envelope {
	return &DefaultEnvelope{}
}

func (e *DefaultEnvelope) Activate(amplitude float64, durationSamples int64, parameter Parameter) {
	envParam := any(parameter).(DefaultEnvelopeParameter)

	e.envType = envParam.EnvType()

	if e.envType == Parabolic {
		e.currentAmp = 0
		rdur := 1.0 / float64(durationSamples)
		rdur2 := rdur * rdur
		e.slope = 4.0 * amplitude * (rdur - rdur2)
		e.curve = -8.0 * amplitude * rdur2
	} else {
		e.stage = Attack
		e.stageDuration = int64(float64(durationSamples) * envParam.Attack())
		e.releaseDuration = int64(float64(durationSamples) * envParam.Release())
		e.sustainDuration = durationSamples - (e.stageDuration + e.releaseDuration)
		e.increment = 1.0 / float64(e.stageDuration)
		e.phase = 0
		e.maxAmplitude = amplitude
		e.cnt = 0
	}
}

func (e *DefaultEnvelope) Synthesize(buf []float64, bufSize int) {
	for i := 0; i < bufSize; i++ {
		var out float64

		if e.envType == Parabolic {
			out = e.currentAmp
			e.currentAmp = e.currentAmp + e.slope
			e.slope = e.slope + e.curve
		} else {
			if e.stage == Attack {
				if e.envType == RaisedCosineBell {
					out = (1.0 + math.Cos(math.Pi+math.Pi*e.phase)) * (e.maxAmplitude / 2.0)
				} else {
					out = e.phase * e.maxAmplitude
				}
			} else if e.stage == Sustain {
				out = e.maxAmplitude
			} else {
				if e.envType == RaisedCosineBell {
					out = (1.0 + math.Cos(math.Pi*e.phase)) * (e.maxAmplitude / 2.0)
				} else {
					out = (1.0 - e.phase) * e.maxAmplitude
				}
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
		}

		buf[i] = out
	}
}

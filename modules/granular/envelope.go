package granular

import "math"

type EnvelopeType int

const (
	Parabolic EnvelopeType = iota
	Trapezoidal
	RaisedCosineBell
)

type EnvelopeStage int

const (
	Attack EnvelopeStage = iota
	Sustain
	Release
)

type EnvelopeConfiguration struct {
	Amplitude       float64
	DurationSamples int64
	Attack          float64
	Release         float64
	Type            EnvelopeType
}

type Envelope struct {
	envType EnvelopeType

	// Parabolic
	currentAmp float64
	slope      float64
	curve      float64

	// Trapezoidal and Raised Cosine Bell
	stage           EnvelopeStage
	stageDuration   int64
	sustainDuration int64
	releaseDuration int64
	increment       float64
	phase           float64
	maxAmplitude    float64
	cnt             int64
}

func (e *Envelope) Activate(config EnvelopeConfiguration) {
	e.envType = config.Type

	if config.Type == Parabolic {
		e.currentAmp = 0
		rdur := 1.0 / float64(config.DurationSamples)
		rdur2 := rdur * rdur
		e.slope = 4.0 * config.Amplitude * (rdur - rdur2)
		e.curve = -8.0 * config.Amplitude * rdur2
	} else {
		e.stage = Attack
		e.stageDuration = int64(float64(config.DurationSamples) * config.Attack)
		e.releaseDuration = int64(float64(config.DurationSamples) * config.Release)
		e.sustainDuration = config.DurationSamples - (e.stageDuration + e.releaseDuration)
		e.increment = 1.0 / float64(e.stageDuration)
		e.phase = 0
		e.maxAmplitude = config.Amplitude
		e.cnt = 0
	}
}

func (e *Envelope) Synthesize() float64 {
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

	return out
}

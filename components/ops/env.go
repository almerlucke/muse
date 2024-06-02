package ops

import "math"

/*

Rate 0 - 99

 time is equal to distance divided by rate of speed.

That is, T=d/r

Incidentally, a rising envelope segment on the DX will take about 90 seconds to move from
level 00 to level 99 if the rate is set to 00. A falling segment takes 3½ minutes.
So if you set all the values to their greatest extreme, you'll get about a 10-minute envelope.


The envelope is not linear. Appears exponential, and at parity of rate the rising of the envelope is faster than the decay.
Levels are not linear: level 50 is very close to zero, and level 80 is less than half of level 99.

it emerges clearly that the decreasing edges are linear in log scale: this means they're pure exponential. The increasing segments are more complicated.

exp(-d * t)

T = 1 / d
*/

type envelopeDirection int

const (
	cEnvelopeRising envelopeDirection = iota
	cEnvelopeFalling
)

const (
	cRateDurationCurve      = 2.5
	cRateRisingDurationMax  = 90.0
	cRateRisingDurationMin  = 0.001
	cRateFallingDurationMax = 210.0
	cRateFallingDurationMin = 0.001
	cLevelRisingCurve       = 0.75
	cLevelFallingCurve      = 0.95
)

type EnvelopeReleaseMode int

const (
	// EnvelopeAutomaticRelease automatic release when sustain duration has passed
	EnvelopeAutomaticRelease EnvelopeReleaseMode = iota
	// EnvelopeNoteOffRelease release when NoteOff() is called
	EnvelopeNoteOffRelease
	// EnvelopeDurationRelease release when duration has passed
	EnvelopeDurationRelease
)

type EnvelopeListener interface {
	Start(*Envelope)
	Release(*Envelope)
	Finished(*Envelope)
}

type Envelope struct {
	Levels           [4]float64
	Rates            [4]float64
	Listener         EnvelopeListener
	index            int
	from             float64
	to               float64
	ramp             float64
	inc              float64
	curve            float64
	sr               float64
	lastOut          float64
	sampsToGo        int64
	releaseSampsToGo int64
	releaseMode      EnvelopeReleaseMode
}

func NewEnvelope(levels [4]float64, rates [4]float64, sr float64, releaseMode EnvelopeReleaseMode) *Envelope {
	env := &Envelope{
		Levels:      levels,
		Rates:       rates,
		sr:          sr,
		releaseMode: releaseMode,
		lastOut:     levels[3],
		index:       4,
	}

	return env
}

func (env *Envelope) SetReleaseMode(releaseMode EnvelopeReleaseMode) {
	env.releaseMode = releaseMode
}

func (env *Envelope) waitForRelease() bool {
	return env.releaseMode == EnvelopeDurationRelease || env.releaseMode == EnvelopeNoteOffRelease
}

func (env *Envelope) toIndex(index int, from float64) {
	env.index = index
	env.from = from
	env.to = env.Levels[index]
	direction := cEnvelopeRising
	env.curve = cLevelRisingCurve
	if env.to < env.from {
		direction = cEnvelopeFalling
		env.curve = cLevelFallingCurve
	}
	env.ramp = 0
	env.sampsToGo = int64(rateToSeconds(env.Rates[index], direction, cRateDurationCurve) * env.sr)
	env.inc = 1.0 / float64(env.sampsToGo)
}

func (env *Envelope) TriggerHard(duration float64) {
	env.toIndex(0, env.Levels[3])
	env.lastOut = env.from
	env.releaseSampsToGo = int64(duration * env.sr)
	// Start
	if env.Listener != nil {
		env.Listener.Start(env)
	}
}

func (env *Envelope) Trigger(duration float64) {
	env.toIndex(0, env.lastOut)
	env.lastOut = env.from
	env.releaseSampsToGo = int64(duration * env.sr)
	// Start
	if env.Listener != nil {
		env.Listener.Start(env)
	}
}

func (env *Envelope) NoteOff() {
	env.toIndex(3, env.lastOut)
	// Release
	if env.Listener != nil {
		env.Listener.Release(env)
	}
}

func (env *Envelope) Idle() bool {
	return env.index > 3
}

func (env *Envelope) Tick() float64 {
	if env.index > 3 {
		return env.Levels[3]
	}

	if env.releaseMode == EnvelopeDurationRelease && env.index <= 2 {
		env.releaseSampsToGo--
		if env.releaseSampsToGo <= 0 {
			defer env.NoteOff()
		}
	}

	if env.waitForRelease() && env.index == 2 && env.sampsToGo <= 0 {
		return env.Levels[2]
	}

	env.lastOut = env.from + math.Pow(env.ramp, env.curve)*(env.to-env.from)
	env.ramp += env.inc
	env.sampsToGo--

	if env.sampsToGo <= 0 {
		if !(env.waitForRelease() && env.index == 2) {
			if env.index == 2 {
				// Automatic release
				if env.Listener != nil {
					env.Listener.Release(env)
				}
			}
			env.index++
			if env.index <= 3 {
				env.toIndex(env.index, env.Levels[env.index-1])
			} else {
				// Finished
				if env.Listener != nil {
					env.Listener.Finished(env)
				}
			}
		}
	}

	return env.lastOut
}

func rateToSeconds(rate float64, direction envelopeDirection, curve float64) float64 {
	var mi float64
	var ma float64

	if direction == cEnvelopeRising {
		mi = cRateRisingDurationMin
		ma = cRateRisingDurationMax
	} else {
		mi = cRateFallingDurationMin
		ma = cRateFallingDurationMax
	}

	return mi + (ma-mi)*math.Pow(1.0-rate, curve)
}

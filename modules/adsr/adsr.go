package adsr

import (
	"math"

	"github.com/almerlucke/muse"
)

type ADSRStage int

const (
	AttackStage ADSRStage = iota
	DecayStage
	SustainStage
	ReleaseStage
	IdleStage
)

type ADSRStep struct {
	LevelRatio    float64
	DurationRatio float64
	Shape         float64
}

type ADSR struct {
	stage      ADSRStage
	steps      []ADSRStep
	sampleRate float64
	maxAmp     float64
	cnt        int64
	exponent   float64
	from       float64
	to         float64
	normalized float64
	increment  float64
}

func (adsr *ADSR) durationRatioToSamps(ratio float64) int64 {
	return int64((0.001 + math.Pow(ratio, 4.0)*19.999) * adsr.sampleRate)
}

func (adsr *ADSR) exponentFromShape(shape float64, direction float64) float64 {
	if shape == 0.0 {
		return 1.0
	}

	if (shape > 0.0 && direction > 0) || (shape < 0.0 && direction < 0) {
		return 1.0 / (1.0 + math.Abs(shape)*4.0)
	}

	return 1.0 + math.Abs(shape)*4.0
}

func (adsr *ADSR) Set(steps []ADSRStep, maxAmp float64, sampleRate float64) {
	adsr.sampleRate = sampleRate
	adsr.maxAmp = maxAmp
	adsr.steps = steps
	adsr.normalized = 0.0
	adsr.from = 0.0
	adsr.to = steps[AttackStage].LevelRatio * maxAmp
	adsr.cnt = adsr.durationRatioToSamps(steps[AttackStage].DurationRatio)
	adsr.increment = 1.0 / float64(adsr.cnt)
	adsr.exponent = adsr.exponentFromShape(steps[AttackStage].Shape, adsr.to-adsr.from)
	adsr.stage = AttackStage
}

func (adsr *ADSR) Synthesize() float64 {
	if adsr.stage == IdleStage {
		return 0.0
	}

	out := adsr.from + math.Pow(adsr.normalized, adsr.exponent)*(adsr.to-adsr.from)
	adsr.normalized += adsr.increment
	adsr.cnt--
	if adsr.cnt <= 0 {
		adsr.stage++
		if adsr.stage < IdleStage {
			step := adsr.steps[adsr.stage]
			adsr.normalized = 0
			adsr.from = adsr.to
			adsr.cnt = adsr.durationRatioToSamps(step.DurationRatio)
			adsr.increment = 1.0 / float64(adsr.cnt)
			switch adsr.stage {
			case DecayStage:
				adsr.to = step.LevelRatio * adsr.maxAmp
				adsr.exponent = adsr.exponentFromShape(step.Shape, adsr.to-adsr.from)
			case SustainStage:
				adsr.exponent = 1.0
			case ReleaseStage:
				adsr.to = 0.0
				adsr.exponent = adsr.exponentFromShape(step.Shape, adsr.to-adsr.from)
			}
		}
	}

	return out
}

type ADSRModule struct {
	*muse.BaseModule
}

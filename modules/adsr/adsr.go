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
	ramp       float64
	increment  float64
	lastOut    float64
}

func (adsr *ADSR) durationRatioToSamps(ratio float64) int64 {
	return int64((0.01 + math.Pow(ratio, 4.0)*19.99) * adsr.sampleRate)
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

func (adsr *ADSR) Initialize(steps []ADSRStep, maxAmp float64, sampleRate float64) {
	adsr.sampleRate = sampleRate
	adsr.maxAmp = maxAmp
	adsr.steps = steps
	adsr.ramp = 0.0
	adsr.from = 0.0
	adsr.to = steps[AttackStage].LevelRatio * maxAmp
	adsr.cnt = adsr.durationRatioToSamps(steps[AttackStage].DurationRatio)
	adsr.increment = 1.0 / float64(adsr.cnt)
	adsr.exponent = adsr.exponentFromShape(steps[AttackStage].Shape, adsr.to-adsr.from)
	adsr.stage = IdleStage
}

func (adsr *ADSR) Retrigger(maxAmp float64) {
	adsr.maxAmp = maxAmp
	adsr.ramp = 0.0
	adsr.from = adsr.lastOut
	adsr.stage = AttackStage
	adsr.to = adsr.steps[AttackStage].LevelRatio * adsr.maxAmp
	adsr.cnt = adsr.durationRatioToSamps(adsr.steps[AttackStage].DurationRatio)
	adsr.increment = 1.0 / float64(adsr.cnt)
	adsr.exponent = adsr.exponentFromShape(adsr.steps[AttackStage].Shape, adsr.to-adsr.from)
}

func (adsr *ADSR) SetSteps(steps []ADSRStep) {
	adsr.steps = steps
}

func (adsr *ADSR) Synthesize() float64 {
	if adsr.stage == IdleStage {
		adsr.lastOut = 0.0
		return 0.0
	}

	out := adsr.from + math.Pow(adsr.ramp, adsr.exponent)*(adsr.to-adsr.from)
	adsr.lastOut = out
	adsr.ramp += adsr.increment
	adsr.cnt--
	if adsr.cnt <= 0 {
		adsr.stage++
		if adsr.stage < IdleStage {
			step := adsr.steps[adsr.stage]
			adsr.ramp = 0
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
	adsr *ADSR
}

func NewADSRModule(steps []ADSRStep, maxAmp float64, config *muse.Configuration, identifier string) *ADSRModule {
	am := &ADSRModule{
		BaseModule: muse.NewBaseModule(0, 1, config, identifier),
	}

	am.adsr = &ADSR{}
	am.adsr.Initialize(steps, maxAmp, config.SampleRate)

	return am
}

func (a *ADSRModule) ReceiveMessage(msg any) []*muse.Message {
	bang, ok := msg.(string)
	if ok && bang == "bang" {
		a.adsr.Retrigger(a.adsr.maxAmp)
	}

	return nil
}

func (a *ADSRModule) Synthesize() bool {
	if !a.BaseModule.Synthesize() {
		return false
	}

	out := a.Outputs[0].Buffer

	for i := 0; i < a.Config.BufferSize; i++ {
		out[i] = a.adsr.Synthesize()
	}

	return true
}

package adsr

import (
	"math"
)

type Stage int

const (
	Attack Stage = iota
	Decay
	Sustain
	Release
	Idle
)

type DurationMode int

const (
	Ratio    DurationMode = iota // use ratio for durations
	Absolute                     // use absolute ms for durations
)

type ReleaseMode int

const (
	Automatic ReleaseMode = iota // automatic release when sustain duration has passed
	Duration                     // explicit release when duration has passed
	NoteOff                      // release when Release() is called (note off)
)

const (
	ShapeMult        = 2.0
	MinDuration      = 0.01
	MaxDuration      = 19.99
	DurationRatioExp = 6.0
)

type Step struct {
	Level         float64 // level of the stage, is relative to max level
	DurationRatio float64
	Duration      float64 // duration in milliseconds
	Shape         float64
}

type StepProvider interface {
	GetSteps() []Step
}

type BasicStepProvider struct {
	Steps []Step
}

func NewBasicStepProvider() *BasicStepProvider {
	return &BasicStepProvider{
		Steps: make([]Step, 4),
	}
}

func (bsp *BasicStepProvider) GetSteps() []Step {
	return bsp.Steps
}

func (s *Step) GetState() map[string]any {
	return map[string]any{
		"level":         s.Level,
		"durationRatio": s.DurationRatio,
		"duration":      s.Duration,
		"shape":         s.Shape,
	}
}

func (s *Step) SetState(state map[string]any) {
	s.Level = state["level"].(float64)
	s.DurationRatio = state["durationRatio"].(float64)
	s.Duration = state["duration"].(float64)
	s.Shape = state["shape"].(float64)
}

type ADSR struct {
	releaseMode  ReleaseMode
	durationMode DurationMode
	stage        Stage
	steps        []Step
	releaseCnt   int64
	sampleRate   float64
	maxLevel     float64
	stepCnt      int64
	exponent     float64
	from         float64
	to           float64
	ramp         float64
	increment    float64
	lastOut      float64
	outVector    [1]float64
}

func (adsr *ADSR) durationRatioToSamps(ratio float64) int64 {
	return int64((MinDuration + math.Pow(ratio, DurationRatioExp)*MaxDuration) * adsr.sampleRate)
}

func (adsr *ADSR) exponentFromShape(shape float64, direction float64) float64 {
	if shape == 0.0 {
		return 1.0
	}

	if (shape > 0.0 && direction > 0) || (shape < 0.0 && direction < 0) {
		return 1.0 / (1.0 + math.Abs(shape)*ShapeMult)
	}

	return 1.0 + math.Abs(shape)*ShapeMult
}

func (adsr *ADSR) Initialize(steps []Step, durationMode DurationMode, releaseMode ReleaseMode, sampleRate float64) {
	adsr.sampleRate = sampleRate
	adsr.steps = steps
	adsr.durationMode = durationMode
	adsr.releaseMode = releaseMode
	adsr.stage = Idle
}

func (adsr *ADSR) TriggerFull(duration float64, maxLevel float64, steps []Step, durationMode DurationMode, releaseMode ReleaseMode) {
	adsr.releaseCnt = int64(duration * adsr.sampleRate * 0.001)
	adsr.steps = steps
	adsr.durationMode = durationMode
	adsr.releaseMode = releaseMode
	adsr.Trigger(maxLevel)
}

func (adsr *ADSR) TriggerWithDuration(duration float64, maxLevel float64) {
	adsr.releaseCnt = int64(duration * adsr.sampleRate * 0.001)
	adsr.Trigger(maxLevel)
}

func (adsr *ADSR) Trigger(maxLevel float64) {
	adsr.stage = Attack
	adsr.maxLevel = maxLevel
	adsr.ramp = 0.0
	adsr.from = adsr.lastOut
	adsr.to = adsr.steps[Attack].Level * adsr.maxLevel
	if adsr.durationMode == Ratio {
		adsr.stepCnt = adsr.durationRatioToSamps(adsr.steps[Attack].DurationRatio)
	} else {
		adsr.stepCnt = int64(adsr.steps[Attack].Duration * adsr.sampleRate * 0.001)
	}
	adsr.increment = 1.0 / float64(adsr.stepCnt)
	adsr.exponent = adsr.exponentFromShape(adsr.steps[Attack].Shape, adsr.to-adsr.from)
}

func (adsr *ADSR) Release() {
	adsr.from = adsr.lastOut
	adsr.stage = Release
	adsr.to = 0.0
	adsr.ramp = 0
	step := adsr.steps[Release]
	adsr.exponent = adsr.exponentFromShape(step.Shape, adsr.to-adsr.from)
	if adsr.durationMode == Ratio {
		adsr.stepCnt = adsr.durationRatioToSamps(step.DurationRatio)
	} else {
		adsr.stepCnt = int64(step.Duration * adsr.sampleRate * 0.001)
	}
	adsr.increment = 1.0 / float64(adsr.stepCnt)
}

func (adsr *ADSR) IsFinished() bool {
	return adsr.stage == Idle
}

func (adsr *ADSR) DurationMode() DurationMode {
	return adsr.durationMode
}

func (adsr *ADSR) ReleaseMode() ReleaseMode {
	return adsr.releaseMode
}

func (adsr *ADSR) NumDimensions() int {
	return 1
}

func (adsr *ADSR) Tick() []float64 {
	if adsr.stage == Idle {
		adsr.lastOut = 0.0
		adsr.outVector[0] = 0.0
		return adsr.outVector[:]
	}

	if adsr.releaseMode == Duration && adsr.releaseCnt > 0 {
		adsr.releaseCnt--
		if adsr.releaseCnt == 0 {
			adsr.Release()
		}
	}

	if (adsr.releaseMode == NoteOff || adsr.releaseMode == Duration) && adsr.stage == Sustain {
		adsr.lastOut = adsr.from
		adsr.outVector[0] = adsr.from
		return adsr.outVector[:]
	}

	out := adsr.from + math.Pow(adsr.ramp, adsr.exponent)*(adsr.to-adsr.from)
	adsr.lastOut = out
	adsr.ramp += adsr.increment
	adsr.stepCnt--
	if adsr.stepCnt <= 0 {
		adsr.stage++
		if adsr.stage < Idle {
			step := adsr.steps[adsr.stage]
			adsr.ramp = 0
			adsr.from = adsr.to
			if adsr.durationMode == Ratio {
				adsr.stepCnt = adsr.durationRatioToSamps(step.DurationRatio)
			} else {
				adsr.stepCnt = int64(step.Duration * adsr.sampleRate * 0.001)
			}
			adsr.increment = 1.0 / float64(adsr.stepCnt)
			switch adsr.stage {
			case Decay:
				adsr.to = step.Level * adsr.maxLevel
				adsr.exponent = adsr.exponentFromShape(step.Shape, adsr.to-adsr.from)
			case Sustain:
				adsr.exponent = 1.0
			case Release:
				adsr.to = 0.0
				adsr.exponent = adsr.exponentFromShape(step.Shape, adsr.to-adsr.from)
			}
		}
	}

	adsr.outVector[0] = out
	return adsr.outVector[:]
}

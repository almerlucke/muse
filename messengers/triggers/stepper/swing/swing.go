package swing

import (
	"github.com/almerlucke/genny"
	"math"
	"math/rand"
)

var _shuffleMultiplier = 0.7

type Step struct {
	Skip        bool    `json:"skip"`
	Shuffle     float64 `json:"shuffle"` // 0.0 - 1.0 -> (0.5 + shuffle * 0.35) * milliPerNote * 2.0  50% - 85%  --- 50, 54, 58, 62, 66, 70, 74, 78, 82, 86
	ShuffleRand float64 `json:"shuffleRand"`
	SkipChance  float64 `json:"skipChance"` // 0% - 90% chance of skipping
	Multiply    float64 `json:"multiply"`
	BurstChance float64 `json:"burstChance"`
	NumBurst    int     `json:"numBurst"`
}

func (s *Step) shuffleDelay(milliPerNote float64) float64 {
	shuffleRandBandwidth := math.Min(s.Shuffle, 1.0-s.Shuffle)
	shuffleAmount := s.Shuffle + (rand.Float64()*2.0-1.0)*math.Min(s.ShuffleRand, shuffleRandBandwidth)

	return _shuffleMultiplier * shuffleAmount * milliPerNote
}

func (s *Step) Burst() bool {
	return rand.Float64() < s.BurstChance
}

func (s *Step) SkipStep() bool {
	return s.Skip || rand.Float64() < s.SkipChance
}

/*
[]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	})


*/

func QuickSteps(args ...any) []*Step {
	steps := make([]*Step, len(args))

	for index, arg := range args {
		var step *Step

		switch v := arg.(type) {
		case int:
			step = &Step{
				Skip: v == 0,
			}
		case *Step:
			step = v
		default:
			step = &Step{
				Skip: true,
			}
		}

		steps[index] = step
	}

	return steps
}

type Swing struct {
	steps             genny.Generator[*Step]
	bpm               int
	noteDivision      int
	milliPerNote      float64
	remainingDuration float64
	delayed           bool
	burstCount        int
	burstMode         bool
	burstDuration     float64
}

func New(bpm int, noteDivision int, steps genny.Generator[*Step]) *Swing {
	return &Swing{
		steps:        steps,
		noteDivision: noteDivision,
		bpm:          bpm,
		milliPerNote: (60000.0 / float64(bpm)) / float64(noteDivision),
	}
}

func (sw *Swing) Generate() float64 {
	if sw.delayed {
		sw.delayed = false
		return sw.remainingDuration
	}

	if sw.burstMode {
		sw.burstCount--
		if sw.burstCount == 0 {
			sw.burstMode = false
		}

		return sw.burstDuration
	}

	step := sw.steps.Generate()

	multiply := 1.0

	if step.Multiply > 0.0 {
		multiply = step.Multiply
	}

	milliPerNote := sw.milliPerNote * multiply

	if step.SkipStep() {
		return -milliPerNote
	}

	if step.Burst() {
		sw.burstMode = true
		sw.burstCount = step.NumBurst
		sw.burstDuration = milliPerNote / float64(step.NumBurst)
		return sw.Generate()
	}

	delay := step.shuffleDelay(milliPerNote)
	if delay > 0.0 {
		sw.remainingDuration = milliPerNote - delay
		sw.delayed = true
		return -delay
	}

	return milliPerNote
}

func (sw *Swing) Done() bool {
	return sw.steps.Done()
}

func (sw *Swing) Reset() {
	sw.steps.Reset()
	sw.delayed = false
	sw.burstMode = false
}

func (sw *Swing) Continuous() bool {
	return sw.steps.Continuous()
}

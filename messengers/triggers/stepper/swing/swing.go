package swing

import (
	"math"
	"math/rand"

	"github.com/almerlucke/muse/value"
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

type Swing struct {
	steps             value.Valuer[*Step]
	bpm               int
	noteDivision      int
	milliPerNote      float64
	remainingDuration float64
	delayed           bool
	burstCount        int
	burstMode         bool
	burstDuration     float64
}

func New(bpm int, noteDivision int, steps value.Valuer[*Step]) *Swing {
	return &Swing{
		steps:        steps,
		noteDivision: noteDivision,
		bpm:          bpm,
		milliPerNote: (60000.0 / float64(bpm)) / float64(noteDivision),
	}
}

func (sw *Swing) NextStep() float64 {
	if sw.steps.Finished() {
		sw.steps.Reset()
	}

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

	step := sw.steps.Value()

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
		return sw.NextStep()
	}

	delay := step.shuffleDelay(milliPerNote)
	if delay > 0.0 {
		sw.remainingDuration = milliPerNote - delay
		sw.delayed = true
		return -delay
	}

	return milliPerNote
}

func (sw *Swing) GetState() map[string]any {
	return map[string]any{
		"steps":             sw.steps.GetState(),
		"noteDivision":      sw.noteDivision,
		"milliPerNote":      sw.milliPerNote,
		"bpm":               sw.bpm,
		"remainingDuration": sw.remainingDuration,
		"delayed":           sw.delayed,
		"burstCount":        sw.burstCount,
		"burstMode":         sw.burstMode,
		"burstDuration":     sw.burstDuration,
	}
}

func (sw *Swing) SetState(state map[string]any) {
	sw.steps.SetState(state["steps"].(map[string]any))
	sw.noteDivision = state["noteDivision"].(int)
	sw.bpm = state["bpm"].(int)
	sw.milliPerNote = (60000.0 / float64(sw.bpm)) / float64(sw.noteDivision)
	sw.remainingDuration = state["remainingDuration"].(float64)
	sw.delayed = state["delayed"].(bool)
	sw.burstCount = state["burstCount"].(int)
	sw.burstMode = state["burstMode"].(bool)
	sw.burstDuration = state["burstDuration"].(float64)
}

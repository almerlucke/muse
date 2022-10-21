package swing

import (
	"math/rand"

	"github.com/almerlucke/muse/value"
)

type Step struct {
	Skip        bool    `json:"skip"`
	Shuffle     float64 `json:"shuffle"` // 0.0 - 1.0 -> (0.5 + shuffle * 0.35) * milliPerNote * 2.0  50% - 85%  --- 50, 54, 58, 62, 66, 70, 74, 78, 82, 86
	ShuffleRand float64 `json:"shuffleRand"`
	SkipFactor  float64 `json:"skipFactor"` // 0% - 90% chance of skipping
}

func (s *Step) shuffleNote(milliPerNote float64) float64 {
	shuffleRandBandwidth := s.Shuffle

	if s.Shuffle > 0.5 {
		shuffleRandBandwidth = 1.0 - s.Shuffle
	}

	if s.ShuffleRand < shuffleRandBandwidth {
		shuffleRandBandwidth = s.ShuffleRand
	}

	finalShuffle := s.Shuffle + (rand.Float64()*2.0-1.0)*shuffleRandBandwidth

	return (0.5 + finalShuffle*0.35) * milliPerNote * 2.0
}

type Swing struct {
	steps        value.Valuer[*Step]
	bpm          int
	noteDivision int
	milliPerNote float64
	delay        float64
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
	step := sw.steps.Value()

	if sw.steps.Finished() {
		sw.steps.Reset()
	}

	dur := step.shuffleNote(sw.milliPerNote)
	delay := dur - sw.milliPerNote
	dur -= sw.delay
	sw.delay = delay

	if step.Skip || rand.Float64() < step.SkipFactor {
		return -dur
	}

	return dur
}

func (sw *Swing) GetState() map[string]any {
	return map[string]any{
		"steps":        sw.steps.GetState(),
		"noteDivision": sw.noteDivision,
		"bpm":          sw.bpm,
		"delay":        sw.delay,
	}
}

func (sw *Swing) SetState(state map[string]any) {
	sw.steps.SetState(state["steps"].(map[string]any))
	sw.noteDivision = state["noteDivision"].(int)
	sw.bpm = state["bpm"].(int)
	sw.milliPerNote = (60000.0 / float64(sw.bpm)) / float64(sw.noteDivision)
	sw.delay = state["delay"].(float64)
}

package swing

import (
	"math/rand"

	"github.com/almerlucke/muse/values"
)

type Step struct {
	Skip        bool    `json:"skip"`
	Shuffle     float64 `json:"shuffle"` // 0.0 - 1.0 -> (0.5 + shuffle * 0.35) * milliPerNote * 2.0  50% - 85%  --- 50, 54, 58, 62, 66, 70, 74, 78, 82, 86
	ShuffleRand float64 `json:"shuffleRand"`
	SkipFactor  float64 `json:"skipFactor"` // 0% - 90% chance of skipping
}

func (s *Step) shuffleNote(milliPerNote float64) float64 {
	finalShuffle := s.Shuffle + (rand.Float64()*2.0-1.0)*s.ShuffleRand

	if finalShuffle < 0.0 {
		finalShuffle += 1.0
	}

	if finalShuffle > 1.0 {
		finalShuffle -= 1.0
	}

	return (0.5 + finalShuffle*0.35) * milliPerNote * 2.0
}

type Swing struct {
	steps        values.Valuer[*Step]
	bpm          values.Valuer[float64]
	noteDivision values.Valuer[float64]
	delay        float64
}

func New(bpm values.Valuer[float64], noteDivision values.Valuer[float64], steps values.Valuer[*Step]) *Swing {
	return &Swing{
		steps:        steps,
		noteDivision: noteDivision,
		bpm:          bpm,
	}
}

func (sw *Swing) NextStep() float64 {
	milliPerBeat := 60000.0 / sw.bpm.Value()
	milliPerNote := milliPerBeat / sw.noteDivision.Value()

	step := sw.steps.Value()

	if sw.steps.Finished() {
		sw.steps.Reset()
	}

	dur := step.shuffleNote(milliPerNote)
	delay := dur - milliPerNote
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
		"noteDivision": sw.noteDivision.GetState(),
		"bpm":          sw.bpm.GetState(),
		"delay":        sw.delay,
	}
}

func (sw *Swing) SetState(state map[string]any) {
	sw.steps.SetState(state["steps"].(map[string]any))
	sw.noteDivision.SetState(state["noteDivision"].(map[string]any))
	sw.bpm.SetState(state["bpm"].(map[string]any))
	sw.delay = state["delay"].(float64)
}

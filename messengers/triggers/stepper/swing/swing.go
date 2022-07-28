package swing

import (
	"math/rand"

	"github.com/almerlucke/muse/values"
)

type Step struct {
	Skip        bool
	Shuffle     float64 // 0.0 - 1.0 -> (0.5 + shuffle * 0.35) * milliPerNote * 2.0  50% - 85%  --- 50, 54, 58, 62, 66, 70, 74, 78, 82, 86
	ShuffleRand float64
	SkipFactor  float64 // 0% - 90% chance of skipping
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
	currentStep  int
	steps        values.Generator[*Step]
	milliPerNote float64
	delay        float64
}

func New(bpm float64, noteDivision float64, steps values.Generator[*Step]) *Swing {
	milliPerBeat := 60000.0 / bpm
	milliPerNote := milliPerBeat / noteDivision

	return &Swing{
		steps:        steps,
		milliPerNote: milliPerNote,
	}
}

func (sw *Swing) NextStep() float64 {
	step := sw.steps.Next()

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

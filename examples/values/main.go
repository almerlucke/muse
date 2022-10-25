package main

import (
	"log"

	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/value"
)

func main() {
	s := swing.New(105, 8, value.NewSequence([]*swing.Step{
		{}, {Shuffle: 0.1}, {Skip: true}, {Shuffle: 0.1}, {}, {Shuffle: 0.1}, {}, {Shuffle: 0.1}, {}, {Shuffle: 0.1},
	}))

	for i := 0; i < 16; i++ {
		dur := s.NextStep()

		log.Printf("step: %d, dur: %f", i+1, dur)
	}
}

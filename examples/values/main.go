package main

import (
	"log"

	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
)

func main() {
	s := swing.New(105, 8, sequence.NewLoop([]*swing.Step{
		{}, {Shuffle: 0.1}, {Skip: true}, {Shuffle: 0.1}, {}, {Shuffle: 0.1}, {}, {Shuffle: 0.1}, {}, {Shuffle: 0.1},
	}...))

	for i := 0; i < 16; i++ {
		dur := s.NextStep()

		log.Printf("step: %d, dur: %f", i+1, dur)
	}
}

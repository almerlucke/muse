package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/almerlucke/muse/value/markov"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	var stopState *markov.State[int]

	state1 := markov.NewState(1, nil)
	state2 := markov.NewState(2, nil)
	state3 := markov.NewState(3, nil)
	state4 := markov.NewState(4, nil)
	state5 := markov.NewState(5, nil)

	state1.Transitioner = markov.NewProbabilityTransitionerVariadic[int](state2, 1.0, state3, 1.0)
	state2.Transitioner = markov.NewProbabilityTransitionerVariadic[int](state3, 1.0, state4, 1.0)
	state3.Transitioner = markov.NewProbabilityTransitionerVariadic[int](state4, 1.0, state5, 1.0)
	state4.Transitioner = markov.NewProbabilityTransitionerVariadic[int](state5, 1.0, state1, 1.0)
	state5.Transitioner = markov.NewProbabilityTransitionerVariadic[int](state1, 1.0, stopState, 1.0)

	m := markov.NewMarkov[int](markov.NewStateStarter(state1), 3)

	for !m.Finished() {
		log.Printf("state: %d", m.Value())
	}
}

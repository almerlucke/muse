package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/almerlucke/muse/value/markov"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	state1 := markov.NewState(1, nil)
	state2 := markov.NewState(2, nil)
	state3 := markov.NewState(3, nil)
	state4 := markov.NewState(4, nil)
	state5 := markov.NewState(5, nil)

	state1.Transitioner = markov.NewProbabilityTransitioner(
		[]*markov.ProbabilityTransition[int]{
			markov.NewProbabilityTransition(state2, 1),
			markov.NewProbabilityTransition(state3, 1),
		},
	)

	state2.Transitioner = markov.NewProbabilityTransitioner(
		[]*markov.ProbabilityTransition[int]{
			markov.NewProbabilityTransition(state3, 1),
			markov.NewProbabilityTransition(state4, 1),
		},
	)

	state3.Transitioner = markov.NewProbabilityTransitioner(
		[]*markov.ProbabilityTransition[int]{
			markov.NewProbabilityTransition(state4, 1),
			markov.NewProbabilityTransition(state5, 1),
		},
	)

	state4.Transitioner = markov.NewProbabilityTransitioner(
		[]*markov.ProbabilityTransition[int]{
			markov.NewProbabilityTransition(state5, 1),
			markov.NewProbabilityTransition(state1, 1),
		},
	)

	state5.Transitioner = markov.NewProbabilityTransitioner(
		[]*markov.ProbabilityTransition[int]{
			markov.NewProbabilityTransition(state1, 1),
			markov.NewProbabilityTransition[int](nil, 1),
		},
	)

	m := markov.NewMarkov[int](markov.NewStateStarter(state1), 3)

	for !m.Finished() {
		log.Printf("state: %d", m.Value())
	}
}

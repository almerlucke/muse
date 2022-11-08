package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/value"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	sw := swing.New(80, 4,
		value.NewSequence([]*swing.Step{{}, {Skip: true}}),
	)

	for i := 0; i < 100; i++ {
		st := sw.NextStep()
		log.Printf("step: %f", st)
	}
}

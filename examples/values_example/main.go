package main

import (
	"log"

	"github.com/almerlucke/muse/components/phasor"
	"github.com/almerlucke/muse/components/waveshaping"
)

func main() {
	ph := phasor.NewPhasor(100.0, 4410.0, 0.0)
	sh := waveshaping.NewSineTable(512.0)

	for i := 0; i < 1000; i++ {
		log.Printf("tick: %v", sh.Shape(ph.Tick()))
	}
}

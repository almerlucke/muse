package main

import (
	"log"

	"github.com/almerlucke/muse/modules/adsr"
)

func main() {
	steps := []adsr.ADSRStep{{
		LevelRatio:    1.0,
		DurationRatio: 0.05,
		Shape:         -0.5,
	}, {
		LevelRatio:    0.4,
		DurationRatio: 0.05,
		Shape:         -0.5,
	}, {
		DurationRatio: 0.05,
	}, {
		DurationRatio: 0.05,
		Shape:         -0.5,
	}}

	adsrEnv := &adsr.ADSR{}

	adsrEnv.Initialize(steps, 0.75, 44100.0)

	for i := 0; i < 100; i++ {
		log.Printf("out %v: %v", i+1, adsrEnv.Synthesize())
	}

	adsrEnv.Retrigger(1.0)

	for i := 0; i < 200; i++ {
		log.Printf("retrigger %v: %v", i+1, adsrEnv.Synthesize())
	}
}

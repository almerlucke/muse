package main

import (
	"log"

	"github.com/almerlucke/muse/components/envelopes/adsr"
)

func testRatioAuto() {
	steps := []adsr.Step{{
		Level:         1.0,
		DurationRatio: 0.7,
		Shape:         -0.5,
	}, {
		Level:         0.4,
		DurationRatio: 0.7,
		Shape:         -0.5,
	}, {
		DurationRatio: 0.7,
	}, {
		DurationRatio: 0.7,
		Shape:         -0.5,
	}}

	adsrEnv := &adsr.ADSR{}

	adsrEnv.Initialize(steps, adsr.Ratio, adsr.Automatic, 44100.0)
	adsrEnv.Trigger(1.0)

	index := 0
	for !adsrEnv.IsFinished() {
		index++
		log.Printf("out %v: %v", index, adsrEnv.Synthesize())
	}
}

func testAbsoluteAuto() {
	steps := []adsr.Step{{
		Level:    1.0,
		Duration: 1,
		Shape:    -0.5,
	}, {
		Level:    0.4,
		Duration: 1,
		Shape:    -0.5,
	}, {
		Duration: 1,
	}, {
		Duration: 1,
		Shape:    -0.5,
	}}

	adsrEnv := &adsr.ADSR{}

	adsrEnv.Initialize(steps, adsr.Absolute, adsr.Automatic, 44100.0)
	adsrEnv.Trigger(1.0)

	index := 0
	for !adsrEnv.IsFinished() {
		index++
		log.Printf("out %v: %v", index, adsrEnv.Synthesize())
	}
}

func testAbsoluteDuration() {
	steps := []adsr.Step{{
		Level:    1.0,
		Duration: 1,
		Shape:    -0.5,
	}, {
		Level:    0.4,
		Duration: 1,
		Shape:    -0.5,
	}, {
		Duration: 0.2,
	}, {
		Duration: 1,
		Shape:    -0.5,
	}}

	adsrEnv := &adsr.ADSR{}

	adsrEnv.Initialize(steps, adsr.Absolute, adsr.Automatic, 44100.0)
	adsrEnv.TriggerWithDuration(2.4, 1.0)
	adsrEnv.Trigger(1.0)

	index := 0
	for !adsrEnv.IsFinished() {
		index++
		log.Printf("out %v: %v", index, adsrEnv.Synthesize())
	}
}

func main() {
	testAbsoluteDuration()
}

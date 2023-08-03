package main

import (
	"log"

	"github.com/almerlucke/muse/components/envelopes/adsr"
)

func testAuto() {
	adsrEnv := adsr.New(adsr.NewSetting(1.0, 1.0, 0.2, 1.0, 1.0, 1.0), adsr.Automatic, 44100.0)
	adsrEnv.Trigger(1.0)

	index := 0
	for !adsrEnv.IsFinished() {
		index++
		log.Printf("out %v: %v", index, adsrEnv.Tick())
	}
}

func testDuration() {
	adsrEnv := adsr.New(adsr.NewSetting(1.0, 1.0, 0.2, 1.0, 1.0, 1.0), adsr.Duration, 44100.0)
	adsrEnv.TriggerWithDuration(2.4, 1.0)

	index := 0
	for index < 40 {
		index++
		log.Printf("out %v: %v", index, adsrEnv.Tick())
	}

	adsrEnv.Release()

	for !adsrEnv.IsFinished() {
		index++
		log.Printf("out %v: %v", index, adsrEnv.Tick())
	}

	index = 0
	adsrEnv.TriggerWithDuration(2.4, 1.0)
	for !adsrEnv.IsFinished() {
		index++
		log.Printf("out %v: %v", index, adsrEnv.Tick())
	}
}

func main() {
	testDuration()
}

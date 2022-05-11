package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/sequencer"
	"github.com/almerlucke/muse/messengers/timer"
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/blosc"
	"github.com/almerlucke/muse/modules/common"
)

func main() {
	env := muse.NewEnvironment(2, 44100, 128)

	// data, err := os.ReadFile("examples/blosc_example/events.json")
	// if err != nil {
	// 	log.Fatalf("error reading events: %v", err)
	// }

	// msgr, err := scheduler.NewSchedulerWithJSONData(data, "scheduler")
	// if err != nil {
	// 	log.Fatalf("error unmarshalling json events: %v", err)
	// }

	fs1 := []float64{100.0, 200.0, 50.0, 50.0}
	fs2 := []float64{201.0, 302.0, 51.0, 103.0, 400.0}
	sequence1 := [][]*muse.Message{}
	sequence2 := [][]*muse.Message{}

	for _, f := range fs1 {
		sequence1 = append(sequence1, []*muse.Message{
			{Address: "blosc1", Content: map[string]any{"frequency": f}},
		})
	}

	for _, f := range fs2 {
		sequence2 = append(sequence2, []*muse.Message{
			{Address: "blosc2", Content: map[string]any{"frequency": f}},
		})
	}

	env.AddMessenger(sequencer.NewSequencer(sequence1, "sequencer1"))
	env.AddMessenger(sequencer.NewSequencer(sequence2, "sequencer2"))
	env.AddMessenger(timer.NewTimer(0.25, []string{"sequencer1", "adsr1"}, env.Config, ""))
	env.AddMessenger(timer.NewTimer(0.125, []string{"sequencer2", "adsr2"}, env.Config, ""))

	steps := []adsr.ADSRStep{
		{LevelRatio: 1.0, DurationRatio: 0.1, Shape: 0.1},
		{LevelRatio: 0.4, DurationRatio: 0.1, Shape: -0.1},
		{DurationRatio: 0.1},
		{DurationRatio: 0.3, Shape: -0.1},
	}

	adsrEnv1 := env.AddModule(adsr.NewADSRModule(steps, 1.0, env.Config, "adsr1"))
	adsrEnv2 := env.AddModule(adsr.NewADSRModule(steps, 1.0, env.Config, "adsr2"))
	mult1 := env.AddModule(common.NewMult(2, env.Config, ""))
	mult2 := env.AddModule(common.NewMult(2, env.Config, ""))
	osc1 := env.AddModule(blosc.NewBloscModule(100.0, 0.0, 1.0, env.Config, "blosc1"))
	osc2 := env.AddModule(blosc.NewBloscModule(400.0, 0.0, 1.0, env.Config, "blosc2"))

	muse.Connect(osc1, 2, mult1, 0)
	muse.Connect(osc2, 2, mult2, 0)
	muse.Connect(adsrEnv1, 0, mult1, 1)
	muse.Connect(adsrEnv2, 0, mult2, 1)
	muse.Connect(mult1, 0, env, 0)
	muse.Connect(mult2, 0, env, 1)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/test.aiff", 15.0)
}

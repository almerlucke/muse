package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/modules/blosc"
)

type ParamMessenger struct {
	ds      []float64
	fs      []float64
	index   int
	address string
}

func (pm *ParamMessenger) Post(timestamp int64, config *muse.Configuration) []*muse.Message {
	msgs := []*muse.Message{}

	if pm.index >= len(pm.ds) {
		return msgs
	}

	sec := float64(timestamp) / config.SampleRate

	if pm.ds[pm.index] < sec {
		msgs = append(msgs, &muse.Message{Address: pm.address, Content: map[string]float64{"frequency": pm.fs[pm.index]}})
		pm.index += 1
	}

	return msgs
}

func main() {
	env := muse.NewEnvironment(2, 44100, 128)
	msgr := &ParamMessenger{
		ds:      []float64{0.0, 1.0, 2.0, 3.0, 4.0},
		fs:      []float64{100, 200, 300, 400, 800},
		address: "blosc2",
	}

	osc1 := blosc.NewBloscModule(100.0, 0.0, 0.1, "blosc1")
	osc2 := blosc.NewBloscModule(400.0, 0.0, 1.0, "blosc2")

	env.AddMessenger(msgr)
	env.AddModule(osc1)
	env.AddModule(osc2)

	muse.Connect(osc1, 0, osc2, 1)
	muse.Connect(osc2, 0, env, 0)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/test.aiff", 5.0)
}

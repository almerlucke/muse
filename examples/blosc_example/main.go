package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/modules/blosc"
)

func main() {
	env := muse.NewEnvironment(2, 44100, 128)

	osc1 := blosc.NewBloscModule(100.0, 0.0, 0.1, "blosc")
	osc2 := blosc.NewBloscModule(400.0, 0.0, 1.0, "blosc")

	env.AddModule(osc1)
	env.AddModule(osc2)

	muse.Connect(osc1, 0, osc2, 1)
	muse.Connect(osc2, 0, env, 0)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/test.aiff", 4.0)
}

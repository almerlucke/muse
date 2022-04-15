package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/modules"
)

func main() {
	env := muse.NewEnvironment(2, 44100, 128)

	osc1 := modules.NewBloscModule(100.0, 0.0, 0.1, "blosc")
	osc2 := modules.NewBloscModule(400.0, 0.0, 1.0, "blosc")

	env.AddModule(osc1)
	env.AddModule(osc2)

	muse.Connect(osc1, 0, osc2, 1)
	muse.Connect(osc2, 0, env, 0)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/test.aiff", 4.0)
}

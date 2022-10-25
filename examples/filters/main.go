package main

import (
	"github.com/almerlucke/muse"
	rbjc "github.com/almerlucke/muse/components/filters/rbj"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/modules/filters/rbj"
	"github.com/almerlucke/muse/modules/noise"
)

func main() {
	env := muse.NewEnvironment(1, 44100.0, 512)

	noise := env.AddModule(noise.NewNoise(0, env.Config))
	filter := env.AddModule(rbj.NewRBJFilter(rbjc.Lowpass, 400.0, 1.8, env.Config, "filter"))

	lfo := env.AddControl(lfo.NewBasicControlLFO(0.1, 50.0, 4000.0, env.Config, ""))

	lfo.ConnectToControl(0, filter, 0)
	muse.Connect(noise, 0, filter, 0)
	muse.Connect(filter, 0, env, 0)

	env.QuickPlayAudio()
}

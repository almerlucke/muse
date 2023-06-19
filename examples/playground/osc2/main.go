package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/modules/blosc"
)

func main() {
	env := muse.NewEnvironment(1, 44100, 512)

	lfo := env.AddControl(lfo.NewBasicControlLFO(0.2, 0.1, 0.9, env.Config, ""))
	osc := env.AddModule(blosc.NewOsc2(100.0, 0, 0.2, 1.0, blosc.MODIFIED_TRIANGLE, env.Config, ""))

	lfo.CtrlConnect(0, osc, 1)
	osc.Connect(0, env, 0)

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/rect.aiff", 2.0, 44100.0, false, sndfile.SF_FORMAT_AIFF)
	env.QuickPlayAudio()
}

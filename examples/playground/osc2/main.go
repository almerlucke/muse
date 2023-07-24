package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/modules/osc"
)

func main() {
	env := muse.NewEnvironment(1)

	lfo := lfo.NewBasicControlLFO(0.2, 0.1, 0.9).CtrlAdd(env)
	osc := osc.NewOsc2(100.0, 0, 0.2, 1.0, osc.MODIFIED_TRIANGLE).Add(env)

	osc.CtrlIn(lfo, 0, 1)
	env.In(osc)

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/rect.aiff", 2.0, 44100.0, false, sndfile.SF_FORMAT_AIFF)
	env.QuickPlayAudio()
}

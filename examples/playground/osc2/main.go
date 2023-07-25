package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/modules/osc"
)

func main() {
	root := muse.New(1)

	lfo := lfo.NewBasicControlLFO(0.2, 0.1, 0.9).CtrlAdd(root)
	osc := osc.NewOsc2(100.0, 0, 0.2, 1.0, osc.MODIFIED_TRIANGLE).Add(root)

	osc.CtrlIn(lfo, 0, 1)
	root.In(osc)

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/rect.aiff", 2.0, 44100.0, false, sndfile.SF_FORMAT_AIFF)
	root.RenderLive()
}

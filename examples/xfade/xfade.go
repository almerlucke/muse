package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/modules/osc"
	"github.com/almerlucke/muse/modules/xfade"
)

func main() {
	root := muse.New(1)

	source1 := root.AddModule(osc.New(110.0, 0.0))
	source2 := root.AddModule(osc.New(140.0, 0.0))
	source3 := root.AddModule(osc.New(160.0, 0.0))
	source4 := root.AddModule(osc.New(170.0, 0.0))

	fader1 := root.AddModule(xfade.New(0.0))
	fader2 := root.AddModule(xfade.New(0.0))
	fader3 := root.AddModule(xfade.New(0.0))

	source1.Connect(2, fader1, 0)
	source2.Connect(3, fader1, 1)
	source3.Connect(0, fader2, 0)
	source4.Connect(1, fader2, 1)
	fader1.Connect(0, fader3, 0)
	fader2.Connect(0, fader3, 1)

	fadeLfo1 := root.AddControl(lfo.NewBasicControlLFO(0.2, 0.0, 1.0))
	fadeLfo2 := root.AddControl(lfo.NewBasicControlLFO(0.31, 0.0, 1.0))
	fadeLfo3 := root.AddControl(lfo.NewBasicControlLFO(0.567, 0.0, 1.0))

	fadeLfo1.CtrlConnect(0, fader1, 0)
	fadeLfo2.CtrlConnect(0, fader2, 0)
	fadeLfo3.CtrlConnect(0, fader3, 0)

	fader3.Connect(0, root, 0)

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/xfade.aiff", 10.0, 44100.0, false, sndfile.SF_FORMAT_AIFF)

	// env.PlotModule(fader3, 0, 10, 1000, 300, "/Users/almerlucke/Desktop/fader.png")

	root.RenderLive()
}

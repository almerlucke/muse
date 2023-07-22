package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/modules/osc"
	"github.com/almerlucke/muse/modules/xfade"
)

func main() {
	env := muse.NewEnvironment(1, 44100.0, 512)

	source1 := env.AddModule(osc.New(110.0, 0.0, env.Config))
	source2 := env.AddModule(osc.New(140.0, 0.0, env.Config))
	source3 := env.AddModule(osc.New(160.0, 0.0, env.Config))
	source4 := env.AddModule(osc.New(170.0, 0.0, env.Config))

	fader1 := env.AddModule(xfade.New(0.0, env.Config))
	fader2 := env.AddModule(xfade.New(0.0, env.Config))
	fader3 := env.AddModule(xfade.New(0.0, env.Config))

	source1.Connect(2, fader1, 0)
	source2.Connect(3, fader1, 1)
	source3.Connect(0, fader2, 0)
	source4.Connect(1, fader2, 1)
	fader1.Connect(0, fader3, 0)
	fader2.Connect(0, fader3, 1)

	fadeLfo1 := env.AddControl(lfo.NewBasicControlLFO(0.2, 0.0, 1.0, env.Config, ""))
	fadeLfo2 := env.AddControl(lfo.NewBasicControlLFO(0.31, 0.0, 1.0, env.Config, ""))
	fadeLfo3 := env.AddControl(lfo.NewBasicControlLFO(0.567, 0.0, 1.0, env.Config, ""))

	fadeLfo1.CtrlConnect(0, fader1, 0)
	fadeLfo2.CtrlConnect(0, fader2, 0)
	fadeLfo3.CtrlConnect(0, fader3, 0)

	fader3.Connect(0, env, 0)

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/xfade.aiff", 10.0, 44100.0, false, sndfile.SF_FORMAT_AIFF)

	env.PlotModule(fader3, 0, 10, 1000, 300, "/Users/almerlucke/Desktop/fader.png")

	env.QuickPlayAudio()
}

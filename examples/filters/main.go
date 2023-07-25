package main

import (
	"github.com/almerlucke/muse"
	rbjc "github.com/almerlucke/muse/components/filters/rbj"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/modules/filters/rbj"
	"github.com/almerlucke/muse/modules/noise"
)

func main() {
	root := muse.New(1)

	noise := root.AddModule(noise.New(0))
	filter := root.AddModule(rbj.New(rbjc.Lowpass, 400.0, 1.8))

	lfo := root.AddControl(lfo.NewBasicControlLFO(0.1, 50.0, 4000.0))

	lfo.CtrlConnect(0, filter, 0)
	noise.CtrlConnect(0, filter, 0)
	filter.CtrlConnect(0, root, 0)

	root.RenderAudio()
}

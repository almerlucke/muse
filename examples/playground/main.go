package main

import (
	"github.com/almerlucke/muse"
	lfo2 "github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/modules/effects/flanger"
	"github.com/almerlucke/muse/modules/player"
	"github.com/almerlucke/sndfile"
)

func main() {
	muse.PushConfiguration(&muse.Configuration{
		SampleRate: 44100.0,
		BufferSize: 1024,
	})

	root := muse.New(1)

	sf, _ := sndfile.NewSoundFile("resources/sounds/elisa.wav")

	pl := player.New(sf, 1.0, 1.0, false).AddTo(root)
	fl := flanger.New(0.3, 0.9, 0.7, false).AddTo(root).In(pl)
	fllfo := lfo2.NewBasicControlLFO(0.4, 0.1, 0.9).CtrlAddTo(root)

	fl.CtrlIn(fllfo)

	root.In(fl)

	_ = root.RenderAudio()
}

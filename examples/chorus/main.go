package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/modules/effects/chorus"
	"github.com/almerlucke/muse/modules/player"
	"github.com/almerlucke/sndfile"
)

func main() {
	root := muse.New(2)

	sf, _ := sndfile.NewSoundFile("resources/sounds/elisa.wav")

	pl := player.New(sf, 1.0, 0.4, false).AddTo(root)

	var ch muse.Module

	if sf.NumChannels() == 2 {
		ch = chorus.New(0.27, 0.4, 0.4, 0.1, 1.0, 0.5, nil).AddTo(root).In(pl, pl, 1)
	} else {
		ch = chorus.New(0.47, 0.5, 0.4, 0.3, 1.0, 0.7, nil).AddTo(root).In(pl)
	}

	root.In(ch, ch, 1)

	_ = root.RenderAudio()
}

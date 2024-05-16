package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/modules/delay"
	"github.com/almerlucke/muse/modules/lfo"
	"github.com/almerlucke/muse/modules/player"
	"github.com/almerlucke/sndfile"
)

func main() {
	muse.PushConfiguration(&muse.Configuration{
		SampleRate: 44100.0,
		BufferSize: 1024,
	})

	root := muse.New(1)

	congaSF, _ := sndfile.NewSoundFile("resources/sounds/elisa.wav")

	pl := player.New(congaSF, 1.0, 1.0, false).AddTo(root)
	dl := delay.New(1000.0, 500.0).AddTo(root)
	dlfo := lfo.New(2.6, 500.0, 530.0).AddTo(root)

	dl.In(pl, dlfo, 0, 1)

	root.In(dl)

	_ = root.RenderAudio()
}

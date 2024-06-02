package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/sndfile/writer"
	"log"
)

func main() {
	muse.PushConfiguration(&muse.Configuration{
		SampleRate: 44100.0,
		BufferSize: 128,
	})

	root := muse.New(2)

	bpm := 120

	drumPatch := createDrumPatch(bpm).AddTo(root)

	root.In(drumPatch, drumPatch, 1)

	err := root.RenderToSoundFile("/home/almer/Documents/backOutside", writer.AIFC, 240, 44100.0, true)
	if err != nil {
		log.Printf("error rendering drums! %v", err)
	}

	//_ = root.RenderAudio()
}

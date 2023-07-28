package main

import (
	"log"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/modules/oversampler"
	"github.com/almerlucke/muse/modules/phasor"
	"github.com/dh1tw/gosamplerate"
)

func main() {
	root := muse.New(1)
	config := muse.CurrentConfiguration()

	muse.PushConfiguration(&muse.Configuration{
		SampleRate: 4 * config.SampleRate,
		BufferSize: config.BufferSize,
	})

	ph := phasor.New(100.0, 0.0)

	muse.PopConfiguration()

	osa, err := oversampler.New(ph, 4, gosamplerate.SRC_SINC_BEST_QUALITY)
	if err != nil {
		log.Fatalf("failed to create oversampler: %v", err)
	}

	osa.Add(root)
	root.In(osa)

	root.RenderToSoundFile("/Users/almerlucke/Desktop/testosa.aiff", 4.0, 44100.0, io.AIFF)
	root.RenderAudio()
}

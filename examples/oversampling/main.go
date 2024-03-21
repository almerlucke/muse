package main

import (
	"github.com/almerlucke/sndfile/writer"
	"log"

	"github.com/almerlucke/muse"
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

	osa, err := oversampler.New(ph, gosamplerate.SRC_SINC_BEST_QUALITY)
	if err != nil {
		log.Fatalf("failed to create oversampler: %v", err)
	}

	osa.AddTo(root)
	root.In(osa)

	_ = root.RenderToSoundFile("/Users/almerlucke/Desktop/testosa", writer.AIFC, 4.0, 44100.0, false)
	//root.RenderAudio()
}

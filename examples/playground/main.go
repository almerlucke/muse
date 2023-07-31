package main

import (
	"log"

	"github.com/almerlucke/muse/io/sndfile/aifc"
)

func main() {
	// root := muse.New(1)
	// config := muse.CurrentConfiguration()

	// muse.PushConfiguration(&muse.Configuration{
	// 	SampleRate: 4 * config.SampleRate,
	// 	BufferSize: config.BufferSize,
	// })

	// ph := phasor.New(100.0, 0.0)

	// muse.PopConfiguration()

	// osa, err := oversampler.New(ph, 4, gosamplerate.SRC_SINC_BEST_QUALITY)
	// if err != nil {
	// 	log.Fatalf("failed to create oversampler: %v", err)
	// }

	// osa.Add(root)
	// root.In(osa)

	// root.RenderToSoundFile("/Users/almerlucke/Desktop/testosa.aiff", 4.0, 44100.0, io.AIFF)
	// root.RenderAudio()

	file, err := aifc.Open("/Users/almerlucke/Desktop/test.aifc", 1, 44100.0)
	if err != nil {
		log.Fatalf("err opening aifc file: %v", err)
	}

	defer file.Close()

	items := make([]float64, 512)
	for i := 0; i < 512; i++ {
		items[i] = float64(i) / 512.0
		log.Printf("items[i] %v", items[i])
	}

	// 2132
	// 512 * 4 = 2048

	// 44
	// num channels = 1
	// number of frames = 512
	// sample size = 32
	// Sample rate = 00000 00000 30303030303030303030

	// 40 0E AC 44 00 00 00 00 00 00

	file.WriteItems(items)
}

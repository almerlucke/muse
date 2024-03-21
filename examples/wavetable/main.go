package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/phasor"
	"github.com/almerlucke/muse/controls/gen"
	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/modules/wtosc"
	"log"
)

func main() {
	root := muse.New(1)

	sf, err := io.NewWaveTableSoundFile("/home/almer/Downloads/KRC Free VAE2 Wav Format/orbit_vae_V1.07_dims24_rad_1.5_6_3_69_free.wav", 2048)
	if err != nil {
		log.Fatalf("err loading sound file: %v", err)
	}

	ctrlRate := root.Config.SampleRate / float64(root.Config.BufferSize)

	g := gen.NewGen(phasor.New(0.05, ctrlRate, 0.0), nil, nil)
	g.CtrlAddTo(root)

	osc := wtosc.New(sf, 30.0, 0.0, 0.0, 0.4).AddTo(root)
	osc.CtrlIn(g, 0, 2)

	root.In(osc)

	err = root.RenderAudio()
	if err != nil {
		log.Fatalf("err rendering audio: %v", err)
	}
}

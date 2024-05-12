package main

import (
	"github.com/almerlucke/genny/float/phasor"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/controls/gen"
	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/modules/wtosc"
	"log"
)

func main() {
	root := muse.New(1)

	sf, err := io.NewWaveTableSoundFile("/home/almer/Music/wavetables/Free Wavetables[2048]/Wavetrables[2048samples-44.1khz-32bit]/Melda Oscillator[2048-44.1khz-32bit]/Melda CustumWave3 IcarusDenoise.wav", 2048)
	if err != nil {
		log.Fatalf("err loading sound file: %v", err)
	}

	ctrlRate := root.Config.SampleRate / float64(root.Config.BufferSize)

	g := gen.New[float64](phasor.New(0.05, ctrlRate, 0.2), true)
	g.CtrlAddTo(root)

	osc := wtosc.New(sf, 60.0, 0.0, 0.44, 0.4).AddTo(root)
	osc.CtrlIn(g, 0, 2)

	root.In(osc)

	//err = root.RenderToSoundFile("/home/almer/Music/wavetable_test", writer.AIFC, 5.0, 44100.0, false)
	//if err != nil {
	//	log.Fatalf("err rendering audio: %v", err)
	//}
	err = root.RenderAudio()
	if err != nil {
		log.Fatalf("err rendering audio: %v", err)
	}
}

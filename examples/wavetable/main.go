package main

import (
	"github.com/almerlucke/genny/float"
	"github.com/almerlucke/genny/float/interp"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/controls/gen"
	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/modules/wtscan"
	"log"
)

func main() {
	root := muse.New(1)

	// sf, err := io.NewWaveTableSoundFile("/home/almer/Music/wavetables/Free Wavetables[2048]/Wavetrables[2048samples-44.1khz-32bit]/K-Devicees Terra[2048-44.1khz-32bit]/Terra Bend+PD Stack.wav", 2048)
	sf, err := io.NewWaveTableSoundFile("/home/almer/Music/wavetables/Free Wavetables[2048]/Wavetrables[2048samples-44.1khz-32bit]/Prophet VS[2048-44.1khz-32bit]/VS Morph3.wav", 2048)
	if err != nil {
		log.Fatalf("err loading sound file: %v", err)
	}

	sc := wtscan.New(sf, 160.0, 0.0, 0.0, 0.4).AddTo(root)

	gc := gen.New[float64](float.FromFrame(interp.New(float.ToFrame(function.NewRandom(0.0, 1.0)), interp.Linear, 0.0063), 0), true).CtrlAddTo(root)

	sc.CtrlIn(gc, 0, 2)

	root.In(sc)

	//err = root.RenderToSoundFile("/home/almer/Music/wavetable_test", writer.AIFC, 5.0, 44100.0, false)
	//if err != nil {
	//	log.Fatalf("err rendering audio: %v", err)
	//}

	err = root.RenderAudio()
	if err != nil {
		log.Fatalf("err rendering audio: %v", err)
	}
}

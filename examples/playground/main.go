package main

import (
	"log"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/controls/val"
	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/messengers/triggers/timer"
	"github.com/almerlucke/muse/modules/player"
	"github.com/almerlucke/muse/value"
)

func notMipMapped() {
	env := muse.NewEnvironment(2, 44100, 512)

	mm, err := io.NewSoundFile("resources/sounds/mixkit-angelical-choir-654.wav")
	if err != nil {
		log.Fatalf("err loading file: %v", err)
	}

	timer := env.AddControl(timer.NewControlTimer(1000.0, env.Config, ""))
	player := env.AddModule(player.NewPlayer(mm, 4.0, 1.0, true, env.Config, ""))

	timer.CtrlConnect(0, player, 0)

	player.Connect(0, env, 0)
	player.Connect(1, env, 1)

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/rect.aiff", 2.0, 44100.0, false, sndfile.SF_FORMAT_AIFF)
	env.QuickPlayAudio()
}

func mipMapped() {
	env := muse.NewEnvironment(2, 44100, 512)

	sf, err := io.NewMipMapSoundFile("resources/sounds/elisa.wav", 5)
	if err != nil {
		log.Fatalf("err loading mipmap: %v", err)
	}

	timer := env.AddControl(timer.NewControlTimer(1500.0, env.Config, ""))
	speed := val.NewVal[float64](value.NewSequence([]float64{1.0, 1.25, 1.25, 1.5, 0.75, 1.25}), "")
	startOffset := val.NewVal[float64](value.NewSequence([]float64{3.0, 9.0, 20.0, 27.0, 31.0}), "")
	endOffset := val.NewVal[float64](value.NewSequence([]float64{6.0, 12.0, 23.0, 30.0, 33.0}), "")
	player := env.AddModule(player.NewPlayerX(sf, 1.0, 1.0, 20.0, 23.5, true, env.Config, ""))

	timer.CtrlConnect(0, speed, 0)
	timer.CtrlConnect(0, startOffset, 0)
	timer.CtrlConnect(0, endOffset, 0)
	speed.CtrlConnect(0, player, 1)
	startOffset.CtrlConnect(0, player, 2)
	endOffset.CtrlConnect(0, player, 3)
	timer.CtrlConnect(0, player, 0)

	player.Connect(0, env, 0)
	player.Connect(1, env, 1)

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/rect.aiff", 2.0, 44100.0, false, sndfile.SF_FORMAT_AIFF)
	env.QuickPlayAudio()
}

func main() {
	mipMapped()
}

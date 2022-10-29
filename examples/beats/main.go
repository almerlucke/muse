package main

import (
	"math/rand"
	"time"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"

	"github.com/almerlucke/muse/modules/player"
	"github.com/almerlucke/muse/value"
	"github.com/almerlucke/muse/value/template"
)

type DrumKit struct {
	HiHat string
	Kick  string
	Snare string
	Ride  string
}

func addDrumTrack(env *muse.Environment, moduleName string, soundBuffer *io.SoundFileBuffer, tempo int, division int, lowSpeed float64, highSpeed float64, amp float64, steps value.Valuer[*swing.Step]) muse.Module {
	identifier := moduleName + "Speed"

	env.AddMessenger(stepper.NewStepper(swing.New(tempo, division, steps), []string{identifier}, ""))

	env.AddMessenger(banger.NewTemplateGenerator([]string{moduleName}, template.Template{
		"speed": value.NewFunction(func() any { return rand.Float64()*(highSpeed-lowSpeed) + lowSpeed }),
		"bang":  true,
	}, identifier))

	return env.AddModule(player.NewPlayer(soundBuffer, 1.0, amp, true, env.Config, moduleName))
}

func main() {
	rand.Seed(time.Now().UnixNano())
	env := muse.NewEnvironment(1, 44100.0, 512)

	bpm := 80

	hihatSound, _ := io.NewSoundFileBuffer("resources/drums/hihat/Cymatics - Humble Closed Hihat 1.wav")
	kickSound, _ := io.NewSoundFileBuffer("resources/drums/kick/Cymatics - Humble Sit Down Kick - D.wav")
	snareSound, _ := io.NewSoundFileBuffer("resources/drums/snare/Cymatics - Humble Theory Snare - D#.wav")

	hihatPlayer := addDrumTrack(env, "hihat", hihatSound, bpm, 4, 0.875, 1.125, 1.0, value.NewSequence([]*swing.Step{
		{Skip: true}, {}, {Skip: true}, {Shuffle: 0.05, ShuffleRand: 0.05}, {Skip: true}, {BurstChance: 0.6, NumBurst: 3}, {Skip: true},
		{}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true},
		{}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true},
		{}, {Skip: true}, {}, {Skip: true}, {Multiply: 0.25}, {Multiply: 0.25}, {Multiply: 0.25}, {Multiply: 0.25},
		{}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true},
		{}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {},
	}))

	kickPlayer := addDrumTrack(env, "kick", kickSound, bpm, 4, 0.875, 1.125, 1.0, value.NewSequence([]*swing.Step{
		{}, {Skip: true},
	}))

	snarePlayer := addDrumTrack(env, "snare", snareSound, bpm, 4, 0.875, 1.125, 1.0, value.NewSequence([]*swing.Step{
		{Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true},
		{Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {BurstChance: 0.6, NumBurst: 3}, {Skip: true},
	}))

	kickPlayer.Connect(0, env, 0)
	hihatPlayer.Connect(0, env, 0)
	snarePlayer.Connect(0, env, 0)

	env.QuickPlayAudio()
}

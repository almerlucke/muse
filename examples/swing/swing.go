package main

import (
	"math/rand"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/player"
	"github.com/almerlucke/muse/value"
	"github.com/almerlucke/muse/value/template"
)

func addDrumTrack(env *muse.Environment, moduleName string, soundBuffer *io.SoundFileBuffer, tempo int, division int, lowSpeed float64, highSpeed float64, amp float64, steps value.Valuer[*swing.Step]) muse.Module {
	identifier := moduleName + "Speed"

	env.AddMessenger(stepper.NewStepper(swing.New(tempo, division, steps), []string{identifier}, ""))

	env.AddMessenger(banger.NewTemplateGenerator([]string{moduleName}, template.Template{
		"speed": value.NewFunction(func() any { return rand.Float64()*(highSpeed-lowSpeed) + lowSpeed }),
	}, identifier))

	return env.AddModule(player.NewPlayer(soundBuffer, 1.0, amp, true, env.Config, moduleName))
}

func main() {
	env := muse.NewEnvironment(1, 44100.0, 512)

	hihatSound, _ := io.NewSoundFileBuffer("resources/drums/hihat/Cymatics - Humble Closed Hihat 1.wav")
	kickSound, _ := io.NewSoundFileBuffer("resources/drums/kick/Cymatics - Humble Triple Kick - E.wav")
	snareSound, _ := io.NewSoundFileBuffer("resources/drums/clap/Cymatics - Humble Stars Clap.wav")
	bassSound, _ := io.NewSoundFileBuffer("resources/drums/808/Cymatics - Humble 808 5 - G.wav")
	rideSound, _ := io.NewSoundFileBuffer("resources/drums/hihat/Cymatics - Humble Open Hihat 2.wav")
	waterSound, _ := io.NewSoundFileBuffer("resources/sounds/Cymatics - Orchid Live Recording - Waves.wav")
	swirlSound, _ := io.NewSoundFileBuffer("resources/sounds/Cymatics - Orchid KEYS Swirl (C).wav")
	vocalSound, _ := io.NewSoundFileBuffer("resources/sounds/Cymatics - Blurry Vocal - 80 BPM F Min.wav")

	bpm := 80

	hihatPlayer := addDrumTrack(env, "hihat", hihatSound, bpm, 8, 1.875, 2.125, 0.75, value.NewSequence([]*swing.Step{
		{}, {Shuffle: 0.1}, {SkipFactor: 0.8, BurstFactor: 1.0, NumBurst: 3}, {Shuffle: 0.1, ShuffleRand: 0.1}, {Skip: true}, {Shuffle: 0.1}, {}, {SkipFactor: 0.1, Shuffle: 0.1}, {Skip: true}, {Skip: true},
	}))

	kickPlayer := addDrumTrack(env, "kick", kickSound, bpm, 4, 0.875, 1.125, 1.0, value.NewSequence([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {SkipFactor: 0.3, Shuffle: 0.05, ShuffleRand: 0.05}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}))

	snarePlayer := addDrumTrack(env, "snare", snareSound, bpm, 4, 0.875, 1.125, 1.0, value.NewSequence([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Shuffle: 0.05, ShuffleRand: 0.05}, {Skip: true}, {Skip: true}, {Skip: true},
	}))

	bassPlayer := addDrumTrack(env, "bass", bassSound, bpm, 1, 0.875, 1.125, 1.0, value.NewSequence([]*swing.Step{
		{Shuffle: 0.2, ShuffleRand: 0.2}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}))

	ridePlayer := addDrumTrack(env, "ride", rideSound, bpm, 2, 0.875, 1.25, 0.3, value.NewSequence([]*swing.Step{
		{Skip: true}, {Skip: true}, {Shuffle: 0.05, ShuffleRand: 0.05, BurstFactor: 0.2, NumBurst: 4}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}))

	waterPlayer := addDrumTrack(env, "water", waterSound, int(float64(bpm)*0.125), 2, 0.875, 1.25, 0.3, value.NewSequence([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}))

	swirlPlayer := addDrumTrack(env, "swirl", swirlSound, int(float64(bpm)*0.5), 1.0, 0.875, 1.25, 0.2, value.NewSequence([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}))

	vocalPlayer := addDrumTrack(env, "vocal", vocalSound, int(float64(bpm)*0.125), 1.0, 0.975, 1.025, 0.2, value.NewSequence([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}))

	mult := env.AddModule(functor.NewFunctor(1, func(v []float64) float64 { return v[0] * 0.5 }, env.Config))

	muse.Connect(kickPlayer, 0, mult, 0)
	muse.Connect(hihatPlayer, 0, mult, 0)
	muse.Connect(snarePlayer, 0, mult, 0)
	muse.Connect(bassPlayer, 0, mult, 0)
	muse.Connect(ridePlayer, 0, mult, 0)
	muse.Connect(waterPlayer, 0, mult, 0)
	muse.Connect(swirlPlayer, 0, mult, 0)
	muse.Connect(vocalPlayer, 0, mult, 0)
	muse.Connect(mult, 0, env, 0)

	env.QuickPlayAudio()
}

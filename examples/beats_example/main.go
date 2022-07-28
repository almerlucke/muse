package main

import (
	"math/rand"
	"time"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/messengers/generators/prototype"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/mkb218/gosndfile/sndfile"

	"github.com/almerlucke/muse/modules/player"
	"github.com/almerlucke/muse/values"
)

func addRhythm(env *muse.Environment, module string, tempo float64, division float64, lowSpeed float64, highSpeed float64, steps values.Generator[*swing.Step]) {
	identifier := module + "Speed"

	env.AddMessenger(stepper.NewStepper(swing.New(tempo, division, steps), []string{identifier}, ""))

	env.AddMessenger(prototype.NewPrototypeGenerator([]string{module}, values.MapPrototype{
		"speed": values.NewFunction(func() any { return rand.Float64()*(highSpeed-lowSpeed) + lowSpeed }),
	}, identifier))
}

func main() {
	rand.Seed(time.Now().UnixNano())
	env := muse.NewEnvironment(1, 44100.0, 512)

	hihatSound, _ := io.NewSoundFileBuffer("/Users/almerlucke/Documents/Private/Sounds/Cymatics - Humble Hip Hop Sample Pack/Drums - One Shots/Cymbals/Cymatics - Humble Closed Hihat 1.wav")
	kickSound, _ := io.NewSoundFileBuffer("/Users/almerlucke/Documents/Private/Sounds/Cymatics - Humble Hip Hop Sample Pack/Drums - One Shots/Kicks/Cymatics - Humble You Kick - A.wav")
	snareSound, _ := io.NewSoundFileBuffer("/Users/almerlucke/Documents/Private/Sounds/Cymatics - Humble Hip Hop Sample Pack/Drums - One Shots/Snare/Cymatics - Humble Friday Snare - E.wav")

	addRhythm(env, "hihat", 120.0, 4.0, 0.875, 1.125, values.NewSequence([]*swing.Step{
		{}, {Shuffle: 0.3}, {Skip: true}, {Shuffle: 0.3, ShuffleRand: 0.2}, {Skip: true}, {Shuffle: 0.1}, {}, {SkipFactor: 0.4, Shuffle: 0.2}, {Skip: true}, {Skip: true},
	}, true))

	hihatPlayer := env.AddModule(player.NewPlayer(hihatSound, 1.0, true, env.Config, "hihat"))

	addRhythm(env, "kick", 120.0, 4.0, 0.875, 1.125, values.NewSequence([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {SkipFactor: 0.4}, {Shuffle: 0.2}, {Skip: true}, {Skip: true},
	}, true))

	kickPlayer := env.AddModule(player.NewPlayer(kickSound, 1.0, true, env.Config, "kick"))

	addRhythm(env, "snare", 120.0, 2.0, 0.875, 1.125, values.NewSequence([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Shuffle: 0.1, ShuffleRand: 0.1},
	}, true))

	snarePlayer := env.AddModule(player.NewPlayer(snareSound, 1.0, true, env.Config, "snare"))

	muse.Connect(kickPlayer, 0, env, 0)
	muse.Connect(hihatPlayer, 0, env, 0)
	muse.Connect(snarePlayer, 0, env, 0)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/beats.aiff", 20.0, env.Config.SampleRate, sndfile.SF_FORMAT_AIFF)
}

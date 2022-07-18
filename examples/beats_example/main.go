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
	params "github.com/almerlucke/muse/parameters"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	env := muse.NewEnvironment(1, 44100.0, 512)

	hihatSound, _ := io.NewSoundFileBuffer("/Users/almerlucke/Documents/Private/Sounds/Cymatics - Humble Hip Hop Sample Pack/Drums - One Shots/Cymbals/Cymatics - Humble Closed Hihat 1.wav")
	kickSound, _ := io.NewSoundFileBuffer("/Users/almerlucke/Documents/Private/Sounds/Cymatics - Humble Hip Hop Sample Pack/Drums - One Shots/Kicks/Cymatics - Humble You Kick - A.wav")
	snareSound, _ := io.NewSoundFileBuffer("/Users/almerlucke/Documents/Private/Sounds/Cymatics - Humble Hip Hop Sample Pack/Drums - One Shots/Snare/Cymatics - Humble Friday Snare - E.wav")

	hihatPlayer := env.AddModule(player.NewPlayer(hihatSound, 1.0, true, env.Config, "hihat"))
	kickPlayer := env.AddModule(player.NewPlayer(kickSound, 1.0, true, env.Config, "kick"))
	snarePlayer := env.AddModule(player.NewPlayer(snareSound, 1.0, true, env.Config, "snare"))

	env.AddMessenger(stepper.NewStepper(swing.New(120.0, 4.0, []*swing.Step{
		{}, {Shuffle: 0.3}, {Skip: true}, {Shuffle: 0.3, ShuffleRand: 0.2}, {Skip: true}, {Shuffle: 0.1}, {}, {SkipFactor: 0.4, Shuffle: 0.2}, {Skip: true}, {Skip: true},
	}), []string{"hihatSpeed"}, ""))

	env.AddMessenger(env.AddMessenger(prototype.NewPrototypeGenerator([]string{"hihat"}, params.Prototype{
		"speed": params.NewFunction(func() any { return rand.Float64()*0.25 + 0.875 }),
	}, "hihatSpeed")))

	env.AddMessenger(stepper.NewStepper(swing.New(120.0, 4.0, []*swing.Step{
		{}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {SkipFactor: 0.4}, {Shuffle: 0.2}, {Skip: true}, {Skip: true},
	}), []string{"kickSpeed"}, ""))

	env.AddMessenger(env.AddMessenger(prototype.NewPrototypeGenerator([]string{"kick"}, params.Prototype{
		"speed": params.NewFunction(func() any { return rand.Float64()*0.25 + 0.875 }),
	}, "kickSpeed")))

	env.AddMessenger(stepper.NewStepper(swing.New(120.0, 2.0, []*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Shuffle: 0.1, ShuffleRand: 0.1},
	}), []string{"snareSpeed"}, ""))

	env.AddMessenger(env.AddMessenger(prototype.NewPrototypeGenerator([]string{"snare"}, params.Prototype{
		"speed": params.NewFunction(func() any { return rand.Float64()*0.25 + 0.875 }),
	}, "snareSpeed")))

	muse.Connect(kickPlayer, 0, env, 0)
	muse.Connect(hihatPlayer, 0, env, 0)
	muse.Connect(snarePlayer, 0, env, 0)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/beats.aiff", 20.0, env.Config.SampleRate, sndfile.SF_FORMAT_AIFF)
}

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
	"github.com/almerlucke/muse/value/markov"
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

func kickRhythm() value.Valuer[*swing.Step] {
	rhythm1 := markov.NewState([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}, nil)

	rhythm2 := markov.NewState([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
		{Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}, nil)

	rhythm3 := markov.NewState([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
		{Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true},
	}, nil)

	rhythm4 := markov.NewState([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true},
		{Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true},
	}, nil)

	rhythm5 := markov.NewState([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {}, {Skip: true},
		{Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true},
	}, nil)

	rhythm1.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm1, 3.0, rhythm2, 2.0, rhythm3, 1.0)
	rhythm2.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm2, 3.0, rhythm3, 2.0, rhythm4, 1.0)
	rhythm3.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm3, 3.0, rhythm4, 2.0, rhythm5, 1.0)
	rhythm4.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm4, 4.0, rhythm5, 2.0, rhythm1, 1.0)
	rhythm5.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm5, 4.0, rhythm1, 2.0, rhythm2, 1.0)

	m := markov.NewMarkov[[]*swing.Step](markov.NewStateStarter(rhythm1), 1)

	return value.NewFlatten[*swing.Step](m)
}

func snareRhythm() value.Valuer[*swing.Step] {
	rhythm1 := markov.NewState([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true},
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}, nil)

	rhythm2 := markov.NewState([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true},
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true},
	}, nil)

	rhythm3 := markov.NewState([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true},
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {BurstChance: 0.5, NumBurst: 3}, {Skip: true}, {Skip: true}, {Skip: true},
	}, nil)

	rhythm4 := markov.NewState([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {},
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {BurstChance: 0.5, NumBurst: 3}, {Skip: true}, {Skip: true}, {Skip: true},
	}, nil)

	rhythm5 := markov.NewState([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {BurstChance: 0.5, NumBurst: 3},
		{Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {},
	}, nil)

	rhythm1.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm1, 3.0, rhythm2, 2.0, rhythm3, 1.0)
	rhythm2.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm2, 3.0, rhythm3, 2.0, rhythm4, 1.0)
	rhythm3.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm3, 3.0, rhythm4, 2.0, rhythm5, 1.0)
	rhythm4.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm4, 4.0, rhythm5, 2.0, rhythm1, 1.0)
	rhythm5.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm5, 4.0, rhythm1, 2.0, rhythm2, 1.0)

	m := markov.NewMarkov[[]*swing.Step](markov.NewStateStarter(rhythm1), 1)

	return value.NewFlatten[*swing.Step](m)
}

func hihatRhythm() value.Valuer[*swing.Step] {
	rhythm1 := markov.NewState([]*swing.Step{
		{}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true},
	}, nil)

	rhythm2 := markov.NewState([]*swing.Step{
		{BurstChance: 0.5, NumBurst: 3}, {Shuffle: 0.1}, {}, {Skip: true}, {}, {Shuffle: 0.1}, {}, {Skip: true},
	}, nil)

	rhythm3 := markov.NewState([]*swing.Step{
		{BurstChance: 0.5, NumBurst: 3}, {Shuffle: 0.1}, {Skip: true}, {Shuffle: 0.1}, {}, {Shuffle: 0.1}, {Skip: true}, {Shuffle: 0.1},
	}, nil)

	rhythm4 := markov.NewState([]*swing.Step{
		{BurstChance: 0.5, NumBurst: 3}, {Skip: true}, {}, {Shuffle: 0.1}, {}, {Skip: true}, {}, {Shuffle: 0.1},
	}, nil)

	rhythm5 := markov.NewState([]*swing.Step{
		{}, {Shuffle: 0.1}, {}, {Shuffle: 0.1}, {}, {Shuffle: 0.1}, {BurstChance: 0.5, NumBurst: 3}, {Shuffle: 0.1},
	}, nil)

	rhythm1.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm1, 2.0, rhythm2, 1.0, rhythm3, 1.0)
	rhythm2.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm2, 2.0, rhythm3, 1.0, rhythm4, 1.0)
	rhythm3.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm3, 2.0, rhythm4, 1.0, rhythm5, 1.0)
	rhythm4.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm4, 3.0, rhythm5, 1.0, rhythm1, 1.0)
	rhythm5.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm5, 3.0, rhythm1, 1.0, rhythm2, 1.0)

	m := markov.NewMarkov[[]*swing.Step](markov.NewStateStarter(rhythm1), 1)

	return value.NewFlatten[*swing.Step](m)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	env := muse.NewEnvironment(1, 44100.0, 512)

	bpm := 80

	hihatSound, _ := io.NewSoundFileBuffer("resources/drums/hihat/Cymatics - Humble Closed Hihat 1.wav")
	kickSound, _ := io.NewSoundFileBuffer("resources/drums/kick/Cymatics - Humble Triple Kick - E.wav")
	snareSound, _ := io.NewSoundFileBuffer("resources/drums/shots/Cymatics - Orchid Snap - Single.wav")
	shakerSound, _ := io.NewSoundFileBuffer("resources/drums/shots/Cymatics - Orchid Snap - Cream.wav")
	fx1Sound, _ := io.NewSoundFileBuffer("resources/drums/shots/Cymatics - Orchid Shaker - Drew.wav")

	hihatPlayer := addDrumTrack(env, "hihat", hihatSound, bpm, 4, 0.875, 1.125, 0.6, hihatRhythm())

	kickPlayer := addDrumTrack(env, "kick", kickSound, bpm, 8, 0.875, 1.125, 1.0, kickRhythm())

	snarePlayer := addDrumTrack(env, "snare", snareSound, bpm, 8, 1.0, 1.0, 1.0, snareRhythm())

	shakerPlayer := addDrumTrack(env, "shaker", shakerSound, bpm, 2, 1.0, 1.0, 1.0, snareRhythm())

	fx1Player := addDrumTrack(env, "fx1", fx1Sound, bpm, 2, 1.0, 1.0, 1.0, kickRhythm())

	kickPlayer.Connect(0, env, 0)
	hihatPlayer.Connect(0, env, 0)
	snarePlayer.Connect(0, env, 0)
	shakerPlayer.Connect(0, env, 0)
	fx1Player.Connect(0, env, 0)

	env.QuickPlayAudio()
}

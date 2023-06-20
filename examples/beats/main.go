package main

import (
	"math/rand"
	"time"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/synths/drums"
	"github.com/almerlucke/muse/utils"

	"github.com/almerlucke/muse/value"
	"github.com/almerlucke/muse/value/markov"
	"github.com/almerlucke/muse/value/template"
)

func addDrumTrack(env *muse.Environment, polyName string, sounds []string, tempo int, division int, lowSpeed float64, highSpeed float64, amp float64, steps value.Valuer[*swing.Step]) {
	identifier := sounds[0] + "Drum"

	env.AddMessenger(stepper.NewStepper(swing.New(tempo, division, steps), []string{identifier}, ""))

	env.AddMessenger(banger.NewTemplateGenerator([]string{polyName}, template.Template{
		"command":   "trigger",
		"duration":  0.0,
		"amplitude": amp,
		"message": template.Template{
			"speed": value.NewFunction(func() any { return rand.Float64()*(highSpeed-lowSpeed) + lowSpeed }),
			"sound": value.NewSequence(utils.ToAnySlice(sounds)),
		},
	}, identifier))
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

func bassRhythm() value.Valuer[*swing.Step] {
	rhythm1 := markov.NewState([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true},
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}, nil)

	rhythm2 := markov.NewState([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true},
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true},
	}, nil)

	rhythm3 := markov.NewState([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {},
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true},
	}, nil)

	rhythm4 := markov.NewState([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {},
		{Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {},
	}, nil)

	rhythm1.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm1, 3.0, rhythm2, 2.0, rhythm3, 1.0)
	rhythm2.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm2, 3.0, rhythm3, 2.0, rhythm4, 1.0)
	rhythm3.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm3, 3.0, rhythm4, 2.0, rhythm1, 1.0)
	rhythm4.Transitioner = markov.NewProbabilityTransitionerVariadic[[]*swing.Step](rhythm4, 4.0, rhythm1, 2.0, rhythm2, 1.0)

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

	env := muse.NewEnvironment(2, 44100.0, 512)

	bpm := 80

	soundBank := io.SoundBank{}

	soundBank["hihat"], _ = io.NewSoundFile("resources/drums/hihat/Cymatics - Humble Closed Hihat 1.wav")
	soundBank["kick"], _ = io.NewSoundFile("resources/drums/kick/Cymatics - Humble Triple Kick - E.wav")
	soundBank["snare"], _ = io.NewSoundFile("resources/drums/snare/Cymatics - Humble Adequate Snare - E.wav")
	soundBank["808_1"], _ = io.NewSoundFile("resources/drums/808/Cymatics - Humble 808 4 - F.wav")
	soundBank["808_2"], _ = io.NewSoundFile("resources/drums/808/Cymatics - Humble 808 3 - F.wav")
	soundBank["808_3"], _ = io.NewSoundFile("resources/drums/fx/Cymatics - Orchid Impact FX 2.wav")
	soundBank["808_4"], _ = io.NewSoundFile("resources/drums/fx/Cymatics - Orchid Reverse Crash 2.wav")
	soundBank["shaker"], _ = io.NewSoundFile("resources/drums/shots/Cymatics - Orchid Shaker - Drew.wav")

	drums := env.AddModule(drums.NewDrums(soundBank, 20, env.Config, "drums"))

	addDrumTrack(env, "drums", []string{"hihat"}, bpm, 8, 0.875, 1.125, 0.6, hihatRhythm())
	addDrumTrack(env, "drums", []string{"kick"}, bpm, 8, 0.875, 1.125, 1.0, kickRhythm())
	addDrumTrack(env, "drums", []string{"snare"}, bpm, 8, 1.0, 1.0, 0.7, snareRhythm())
	addDrumTrack(env, "drums", []string{"808_1", "808_2", "808_3", "808_4"}, bpm, 2, 1.0, 1.0, 0.3, bassRhythm())
	addDrumTrack(env, "drums", []string{"shaker"}, bpm, 2, 1.0, 1.0, 1.0, kickRhythm())

	drums.Connect(0, env, 0)
	drums.Connect(1, env, 1)

	env.QuickPlayAudio()
}

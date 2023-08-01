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

func addDrumTrack(p muse.Patch, polyName string, sounds []string, tempo int, division int, lowSpeed float64, highSpeed float64, amp float64, steps value.Valuer[*swing.Step]) {
	identifier := sounds[0] + "Drum"

	p.AddMessenger(stepper.NewStepper(swing.New(tempo, division, steps), []string{identifier}))

	p.AddMessenger(banger.NewTemplateGenerator([]string{polyName}, template.Template{
		"command":   "trigger",
		"duration":  0.0,
		"amplitude": amp,
		"message": template.Template{
			"speed": value.NewFunction(func() any { return rand.Float64()*(highSpeed-lowSpeed) + lowSpeed }),
			"sound": value.NewSequence(utils.ToAnySlice(sounds)),
		},
	}).MsgrNamed(identifier))
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

	m := muse.New(2)

	bpm := 80

	soundBank := io.SoundBank{}

	soundBank["hihat"], _ = io.NewMipMapSoundFile("resources/drums/hihat/Cymatics - Humble Closed Hihat 1.wav", 4)
	soundBank["kick"], _ = io.NewMipMapSoundFile("resources/drums/kick/Cymatics - Humble Triple Kick - E.wav", 4)
	soundBank["snare"], _ = io.NewMipMapSoundFile("resources/drums/snare/Cymatics - Humble Adequate Snare - E.wav", 4)
	soundBank["808_1"], _ = io.NewMipMapSoundFile("resources/drums/808/Cymatics - Humble 808 4 - F.wav", 4)
	soundBank["808_2"], _ = io.NewMipMapSoundFile("resources/drums/808/Cymatics - Humble 808 3 - F.wav", 4)
	soundBank["808_3"], _ = io.NewMipMapSoundFile("resources/drums/fx/Cymatics - Orchid Impact FX 2.wav", 4)
	soundBank["808_4"], _ = io.NewMipMapSoundFile("resources/drums/fx/Cymatics - Orchid Reverse Crash 2.wav", 4)
	soundBank["shaker"], _ = io.NewMipMapSoundFile("resources/drums/shots/Cymatics - Orchid Shaker - Drew.wav", 4)

	drums := drums.NewDrums(soundBank, 20).Named("drums").AddTo(m)

	addDrumTrack(m, "drums", []string{"hihat"}, bpm, 8, 0.575, 3.525, 0.6, hihatRhythm())
	addDrumTrack(m, "drums", []string{"kick"}, bpm, 8, 0.575, 3.525, 1.0, kickRhythm())
	addDrumTrack(m, "drums", []string{"snare"}, bpm, 8, 0.6, 3.4, 0.7, snareRhythm())
	addDrumTrack(m, "drums", []string{"808_1", "808_2", "808_3", "808_4"}, bpm, 2, 0.6, 4.0, 0.3, bassRhythm())
	addDrumTrack(m, "drums", []string{"shaker"}, bpm, 2, 0.6, 4.0, 1.0, kickRhythm())

	m.In(drums, drums, 1)

	m.RenderAudio()
}

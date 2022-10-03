package main

import (
	"math/rand"
	"time"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/mkb218/gosndfile/sndfile"

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

func addDrumTrack(env *muse.Environment, moduleName string, soundBuffer *io.SoundFileBuffer, tempo float64, division float64, lowSpeed float64, highSpeed float64, amp float64, steps value.Valuer[*swing.Step]) muse.Module {
	identifier := moduleName + "Speed"

	env.AddMessenger(stepper.NewStepper(swing.New(value.NewConst(tempo), value.NewConst(division), steps), []string{identifier}, ""))

	env.AddMessenger(banger.NewTemplateGenerator([]string{moduleName}, template.Template{
		"speed": value.NewFunction(func() any { return rand.Float64()*(highSpeed-lowSpeed) + lowSpeed }),
	}, identifier))

	return env.AddModule(player.NewPlayer(soundBuffer, 1.0, amp, true, env.Config, moduleName))
}

func main() {
	rand.Seed(time.Now().UnixNano())
	env := muse.NewEnvironment(1, 44100.0, 512)

	bpm := 80.0

	hihatSound, _ := io.NewSoundFileBuffer("resources/drums/hihat/Cymatics - Humble Closed Hihat 1.wav")
	kickSound, _ := io.NewSoundFileBuffer("resources/drums/kick/Cymatics - Humble Friday Kick - E.wav")
	snareSound, _ := io.NewSoundFileBuffer("resources/drums/snare/Cymatics - Humble Adequate Snare - E.wav")

	hihatPlayer := addDrumTrack(env, "hihat", hihatSound, bpm, 4.0, 0.875, 1.125, 1.0, value.NewAnd([]value.Valuer[*swing.Step]{
		value.NewRepeat[*swing.Step](value.NewSequenceNC([]*swing.Step{
			{}, {Shuffle: 0.3}, {Skip: true}, {Shuffle: 0.3, ShuffleRand: 0.2}, {Skip: true}, {Shuffle: 0.1}, {}, {SkipFactor: 0.4, Shuffle: 0.2}, {Skip: true}, {Skip: true},
		}), 2, 3),
		value.NewRepeat[*swing.Step](value.NewSequenceNC([]*swing.Step{
			{}, {Shuffle: 0.3}, {Skip: true}, {Skip: true}, {Skip: true}, {Shuffle: 0.3, ShuffleRand: 0.2}, {Skip: true}, {Shuffle: 0.1}, {SkipFactor: 0.4}, {SkipFactor: 0.4}, {SkipFactor: 0.4, Shuffle: 0.2}, {Skip: true}, {Skip: true},
		}), 1, 2),
	}, true))

	kickPlayer := addDrumTrack(env, "kick", kickSound, bpm, 4.0, 0.875, 1.125, 1.0, value.NewAnd([]value.Valuer[*swing.Step]{
		value.NewRepeat[*swing.Step](value.NewSequenceNC([]*swing.Step{
			{}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {SkipFactor: 0.4}, {Shuffle: 0.2}, {Skip: true}, {Skip: true},
		}), 2, 3),
		value.NewRepeat[*swing.Step](value.NewSequenceNC([]*swing.Step{
			{}, {Skip: true}, {Shuffle: 0.2}, {Skip: true}, {SkipFactor: 0.4}, {Skip: true}, {Skip: true}, {SkipFactor: 0.4}, {Shuffle: 0.2}, {Skip: true}, {Skip: true}, {Skip: true},
		}), 1, 2),
	}, true))

	snarePlayer := addDrumTrack(env, "snare", snareSound, bpm, 2.0, 0.875, 1.125, 1.0, value.NewSequence([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Shuffle: 0.1, ShuffleRand: 0.1},
	}))

	muse.Connect(kickPlayer, 0, env, 0)
	muse.Connect(hihatPlayer, 0, env, 0)
	muse.Connect(snarePlayer, 0, env, 0)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/beats.aiff", 20.0, env.Config.SampleRate, true, sndfile.SF_FORMAT_AIFF)
}

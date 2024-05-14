package main

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/bucket"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/template"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing/euclidean"
	"github.com/almerlucke/muse/modules/allpass"
	"github.com/almerlucke/muse/modules/mixer"
	"github.com/almerlucke/muse/synths/drums"
	"github.com/almerlucke/sndfile"
	"github.com/google/uuid"
)

func addDrumTrack(p muse.Patch, polyIdentifier string, sound genny.Generator[string], speed genny.Generator[float64], amp genny.Generator[float64], steps genny.Generator[*swing.Step], bpm int, noteDivision int) {
	identifier := uuid.New().String()

	p.AddMessenger(stepper.NewStepper(swing.New(bpm, noteDivision, steps), []string{identifier}))

	p.AddMessenger(banger.NewTemplateBang([]string{polyIdentifier}, template.Template{
		"command":   "trigger",
		"duration":  0.0,
		"amplitude": amp,
		"message": template.Template{
			"speed": speed,
			"sound": sound,
		},
	}).MsgrNamed(identifier))
}

func kickRhythm() genny.Generator[*swing.Step] {
	return euclidean.New(12, 5, 0, nil)
}

func bounceRhythm() genny.Generator[*swing.Step] {
	return euclidean.New(18, 5, 0, nil)
}

func hihatRhythm() genny.Generator[*swing.Step] {
	return euclidean.New(16, 9, 0, &euclidean.Config{
		Multiply: 1.0,
		Shuffle:  0.1,
	})
}

func snareRhythm() genny.Generator[*swing.Step] {
	return euclidean.New(17, 4, 1, nil)
}

func main() {
	muse.PushConfiguration(&muse.Configuration{
		SampleRate: 44100.0,
		BufferSize: 128,
	})

	root := muse.New(2)

	soundBank := sndfile.SoundBank{}

	soundBank["hihat"], _ = sndfile.NewMipMapSoundFile("resources/drums/hihat/Cymatics - Humble Closed Hihat 1.wav", 4)
	soundBank["kick1"], _ = sndfile.NewMipMapSoundFile("resources/drums/kick/Cymatics - Humble Theory Kick - B.wav", 4)
	soundBank["kick2"], _ = sndfile.NewMipMapSoundFile("resources/drums/kick/Cymatics - Humble Sit Down Kick - D.wav", 4)
	soundBank["snare"], _ = sndfile.NewMipMapSoundFile("resources/drums/snare/Cymatics - Humble King Snare 1 - A.wav", 4)
	soundBank["808_1"], _ = sndfile.NewMipMapSoundFile("resources/drums/808/Cymatics - Humble 808 4 - F.wav", 4)
	soundBank["808_2"], _ = sndfile.NewMipMapSoundFile("resources/drums/808/Cymatics - Humble 808 3 - F.wav", 4)
	soundBank["808_3"], _ = sndfile.NewMipMapSoundFile("resources/drums/fx/Cymatics - Orchid Impact FX 2.wav", 4)
	soundBank["808_4"], _ = sndfile.NewMipMapSoundFile("resources/drums/fx/Cymatics - Orchid Reverse Crash 2.wav", 4)
	soundBank["shaker"], _ = sndfile.NewMipMapSoundFile("resources/drums/shots/Cymatics - Orchid Shaker - Drew.wav", 4)

	dr := drums.NewDrums(soundBank, 20).Named("drums").AddTo(root)

	addDrumTrack(root, "drums", bucket.NewLoop(bucket.Random, "kick1", "kick2"), function.NewRandom(0.5, 1.25), constant.New(0.5), kickRhythm(), 120, 1)
	addDrumTrack(root, "drums", constant.New("hihat"), function.NewRandom(1.0, 3.0), constant.New(0.5), hihatRhythm(), 120, 4)
	addDrumTrack(root, "drums", constant.New("snare"), function.NewRandom(0.5, 3.5), constant.New(0.5), snareRhythm(), 120, 1)
	addDrumTrack(root, "drums", bucket.NewLoop(bucket.Indexed, "808_1", "808_2", "808_3", "808_4"), function.NewRandom(0.75, 1.25), constant.New(0.375), bounceRhythm(), 120, 1)

	allLeft := allpass.New(875.0, 875.0, 0.2).AddTo(root).In(dr)
	allRight := allpass.New(750.0, 750.0, 0.2).AddTo(root).In(dr, 1)

	mix1 := mixer.New(2)
	mix2 := mixer.New(2)

	mix1.AddTo(root).In(dr, allLeft)
	mix1.SetMixAt(0, 0.925)
	mix1.SetMixAt(1, 0.075)

	mix2.AddTo(root).In(dr, 1, allRight)
	mix2.SetMixAt(0, 0.925)
	mix2.SetMixAt(1, 0.075)

	root.In(mix1, mix2)

	_ = root.RenderAudio()
}

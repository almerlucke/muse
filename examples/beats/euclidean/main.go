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
	"github.com/almerlucke/muse/messengers/triggers/stepper/euclidean"
	"github.com/almerlucke/muse/synths/drums"
	"github.com/almerlucke/sndfile"
	"github.com/google/uuid"
)

func addDrumTrack(p muse.Patch, polyIdentifier string, sound genny.Generator[string], speed genny.Generator[float64], amp genny.Generator[float64], stepProvider stepper.StepProvider) {
	identifier := uuid.New().String()

	p.AddMessenger(stepper.NewStepper(stepProvider, []string{identifier}))

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

func kickRhythm() stepper.StepProvider {
	return euclidean.New(12, 5, 0, 500.0)
}

func bounceRhythm() stepper.StepProvider {
	return euclidean.New(18, 5, 0, 500.0)
}

func hihatRhythm() stepper.StepProvider {
	return euclidean.New(16, 9, 0, 125.0)
}

func snareRhythm() stepper.StepProvider {
	return euclidean.New(16, 4, 1, 500.0)
}

func main() {
	muse.PushConfiguration(&muse.Configuration{
		SampleRate: 44100.0,
		BufferSize: 256,
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

	addDrumTrack(root, "drums", bucket.NewLoop(bucket.Random, "kick1", "kick2"), function.NewRandom(0.5, 1.25), constant.New(0.5), kickRhythm())
	addDrumTrack(root, "drums", constant.New("hihat"), function.NewRandom(1.0, 2.0), constant.New(0.5), hihatRhythm())
	addDrumTrack(root, "drums", constant.New("snare"), function.NewRandom(0.5, 1.5), constant.New(0.5), snareRhythm())
	addDrumTrack(root, "drums", bucket.NewLoop(bucket.Indexed, "808_1", "808_2", "808_3", "808_4"), constant.New(1.0), constant.New(0.5), bounceRhythm())

	root.In(dr, dr, 1)

	_ = root.RenderAudio()
}

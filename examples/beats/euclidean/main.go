package main

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/bucket"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/template"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/scheduler"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing/euclidean"
	"github.com/almerlucke/muse/modules/effects/pingpong"
	"github.com/almerlucke/muse/synths/drums"
	"github.com/almerlucke/muse/utils/timing"
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

	drum := drums.NewDrums(soundBank, 20).Named("drums").AddTo(root)

	kickRhythm := euclidean.New(12, 1, 1, nil)
	bounceRhythm := euclidean.New(18, 1, 2, nil)
	hihatRhythm := euclidean.New(16, 1, 3, &euclidean.StepConfig{
		Multiply: 1.0,
		Shuffle:  0.1,
	})
	snareRhythm := euclidean.New(17, 1, 1, nil)

	bpm := 120
	bpmToMilli := timing.BPMToMilli(bpm)

	addDrumTrack(root, "drums", bucket.NewLoop(bucket.Random, "kick1", "kick2"), function.NewRandom(0.5, 1.25), constant.New(0.5), kickRhythm, bpm, 1)
	addDrumTrack(root, "drums", constant.New("hihat"), function.NewRandom(1.0, 2.0), constant.New(0.5), hihatRhythm, bpm, 4)
	addDrumTrack(root, "drums", constant.New("snare"), function.NewRandom(0.5, 3.5), constant.New(0.5), snareRhythm, bpm, 1)
	addDrumTrack(root, "drums", bucket.NewLoop(bucket.Indexed, "808_1", "808_2", "808_3", "808_4"), function.NewRandom(0.75, 1.25), constant.New(0.375), bounceRhythm, bpm, 1)

	pp := pingpong.New(bpmToMilli*0.75, bpmToMilli*0.75, 0.2, 0.95, 0.05, true).AddTo(root).In(drum, drum, 1)

	root.In(pp, pp, 1)

	sched := scheduler.New()
	sched.CtrlAddTo(root)

	sched.ScheduleFunction(timing.Second*10.0, func() {
		kickRhythm.Set(12, 5, 0)
		bounceRhythm.Set(18, 5, 0)
		hihatRhythm.Set(16, 9, 0)
		snareRhythm.Set(17, 4, 1)
	})
	//sched.ScheduleFunction(timing.Minute*1.5, func() {
	//	kickRhythm.Set(12, 7, 0)
	//	bounceRhythm.Set(18, 6, 0)
	//	hihatRhythm.Set(16, 11, 0)
	//	snareRhythm.Set(17, 7, 1)
	//})

	_ = root.RenderAudio()
}

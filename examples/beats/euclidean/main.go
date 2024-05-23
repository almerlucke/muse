package main

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/bucket"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/template"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/messengers/scheduler"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing/euclidean"
	"github.com/almerlucke/muse/modules/effects/flanger"
	"github.com/almerlucke/muse/modules/effects/pingpong"
	"github.com/almerlucke/muse/synths/drums"
	"github.com/almerlucke/muse/utils/timing"
	"github.com/almerlucke/sndfile"
	"github.com/almerlucke/sndfile/writer"
	"github.com/google/uuid"
	"log"
	"math/rand"
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

	soundBank["hihat1"], _ = sndfile.NewMipMapSoundFile("resources/drums/hihat/Cymatics - Orchid Hihat - Flam.wav", 4)
	soundBank["hihat2"], _ = sndfile.NewMipMapSoundFile("resources/drums/hihat/Cymatics - Orchid Ride - Mysterious.wav", 4)
	soundBank["kick1"], _ = sndfile.NewMipMapSoundFile("resources/drums/kick/Cymatics - Orchid Kick - Dancehall (A#).wav", 4)
	soundBank["kick2"], _ = sndfile.NewMipMapSoundFile("resources/drums/kick/Cymatics - Orchid Kick - Tight (G).wav", 4)
	soundBank["snare1"], _ = sndfile.NewMipMapSoundFile("resources/drums/snare/Cymatics - Orchid Snap - Single.wav", 4)
	soundBank["snare2"], _ = sndfile.NewMipMapSoundFile("resources/drums/snare/Cymatics - Orchid Snap - Cream.wav", 4)
	soundBank["fx1"], _ = sndfile.NewMipMapSoundFile("resources/drums/percussion/Cymatics - Orchid Percussion - Wet 1 (C).wav", 4)
	soundBank["fx2"], _ = sndfile.NewMipMapSoundFile("resources/drums/percussion/Cymatics - Orchid Percussion - Wet 2.wav", 4)
	soundBank["fx3"], _ = sndfile.NewMipMapSoundFile("resources/drums/percussion/Cymatics - Orchid Percussion - Wet 3.wav", 4)
	soundBank["fx4"], _ = sndfile.NewMipMapSoundFile("resources/drums/percussion/Cymatics - Orchid Percussion - Wet 4.wav", 4)

	drum := drums.NewDrums(soundBank, 20).Named("drums").AddTo(root)

	kickRhythm := euclidean.New(12, 1, 1, nil)
	bounceRhythm := euclidean.New(18, 1, 2, nil)
	hihatRhythm := euclidean.New(16, 1, 3, &euclidean.StepConfig{
		Multiply: 1.0,
		Shuffle:  0.1,
	})
	snareRhythm := euclidean.New(17, 1, 1, nil)

	bpm := 82
	bpmToMilli := timing.BPMToMilli(bpm)

	hihatConst := constant.New("hihat1")
	snareConst := constant.New("snare1")
	snareLow := 0.5
	snareHigh := 3.5
	snareRand := function.New(func() float64 { return rand.Float64()*(snareHigh-snareLow) + snareLow })

	addDrumTrack(root, "drums", bucket.NewLoop(bucket.Random, "kick1", "kick2"), function.NewRandom(0.5, 1.25), constant.New(0.5), kickRhythm, bpm, 1)
	addDrumTrack(root, "drums", hihatConst, function.NewRandom(1.0, 2.0), constant.New(0.1), hihatRhythm, bpm, 4)
	addDrumTrack(root, "drums", snareConst, snareRand, constant.New(1.0), snareRhythm, bpm, 1)
	addDrumTrack(root, "drums", bucket.NewLoop(bucket.Indexed, "fx1", "fx2", "fx3", "fx4"), function.NewRandom(0.75, 1.25), constant.New(0.5), bounceRhythm, bpm, 1)

	flang := flanger.New(0.3, 0.5, 0.7, 0.3, true).AddTo(root).In(drum, drum, 1)
	flangLfo := lfo.NewBasicControlLFO(0.05, 0.2, 0.4).CtrlAddTo(root)
	flang.CtrlIn(flangLfo)
	pp := pingpong.New(bpmToMilli*2, bpmToMilli*0.75, 0.2, 0.95, 0.05, true).AddTo(root).In(flang, flang, 1)

	root.In(pp, pp, 1)

	sched := scheduler.New()
	sched.CtrlAddTo(root)

	sched.ScheduleFunction(timing.Second*15.0, func() {
		kickRhythm.Set(12, 5, 0)
		bounceRhythm.Set(18, 5, 0)
		hihatRhythm.Set(16, 9, 0)
		snareRhythm.Set(17, 4, 1)
	})
	sched.ScheduleFunction(timing.Second*30.0, func() {
		log.Printf("pingpong1")
		hihatConst.SetValue("hihat2")
		hihatRhythm.Set(16, 5, 0)
		pp.(*pingpong.PingPong).SetRead(bpmToMilli * 1.25)
	})
	sched.ScheduleFunction(timing.Second*60.0, func() {
		log.Printf("pingpong2")
		hihatConst.SetValue("hihat1")
		hihatRhythm.Set(16, 9, 0)
		pp.(*pingpong.PingPong).SetRead(bpmToMilli * 0.75)
	})
	sched.ScheduleFunction(timing.Second*90.0, func() {
		log.Printf("pingpong3")
		hihatConst.SetValue("hihat2")
		hihatRhythm.Set(16, 5, 0)
		snareConst.SetValue("snare2")
		snareLow = 0.9
		snareHigh = 1.1
		pp.(*pingpong.PingPong).SetRead(bpmToMilli * 1.75)
	})
	sched.ScheduleFunction(timing.Second*120.0, func() {
		log.Printf("pingpong4")
		hihatConst.SetValue("hihat1")
		hihatRhythm.Set(16, 9, 0)
		snareConst.SetValue("snare1")
		snareLow = 0.5
		snareHigh = 3.5
		pp.(*pingpong.PingPong).SetRead(bpmToMilli * 0.75)
	})
	sched.ScheduleFunction(timing.Second*150.0, func() {
		log.Printf("pingpong5")
		hihatConst.SetValue("hihat2")
		hihatRhythm.Set(16, 3, 0)
		snareConst.SetValue("snare2")
		snareLow = 0.9
		snareHigh = 1.1
		pp.(*pingpong.PingPong).SetRead(bpmToMilli * 1.25)
	})
	sched.ScheduleFunction(timing.Second*180.0, func() {
		log.Printf("pingpong6")
		hihatConst.SetValue("hihat1")
		hihatRhythm.Set(16, 9, 0)
		pp.(*pingpong.PingPong).SetRead(bpmToMilli * 0.75)
	})
	//sched.ScheduleFunction(timing.Minute*1.5, func() {
	//	kickRhythm.Set(12, 7, 0)
	//	bounceRhythm.Set(18, 6, 0)
	//	hihatRhythm.Set(16, 11, 0)
	//	snareRhythm.Set(17, 7, 1)
	//})

	// _ = root.RenderAudio()
	_ = root.RenderToSoundFile("/home/almer/Music/BirdsDrumsSlow", writer.AIFC, 240, 44100.0, true)
}

package main

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/float/ramp"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/template"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing/euclidean"
	"github.com/almerlucke/muse/modules/effects/freeverb"
	"github.com/almerlucke/muse/modules/effects/pingpong"
	"github.com/almerlucke/muse/synths/drums"
	"github.com/almerlucke/muse/utils/timing"
	"github.com/almerlucke/sndfile"
)

func addDrumTrack(p muse.Patch, drums muse.Module, sounds genny.Generator[string], tempo int, division int, speed genny.Generator[float64], amp genny.Generator[float64], steps genny.Generator[*swing.Step]) {
	drums.CtrlIn(banger.NewControlTemplate(template.Template{
		"command":   "trigger",
		"duration":  0.0,
		"amplitude": amp,
		"message": template.Template{
			"speed": speed,
			"sound": sounds,
		},
	}).CtrlIn(stepper.NewStepper(swing.New(tempo, division, steps), nil).CtrlAddTo(p)))
}

func main() {
	muse.PushConfiguration(&muse.Configuration{
		SampleRate: 44100.0,
		BufferSize: 1024,
	})

	bpm := 82
	bpmToMilli := timing.BPMToMilli(bpm)

	root := muse.New(2)

	soundBank := sndfile.SoundBank{}

	soundBank["shake1"], _ = sndfile.NewMipMapSoundFile("resources/samples/homemade/ShakesLivingRoom7.wav", 4)
	soundBank["shake2"], _ = sndfile.NewMipMapSoundFile("resources/samples/homemade/ShakesLivingRoom6.wav", 4)
	soundBank["shake3"], _ = sndfile.NewMipMapSoundFile("resources/samples/homemade/ShakesLivingRoom1.wav", 4)
	soundBank["shake4"], _ = sndfile.NewMipMapSoundFile("resources/samples/homemade/RollLivingRoom6.wav", 4)
	soundBank["shake5"], _ = sndfile.NewMipMapSoundFile("resources/samples/homemade/RollLabyrinth4.wav", 4)
	soundBank["shake6"], _ = sndfile.NewMipMapSoundFile("resources/samples/homemade/ShakesLivingRoom4.wav", 4)

	drum := drums.NewDrums(soundBank, 20).AddTo(root)

	addDrumTrack(root, drum, constant.New("shake1"), bpm, 1,
		function.NewRandom(0.35, 0.85), ramp.NewForever(20, 0.0, 0.5, 1.0), euclidean.New(32, 6, 1, nil))
	addDrumTrack(root, drum, constant.New("shake2"), bpm, 1,
		function.NewRandom(0.35, 1.25), ramp.NewForever(20, 0.0, 0.125, 1.0), euclidean.New(32, 4, 2, nil))
	addDrumTrack(root, drum, constant.New("shake3"), bpm, 2,
		function.NewRandom(0.85, 1.25), ramp.NewForever(20, 0.0, 0.25, 1.0), euclidean.New(32, 8, 3, nil))
	addDrumTrack(root, drum, constant.New("shake4"), bpm, 1,
		function.NewRandom(0.85, 3.25), ramp.NewForever(20, 0.0, 0.25, 1.0), euclidean.New(32, 7, 4, nil))
	addDrumTrack(root, drum, constant.New("shake5"), bpm, 1,
		function.NewRandom(0.35, 1.25), ramp.NewForever(20, 0.0, 0.35, 1.0), euclidean.New(32, 5, 5, nil))
	addDrumTrack(root, drum, constant.New("shake6"), bpm, 2,
		function.NewRandom(0.25, 2.25), ramp.NewForever(20, 0.0, 0.15, 1.0), euclidean.New(32, 7, 6, nil))

	pp := pingpong.New(bpmToMilli*2, bpmToMilli*0.75, 0.2, 0.05).AddTo(root).In(drum, drum, 1)
	fv := freeverb.New().AddTo(root).In(pp, pp, 1).Exec(func(obj any) {
		fv := obj.(*freeverb.FreeVerb)
		fv.SetDamp(0.7)
		fv.SetRoomSize(0.9)
		fv.SetWet(0.02)
		fv.SetDry(0.3)
	})

	root.In(fv, fv, 1)

	_ = root.RenderAudio()
}

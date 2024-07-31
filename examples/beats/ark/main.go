package main

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/float/ramp"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/template"
	"github.com/almerlucke/muse"
	rbjc "github.com/almerlucke/muse/components/filters/rbj"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing/euclidean"
	"github.com/almerlucke/muse/modules/effects/freeverb"
	"github.com/almerlucke/muse/modules/effects/pingpong"
	"github.com/almerlucke/muse/modules/filters/rbj"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/noise"
	"github.com/almerlucke/muse/modules/pan"
	"github.com/almerlucke/muse/modules/player"
	"github.com/almerlucke/muse/synths/drums"
	"github.com/almerlucke/muse/utils/timing"
	"github.com/almerlucke/sndfile"
	"github.com/almerlucke/sndfile/writer"
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

	drum := drums.NewDrums(sndfile.SoundBank{
		"shake1": sndfile.MustMipMapSoundFile("resources/samples/homemade/ShakesLivingRoom7.wav", 4),
		"shake2": sndfile.MustMipMapSoundFile("resources/samples/homemade/ShakesLivingRoom6.wav", 4),
		"shake3": sndfile.MustMipMapSoundFile("resources/samples/homemade/ShakesLivingRoom1.wav", 4),
		"shake4": sndfile.MustMipMapSoundFile("resources/samples/homemade/RollLivingRoom6.wav", 4),
		"shake5": sndfile.MustMipMapSoundFile("resources/samples/homemade/RollLabyrinth4.wav", 4),
		"shake6": sndfile.MustMipMapSoundFile("resources/samples/homemade/ShakesLivingRoom4.wav", 4),
	}, 20).AddTo(root)

	addDrumTrack(root, drum, constant.New("shake1"), bpm, 1,
		function.NewRandom(0.35, 0.85), ramp.NewForever(40, 0.0, 0.5, 1.0), euclidean.New(32, 6, 1, nil))
	addDrumTrack(root, drum, constant.New("shake2"), bpm, 1,
		function.NewRandom(0.35, 1.25), ramp.NewForever(40, 0.0, 0.125, 1.0), euclidean.New(32, 4, 2, nil))
	addDrumTrack(root, drum, constant.New("shake3"), bpm, 2,
		function.NewRandom(0.85, 1.25), ramp.NewForever(60, 0.0, 0.25, 1.0), euclidean.New(32, 8, 3, nil))
	addDrumTrack(root, drum, constant.New("shake4"), bpm, 1,
		function.NewRandom(0.85, 3.25), ramp.NewForever(50, 0.0, 0.25, 1.0), euclidean.New(32, 7, 4, nil))
	addDrumTrack(root, drum, constant.New("shake5"), bpm, 1,
		function.NewRandom(0.35, 1.25), ramp.NewForever(40, 0.0, 0.35, 1.0), euclidean.New(32, 5, 5, nil))
	addDrumTrack(root, drum, constant.New("shake6"), bpm, 2,
		function.NewRandom(0.25, 2.25), ramp.NewForever(50, 0.0, 0.15, 1.0), euclidean.New(32, 7, 6, nil))

	pp := pingpong.New(bpmToMilli*2, bpmToMilli*0.75, 0.2, 0.05).AddTo(root).In(drum, drum, 1)

	const na = 0.075
	n := noise.New(10).AddTo(root)
	f1 := rbj.New(rbjc.Lowpass, 1000.0, 4.0).AddTo(root).In(n).CtrlIn(lfo.NewBasicControlLFO(0.05, 100.0, 1000.0).CtrlAddTo(root))
	m1 := functor.NewAmp(na).AddTo(root).In(f1)
	p1 := pan.New(0.5).AddTo(root).In(m1).CtrlIn(lfo.NewBasicControlLFO(0.03123, 0.25, 0.75).CtrlAddTo(root))
	f2 := rbj.New(rbjc.Lowpass, 300.0, 5.0).AddTo(root).In(n).CtrlIn(lfo.NewBasicControlLFO(0.04567, 100.0, 1300.0).CtrlAddTo(root))
	m2 := functor.NewAmp(na).AddTo(root).In(f2)
	p2 := pan.New(0.35).AddTo(root).In(m2).CtrlIn(lfo.NewBasicControlLFO(0.02123, 0.25, 0.75).CtrlAddTo(root))

	pl1 := player.New(sndfile.MustSoundFile("resources/sounds/rain-and-distant-thunder-60230.wav"), 1.0, 0.4, false).AddTo(root)
	pl2 := player.New(sndfile.MustSoundFile("resources/sounds/boat_waves-6099.wav"), 1.0, 0.2, false).AddTo(root)
	pl3 := player.New(sndfile.MustSoundFile("resources/sounds/seagulls-by-the-sea-7042.wav"), 1.0, 0.2, false).AddTo(root)
	pl4 := player.New(sndfile.MustSoundFile("resources/sounds/animals.wav"), 1.0, 0.125, false).AddTo(root)

	fv := freeverb.New().AddTo(root).In(pp, pp, 1, p1, 0, 0, p1, 1, 1, p2, 0, 0, p2, 1, 1, pl1, 0, 0, pl1, 1, 1, pl2, 0, 0, pl2, 1, 1, pl3, 0, 0, pl3, 1, 1, pl4, 0, 0, pl4, 1, 1).Exec(func(obj any) {
		fv := obj.(*freeverb.FreeVerb)
		fv.SetDamp(0.7)
		fv.SetRoomSize(0.9)
		fv.SetWet(0.02)
		fv.SetDry(0.3)
	})

	root.In(fv, fv, 1)

	_ = root.RenderToSoundFile("/home/almer/Music/MuseRenders/NoahsArk.wav", writer.WAV, 7*60, 44100.0, true)

	//_ = root.RenderAudio()
}

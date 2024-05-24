package main

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/bucket"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/flatten"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/markov"
	"github.com/almerlucke/genny/template"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/controls/perform"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/messengers/triggers/timer"
	"github.com/almerlucke/muse/modules/effects/flanger"
	"github.com/almerlucke/muse/modules/effects/pingpong"
	"github.com/almerlucke/muse/synths/drums"
	"github.com/almerlucke/muse/utils/timing"
	"github.com/almerlucke/sndfile"
	"github.com/google/uuid"
)

func addDrumTrack(p muse.Patch, polyName string, sounds genny.Generator[string], tempo int, division int, speed genny.Generator[float64], amp genny.Generator[float64], steps genny.Generator[*swing.Step]) {
	identifier := uuid.New().String()

	p.AddMessenger(stepper.NewStepper(swing.New(tempo, division, steps), []string{identifier}))

	p.AddMessenger(banger.NewTemplateBang([]string{polyName}, template.Template{
		"command":   "trigger",
		"duration":  0.0,
		"amplitude": amp,
		"message": template.Template{
			"speed": speed,
			"sound": sounds,
		},
	}).MsgrNamed(identifier))
}

func kickRhythm() genny.Generator[*swing.Step] {
	rhythm1 := markov.NewProbabilityState(swing.QuickSteps(1, 0, 0, 1, 0, 0, 0, 0 /* 8 */, 0, 0, 0, 0, 0, 0, 0, 0 /* 16 */))
	rhythm2 := markov.NewProbabilityState(swing.QuickSteps(1, 0, 0, 1, 0, 0, 0, 0 /* 8 */, 0, 0, 1, 0, 0, 0, 0, 0 /* 16 */))
	rhythm3 := markov.NewProbabilityState(swing.QuickSteps(1, 0, 0, 1, 0, 0, 0, 0 /* 8 */, 0, 0, 1, 0, 0, 0, 1, 0 /* 16 */))
	rhythm4 := markov.NewProbabilityState(swing.QuickSteps(1, 0, 0, 1, 0, 0, 1, 0 /* 8 */, 0, 0, 1, 0, 0, 0, 1, 0 /* 16 */))
	rhythm5 := markov.NewProbabilityState(swing.QuickSteps(1, 0, 0, 1, 0, 0, 1, 0 /* 8 */, 0, 0, 1, 0, 0, 0, 1, 0 /* 16 */))

	rhythm1.SetProbabilities(rhythm1, 3.0, rhythm2, 2.0, rhythm3, 1.0)
	rhythm2.SetProbabilities(rhythm2, 3.0, rhythm3, 2.0, rhythm4, 1.0)
	rhythm3.SetProbabilities(rhythm3, 3.0, rhythm4, 2.0, rhythm5, 1.0)
	rhythm4.SetProbabilities(rhythm4, 4.0, rhythm5, 2.0, rhythm1, 1.0)
	rhythm5.SetProbabilities(rhythm5, 4.0, rhythm1, 2.0, rhythm2, 1.0)

	return flatten.New[*swing.Step](markov.New[[]*swing.Step](rhythm1, 1))
}

func snareRhythm() genny.Generator[*swing.Step] {
	b := &swing.Step{BurstChance: 0.5, NumBurst: 3}

	rhythm1 := markov.NewProbabilityState(swing.QuickSteps(0, 0, 0, 0, 1, 0, 0, 0 /* 8 */, 0, 0, 0, 0, 0, 0, 0, 0 /* 16 */))
	rhythm2 := markov.NewProbabilityState(swing.QuickSteps(0, 0, 0, 0, 1, 0, 0, 0 /* 8 */, 0, 0, 0, 0, 1, 0, 0, 0 /* 16 */))
	rhythm3 := markov.NewProbabilityState(swing.QuickSteps(0, 0, 0, 0, 1, 0, 0, 0 /* 8 */, 0, 0, 0, 0, b, 0, 0, 0 /* 16 */))
	rhythm4 := markov.NewProbabilityState(swing.QuickSteps(0, 0, 0, 0, 1, 0, 0, 1 /* 8 */, 0, 0, 0, 0, b, 0, 0, 0 /* 16 */))
	rhythm5 := markov.NewProbabilityState(swing.QuickSteps(0, 0, 0, 0, 1, 0, 0, b /* 8 */, 0, 1, 0, 0, 0, 0, 0, 1 /* 16 */))

	rhythm1.SetProbabilities(rhythm1, 3.0, rhythm2, 2.0, rhythm3, 1.0)
	rhythm2.SetProbabilities(rhythm2, 3.0, rhythm3, 2.0, rhythm4, 1.0)
	rhythm3.SetProbabilities(rhythm3, 3.0, rhythm4, 2.0, rhythm5, 1.0)
	rhythm4.SetProbabilities(rhythm4, 4.0, rhythm5, 2.0, rhythm1, 1.0)
	rhythm5.SetProbabilities(rhythm5, 4.0, rhythm1, 2.0, rhythm2, 1.0)

	return flatten.New[*swing.Step](markov.New[[]*swing.Step](rhythm1, 1))
}

func bassRhythm() genny.Generator[*swing.Step] {
	rhythm1 := markov.NewProbabilityState(swing.QuickSteps(0, 0, 0, 0, 1, 0, 0, 0 /* 8 */, 0, 0, 0, 0, 0, 0, 0, 0 /* 16 */))
	rhythm2 := markov.NewProbabilityState(swing.QuickSteps(0, 0, 0, 0, 1, 0, 0, 0 /* 8 */, 0, 0, 0, 0, 1, 0, 0, 0 /* 16 */))
	rhythm3 := markov.NewProbabilityState(swing.QuickSteps(0, 0, 0, 0, 1, 0, 0, 1 /* 8 */, 0, 0, 0, 0, 1, 0, 0, 0 /* 16 */))
	rhythm4 := markov.NewProbabilityState(swing.QuickSteps(0, 0, 0, 0, 1, 0, 0, 1 /* 8 */, 0, 1, 0, 0, 0, 0, 0, 1 /* 16 */))

	rhythm1.SetProbabilities(rhythm1, 3.0, rhythm2, 2.0, rhythm3, 1.0)
	rhythm2.SetProbabilities(rhythm2, 3.0, rhythm3, 2.0, rhythm4, 1.0)
	rhythm3.SetProbabilities(rhythm3, 3.0, rhythm4, 2.0, rhythm1, 1.0)
	rhythm4.SetProbabilities(rhythm4, 4.0, rhythm1, 2.0, rhythm2, 1.0)

	return flatten.New[*swing.Step](markov.New[[]*swing.Step](rhythm1, 1))
}

func hihatRhythm() genny.Generator[*swing.Step] {
	rhythm1 := markov.NewProbabilityState(swing.QuickSteps(1, 0, 1, 0, 1, 0, 1, 0))
	rhythm2 := markov.NewProbabilityState(swing.QuickSteps(&swing.Step{BurstChance: 0.5, NumBurst: 3}, &swing.Step{Shuffle: 0.1}, 1, 0, 1, &swing.Step{Shuffle: 0.1}, 1, 0))
	rhythm3 := markov.NewProbabilityState([]*swing.Step{
		{BurstChance: 0.5, NumBurst: 3}, {Shuffle: 0.1}, {Skip: true}, {Shuffle: 0.1}, {}, {Shuffle: 0.1}, {Skip: true}, {Shuffle: 0.1},
	})

	rhythm4 := markov.NewProbabilityState([]*swing.Step{
		{BurstChance: 0.5, NumBurst: 3}, {Skip: true}, {}, {Shuffle: 0.1}, {}, {Skip: true}, {}, {Shuffle: 0.1},
	})

	rhythm5 := markov.NewProbabilityState([]*swing.Step{
		{}, {Shuffle: 0.1}, {}, {Shuffle: 0.1}, {}, {Shuffle: 0.1}, {BurstChance: 0.5, NumBurst: 3}, {Shuffle: 0.1},
	})

	rhythm1.SetProbabilities(rhythm1, 2.0, rhythm2, 1.0, rhythm3, 1.0)
	rhythm2.SetProbabilities(rhythm2, 2.0, rhythm3, 1.0, rhythm4, 1.0)
	rhythm3.SetProbabilities(rhythm3, 2.0, rhythm4, 1.0, rhythm5, 1.0)
	rhythm4.SetProbabilities(rhythm4, 3.0, rhythm5, 1.0, rhythm1, 1.0)
	rhythm5.SetProbabilities(rhythm5, 3.0, rhythm1, 1.0, rhythm2, 1.0)

	return flatten.New[*swing.Step](markov.New[[]*swing.Step](rhythm1, 1))
}

func main() {
	muse.PushConfiguration(&muse.Configuration{
		SampleRate: 44100.0,
		BufferSize: 256,
	})

	root := muse.New(2)

	bpm := 55
	bpmToMs := timing.BPMToMilli(bpm)

	soundBank := sndfile.SoundBank{}

	soundBank["hihat1"], _ = sndfile.NewMipMapSoundFile("resources/drums/hihat/Cymatics - Humble Closed Hihat 1.wav", 4)
	soundBank["hihat2"], _ = sndfile.NewMipMapSoundFile("resources/drums/hihat/Cymatics - Odyssey Closed Hihat 1.wav", 4)
	soundBank["hihat3"], _ = sndfile.NewMipMapSoundFile("resources/drums/hihat/Cymatics - Odyssey Closed Hihat 2.wav", 4)
	soundBank["hihat4"], _ = sndfile.NewMipMapSoundFile("resources/drums/hihat/Cymatics - Orchid Hihat - Closed 2.wav", 4)
	soundBank["hihat5"], _ = sndfile.NewMipMapSoundFile("resources/drums/hihat/Cymatics - Orchid Hihat - Closed 2.wav", 4)

	soundBank["kick1"], _ = sndfile.NewMipMapSoundFile("resources/drums/kick/Cymatics - Orchid Kick - Clean (F).wav", 4)
	soundBank["kick2"], _ = sndfile.NewMipMapSoundFile("resources/drums/kick/Cymatics - Orchid Kick - Dancehall (A#).wav", 4)
	soundBank["kick3"], _ = sndfile.NewMipMapSoundFile("resources/drums/kick/Cymatics - Orchid Kick - Layered (F#).wav", 4)
	soundBank["kick4"], _ = sndfile.NewMipMapSoundFile("resources/drums/kick/Cymatics - Orchid Kick - Tight (G).wav", 4)

	soundBank["snare1"], _ = sndfile.NewMipMapSoundFile("resources/drums/snare/Cymatics - Odyssey House Snare 4 - F#.wav", 4)
	soundBank["snare2"], _ = sndfile.NewMipMapSoundFile("resources/drums/clap/Cymatics - Odyssey Flam Clap 1.wav", 4)
	soundBank["snare3"], _ = sndfile.NewMipMapSoundFile("resources/drums/snare/Cymatics - Odyssey House Snare 3 - E.wav", 4)
	soundBank["snare4"], _ = sndfile.NewMipMapSoundFile("resources/drums/snare/Cymatics - Odyssey House Snare 1 - C#.wav", 4)
	soundBank["snare5"], _ = sndfile.NewMipMapSoundFile("resources/drums/snare/Cymatics - Odyssey House Snare 2 - D.wav", 4)

	soundBank["fx1"], _ = sndfile.NewMipMapSoundFile("resources/drums/fx/Cymatics - Orchid Transition FX 2 - F Min.wav", 4)
	soundBank["fx2"], _ = sndfile.NewMipMapSoundFile("resources/drums/shots/Cymatics - Odyssey Synth One Shot 9 - E.wav", 4)
	soundBank["fx3"], _ = sndfile.NewMipMapSoundFile("resources/drums/fx/Cymatics - Orchid Reverse Impact 2.wav", 4)
	soundBank["fx4"], _ = sndfile.NewMipMapSoundFile("resources/drums/shots/Cymatics - Odyssey Synth One Shot 24 - A#.wav", 4)
	soundBank["fx5"], _ = sndfile.NewMipMapSoundFile("resources/drums/808/Cymatics - Humble 808 7 - A.wav", 4)
	soundBank["fx6"], _ = sndfile.NewMipMapSoundFile("resources/drums/shots/Cymatics - Odyssey Synth One Shot 17 - F#.wav", 4)
	soundBank["fx7"], _ = sndfile.NewMipMapSoundFile("resources/drums/808/Cymatics - Humble 808 4 - F.wav", 4)

	soundBank["sh1"], _ = sndfile.NewMipMapSoundFile("resources/drums/hihat/Cymatics - Orchid Shaker - Drew.wav", 4)
	soundBank["sh2"], _ = sndfile.NewMipMapSoundFile("resources/drums/percussion/Cymatics - Orchid Percussion - Dry 6.wav", 4)
	soundBank["sh3"], _ = sndfile.NewMipMapSoundFile("resources/drums/percussion/Cymatics - Orchid Percussion - Wet 2.wav", 4)
	soundBank["sh4"], _ = sndfile.NewMipMapSoundFile("resources/drums/percussion/Cymatics - Orchid Percussion - Wet 3.wav", 4)
	soundBank["sh5"], _ = sndfile.NewMipMapSoundFile("resources/drums/percussion/Cymatics - Orchid Percussion - Wet 4.wav", 4)

	dr := drums.NewDrums(soundBank, 20).Named("drums").AddTo(root)

	addDrumTrack(root, "drums", bucket.NewLoop(bucket.Indexed, "hihat1", "hihat2", "hihat3", "hihat4", "hihat5"), bpm, 8, function.NewRandom(0.5, 4.525), constant.New(0.6), hihatRhythm())
	addDrumTrack(root, "drums", bucket.NewLoop(bucket.Indexed, "kick1", "kick2", "kick3", "kick4"), bpm, 8, function.NewRandom(0.75, 2.525), constant.New(0.7), kickRhythm())
	addDrumTrack(root, "drums", bucket.NewLoop(bucket.Indexed, "snare1", "snare2", "snare3", "snare4", "snare5"), bpm, 8, function.NewRandom(0.7, 3.25), constant.New(0.7), snareRhythm())
	addDrumTrack(root, "drums", bucket.NewLoop(bucket.Indexed, "fx1", "fx2", "fx3", "fx4", "fx5", "fx6", "fx7"), bpm, 2, function.NewRandom(0.3, 7.0), constant.New(0.7), bassRhythm())
	addDrumTrack(root, "drums", bucket.NewLoop(bucket.Indexed, "sh1", "sh2", "sh3", "sh4", "sh5"), bpm, 2, function.NewRandom(0.3, 3.0), constant.New(0.5), kickRhythm())

	fl := flanger.New(0.3, 0.5, 0.2, true).AddTo(root).In(dr, dr, 1)
	flDepth := lfo.NewBasicControlLFO(0.05, 0.1, 0.9).CtrlAddTo(root)
	flFb := lfo.NewBasicControlLFO(0.0375, 0.1, 0.8).CtrlAddTo(root)
	flMix := lfo.NewBasicControlLFO(0.0425, 0.05, 0.4).CtrlAddTo(root)
	fl.CtrlIn(flDepth, flFb, 0, 1, flMix, 0, 2)

	pp := pingpong.New(bpmToMs*2.0, bpmToMs*0.375, 0.1, 0.1, true).AddTo(root).In(fl, fl, 1)
	ppReadGen := bucket.NewLoop(bucket.Indexed, 0.375, 1.5, 0.75, 1.875)
	ppMixGen := bucket.NewLoop(bucket.Indexed, 0.1, 0.05, 0.075, 0.125, 0.025)

	perform.New(func() {
		readPos := ppReadGen.Generate()
		pp.(*pingpong.PingPong).SetRead(bpmToMs * readPos)
		pp.(*pingpong.PingPong).SetMix(ppMixGen.Generate())
	}).CtrlIn(timer.NewControl(bpmToMs * 16).CtrlAddTo(root))

	root.In(pp, pp, 1)

	//err := m.RenderToSoundFile("/home/almer/Documents/drums", writer.AIFC, 240, 44100.0, true)
	//if err != nil {
	//	log.Printf("error rendering drums! %v", err)
	//}

	_ = root.RenderAudio()
}

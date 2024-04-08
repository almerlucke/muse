package main

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/flatten"
	"github.com/almerlucke/genny/markov"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/sndfile/writer"
	"log"

	"github.com/almerlucke/sndfile"
	"math/rand"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/synths/drums"
	"github.com/almerlucke/muse/utils"

	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/template"
)

func addDrumTrack(p muse.Patch, polyName string, sounds []string, tempo int, division int, lowSpeed float64, highSpeed float64, amp float64, steps genny.Generator[*swing.Step]) {
	identifier := sounds[0] + "Drum"

	p.AddMessenger(stepper.NewStepper(swing.New(tempo, division, steps), []string{identifier}))

	p.AddMessenger(banger.NewTemplateGenerator([]string{polyName}, template.Template{
		"command":   "trigger",
		"duration":  0.0,
		"amplitude": amp,
		"message": template.Template{
			"speed": function.New(nil, func(ctx any) any { return rand.Float64()*(highSpeed-lowSpeed) + lowSpeed }),
			"sound": sequence.New(utils.ToAnySlice(sounds)...),
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
	rhythm1 := markov.NewProbabilityState(swing.QuickSteps(0, 0, 0, 0, 1, 0, 0, 0 /* 8 */, 0, 0, 0, 0, 0, 0, 0, 0 /* 16 */))
	rhythm2 := markov.NewProbabilityState(swing.QuickSteps(0, 0, 0, 0, 1, 0, 0, 0 /* 8 */, 0, 0, 0, 0, 1, 0, 0, 0 /* 16 */))
	rhythm3 := markov.NewProbabilityState(swing.QuickSteps(0, 0, 0, 0, 1, 0, 0, 0 /* 8 */, 0, 0, 0, 0, &swing.Step{BurstChance: 0.5, NumBurst: 3}, 0, 0, 0 /* 16 */))
	rhythm4 := markov.NewProbabilityState(swing.QuickSteps(0, 0, 0, 0, 1, 0, 0, 1 /* 8 */, 0, 0, 0, 0, &swing.Step{BurstChance: 0.5, NumBurst: 3}, 0, 0, 0 /* 16 */))
	rhythm5 := markov.NewProbabilityState(swing.QuickSteps(0, 0, 0, 0, 1, 0, 0, &swing.Step{BurstChance: 0.5, NumBurst: 3} /* 8 */, 0, 1, 0, 0, 0, 0, 0, 1 /* 16 */))

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

	m := muse.New(2)

	bpm := 80

	soundBank := sndfile.SoundBank{}

	soundBank["hihat"], _ = sndfile.NewMipMapSoundFile("resources/drums/hihat/Cymatics - Humble Closed Hihat 1.wav", 4)
	soundBank["kick"], _ = sndfile.NewMipMapSoundFile("resources/drums/kick/Cymatics - Humble Triple Kick - E.wav", 4)
	soundBank["snare"], _ = sndfile.NewMipMapSoundFile("resources/drums/snare/Cymatics - Humble Adequate Snare - E.wav", 4)
	soundBank["808_1"], _ = sndfile.NewMipMapSoundFile("resources/drums/808/Cymatics - Humble 808 4 - F.wav", 4)
	soundBank["808_2"], _ = sndfile.NewMipMapSoundFile("resources/drums/808/Cymatics - Humble 808 3 - F.wav", 4)
	soundBank["808_3"], _ = sndfile.NewMipMapSoundFile("resources/drums/fx/Cymatics - Orchid Impact FX 2.wav", 4)
	soundBank["808_4"], _ = sndfile.NewMipMapSoundFile("resources/drums/fx/Cymatics - Orchid Reverse Crash 2.wav", 4)
	soundBank["shaker"], _ = sndfile.NewMipMapSoundFile("resources/drums/shots/Cymatics - Orchid Shaker - Drew.wav", 4)

	dr := drums.NewDrums(soundBank, 20).Named("drums").AddTo(m)

	addDrumTrack(m, "drums", []string{"hihat"}, bpm, 8, 0.575, 3.525, 0.6, hihatRhythm())
	addDrumTrack(m, "drums", []string{"kick"}, bpm, 8, 0.575, 3.525, 1.0, kickRhythm())
	addDrumTrack(m, "drums", []string{"snare"}, bpm, 8, 0.6, 3.4, 0.7, snareRhythm())
	addDrumTrack(m, "drums", []string{"808_1", "808_2", "808_3", "808_4"}, bpm, 2, 0.6, 4.0, 0.3, bassRhythm())
	addDrumTrack(m, "drums", []string{"shaker"}, bpm, 2, 0.6, 4.0, 1.0, kickRhythm())

	m.In(dr, dr, 1)

	err := m.RenderToSoundFile("/home/almer/Documents/drums", writer.AIFC, 240, 44100.0, true)
	if err != nil {
		log.Printf("error rendering drums! %v", err)
	}

	//_ = m.RenderAudio()
}

package main

import (
	"github.com/almerlucke/sndfile"
	"github.com/almerlucke/sndfile/writer"
	"math/rand"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/envelopes/adsr"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/scheduler"
	"github.com/almerlucke/muse/messengers/triggers/once"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/modules/allpass"
	"github.com/almerlucke/muse/modules/chorus"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/mixer"
	"github.com/almerlucke/muse/modules/player"
	"github.com/almerlucke/muse/synths/classic"
	"github.com/almerlucke/muse/utils/notes"
	"github.com/almerlucke/muse/value"
	"github.com/almerlucke/muse/value/arpeggio"
	"github.com/almerlucke/muse/value/template"
)

func noteSequence(octave notes.Note) value.Valuer[any] {
	return value.NewAnd(
		[]value.Valuer[any]{
			// value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.CMajor.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 1, 1),
			// value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.EMajor7_3.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 1, 1),
			// value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.AMinor.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 1, 1),
			// value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.CMajor7_3.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 1, 1),
			// value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.FMajor.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 1, 1),
			// value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor7_3.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 1, 1),
			// value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 1, 1),
			// value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor7_3.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 1, 1),

			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.CMajor.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.EMajor7_3.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.AMinor.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.FMajor.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 2, 2),

			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.CMajorInv1.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.EMajorInv1.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.AMinorInv1.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.FMajorInv1.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajorInv1.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 2, 2),
		}, true)
}

func addDrumTrack(p muse.Patch, moduleName string, soundBuffer *sndfile.SoundFile, tempo int, division int, lowSpeed float64, highSpeed float64, amp float64, steps value.Valuer[*swing.Step]) (muse.Messenger, muse.Module) {
	identifier := moduleName + "Speed"

	msgr := stepper.NewStepper(swing.New(tempo, division, steps), []string{identifier}).MsgrNamed(moduleName + "Stepper")

	p.AddMessenger(banger.NewTemplateGenerator([]string{moduleName}, template.Template{
		"speed": value.NewFunction(func() any { return rand.Float64()*(highSpeed-lowSpeed) + lowSpeed }),
		"bang":  true,
	}).MsgrNamed(identifier))

	return msgr, p.AddModule(player.New(soundBuffer, 1.0, amp, true).Named(moduleName))
}

func synthSettings(poly muse.Module) {
	msg := map[string]any{
		"command":         "voice",
		"osc1Mix":         0.6,
		"osc2Mix":         0.6,
		"noiseMix":        0.2,
		"osc1SineMix":     0.0,
		"osc1PulseMix":    0.0,
		"osc1SawMix":      0.0,
		"osc1TriMix":      1.0,
		"osc2SineMix":     0.0,
		"osc2PulseMix":    0.0,
		"osc2SawMix":      0.0,
		"osc2TriMix":      1.0,
		"osc1PulseWidth":  0.5,
		"osc2PulseWidth":  0.5,
		"filterResonance": 0.8,
		"osc2Tuning":      2.01,
		"filterFcMin":     50.0,
		"filterFcMax":     1000.0,
	}

	poly.ReceiveMessage(msg)
}

func createScheduler(bass muse.Messenger, kick muse.Messenger, snare muse.Messenger, hihat muse.Messenger) *scheduler.Scheduler {
	bpm := 105.0
	bassDelay := 32.0 * 60000.0 / bpm
	kickDelay := 64.0 * 60000.0 / bpm
	hihatDelay := 96.0 * 60000.0 / bpm
	removeDrumDelay := 180000.0

	bassEvent := &scheduler.Event{
		Time: bassDelay,
		Messages: []*muse.Message{{
			Address: "root",
			Content: map[string]any{
				"command":   "AddMessenger",
				"messenger": bass,
			},
		},
		},
	}

	kickEvent := &scheduler.Event{
		Time: kickDelay,
		Messages: []*muse.Message{{
			Address: "root",
			Content: map[string]any{
				"command":   "AddMessenger",
				"messenger": kick,
			},
		}, {
			Address: "root",
			Content: map[string]any{
				"command":   "AddMessenger",
				"messenger": snare,
			},
		},
		},
	}

	hihatEvent := &scheduler.Event{
		Time: hihatDelay,
		Messages: []*muse.Message{{
			Address: "root",
			Content: map[string]any{
				"command":   "AddMessenger",
				"messenger": hihat,
			},
		},
		},
	}

	removeBassAndDrumEvent := &scheduler.Event{
		Time: removeDrumDelay,
		Messages: []*muse.Message{{
			Address: "root",
			Content: map[string]any{
				"command":   "RemoveMessenger",
				"messenger": bass,
			},
		}, {
			Address: "root",
			Content: map[string]any{
				"command":   "RemoveMessenger",
				"messenger": hihat,
			},
		}, {
			Address: "root",
			Content: map[string]any{
				"command":   "RemoveMessenger",
				"messenger": snare,
			},
		}, {
			Address: "root",
			Content: map[string]any{
				"command":   "RemoveMessenger",
				"messenger": kick,
			},
		},
		},
	}

	return scheduler.NewSchedulerWithEvents([]*scheduler.Event{bassEvent, kickEvent, hihatEvent, removeBassAndDrumEvent})
}

func main() {
	root := muse.New(2)
	bpm := 105

	root.AddMessageReceiver(root, "env")

	guitarBuffer, _ := sndfile.NewSoundFile("/Users/almerlucke/Desktop/Psalm91_export/Psalm91_guitar.aiff")
	singBuffer, _ := sndfile.NewSoundFile("/Users/almerlucke/Desktop/Psalm91_export/Psalm91_voice.aiff")

	hihatSound, _ := sndfile.NewSoundFile("resources/drums/hihat/Cymatics - Humble Closed Hihat 1.wav")
	kickSound, _ := sndfile.NewSoundFile("resources/drums/kick/Cymatics - Humble Sit Down Kick - D.wav")
	snareSound, _ := sndfile.NewSoundFile("resources/drums/snare/Cymatics - Humble Institution Snare - C#.wav")

	ampEnvSetting := adsr.NewSetting(1.0, 25.0, 0.3, 80.0, 0.0, 300.0)
	filterEnvSetting := adsr.NewSetting(0.9, 25.0, 0.3, 80.0, 0.0, 300.0)

	root.AddMessenger(banger.NewTemplateGenerator([]string{"poly"}, template.Template{
		"command":   "trigger",
		"duration":  value.NewConst[any](500.0),
		"amplitude": value.NewConst[any](1.0),
		"message": template.Template{
			"frequency": noteSequence(notes.O2),
		},
	}).MsgrNamed("control"))

	bassStepper := stepper.NewStepper(
		swing.New(bpm, 2,
			value.NewSequence([]*swing.Step{{}, {Skip: true}, {}, {Skip: true}, {Shuffle: 0.1}, {Skip: true}, {Shuffle: 0.1}, {Skip: true}}),
		),
		[]string{"control"},
	).MsgrNamed("bassStepper")

	guitarPlayer := root.AddModule(player.New(guitarBuffer, 1.0, 1.0, true))
	singPlayer := root.AddModule(player.New(singBuffer, 1.0, 1.0, true))
	synth := classic.New(20, ampEnvSetting, filterEnvSetting).Named("poly").AddTo(root)

	synthSettings(synth)

	guitarChorus := root.AddModule(chorus.New(true, 20.0, 10.0, 0.3, 1.3, 0.2, nil))
	singChorus := root.AddModule(chorus.New(true, 30.0, 15.0, 0.3, 1.7, 0.4, nil))
	synthChorus := root.AddModule(chorus.New(true, 10.0, 7.0, 0.5, 3.8, 0.4, nil))

	hihatStepper, hihatPlayer := addDrumTrack(root, "hihat", hihatSound, bpm, 4, 1.875, 2.125, 0.75, value.NewSequence([]*swing.Step{
		{}, {Shuffle: 0.2}, {}, {Shuffle: 0.2}, {}, {Shuffle: 0.2}, {BurstChance: 0.3, NumBurst: 3}, {Shuffle: 0.2},
	}))

	kickStepper, kickPlayer := addDrumTrack(root, "kick", kickSound, bpm, 4, 0.875, 1.125, 1.0, value.NewSequence([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
		{}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}))

	snareStepper, snarePlayer := addDrumTrack(root, "snare", snareSound, bpm, 4, 0.875, 1.125, 1.0, value.NewSequence([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {Skip: true},
		{Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Shuffle: 0.1}, {Skip: true}, {Skip: true}, {Shuffle: 0.2},
	}))

	root.AddMessenger(createScheduler(bassStepper, kickStepper, snareStepper, hihatStepper))

	drumMixer := root.AddModule(mixer.New(3)).(*mixer.Mixer)
	drumMixer.SetMix([]float64{1.0, 1.0, 0.8})
	kickPlayer.Connect(0, drumMixer, 0)
	snarePlayer.Connect(0, drumMixer, 1)
	hihatPlayer.Connect(0, drumMixer, 2)

	drumEcho := root.AddModule(allpass.New(5000.0, 60000.0/float64(bpm)*1.25, 0.3))
	drumEchoAmp := root.AddModule(functor.NewAmp(0.3))
	drumMixer.Connect(0, drumEcho, 0)
	drumEcho.Connect(0, drumEchoAmp, 0)

	leftMixer := root.AddModule(mixer.New(4)).(*mixer.Mixer)
	rightMixer := root.AddModule(mixer.New(4)).(*mixer.Mixer)

	leftMixer.SetMix([]float64{0.7, 0.6, 0.2, 0.2})
	rightMixer.SetMix([]float64{0.7, 0.6, 0.2, 0.2})

	once := root.AddControl(once.NewControlOnce())

	once.CtrlConnect(0, guitarPlayer, 0)
	once.CtrlConnect(0, singPlayer, 0)

	guitarPlayer.Connect(0, guitarChorus, 0)
	singPlayer.Connect(0, singChorus, 0)
	synth.Connect(0, synthChorus, 0)

	guitarChorus.Connect(0, leftMixer, 0)
	singChorus.Connect(0, leftMixer, 1)
	synthChorus.Connect(0, leftMixer, 2)
	drumMixer.Connect(0, leftMixer, 3)
	drumEchoAmp.Connect(0, leftMixer, 3)

	guitarChorus.Connect(1, rightMixer, 0)
	singChorus.Connect(0, rightMixer, 1)
	synthChorus.Connect(1, rightMixer, 2)
	drumMixer.Connect(0, rightMixer, 3)
	drumEchoAmp.Connect(0, rightMixer, 3)

	leftMixer.Connect(0, root, 0)
	rightMixer.Connect(0, root, 1)

	_ = root.RenderToSoundFile("/Users/almerlucke/Desktop/psalm91_rendered", writer.AIFC, 194.0, 44100.0, true)

	// env.QuickPlayAudio()
}

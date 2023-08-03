package main

import (
	"log"
	"math/rand"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/almerlucke/muse"

	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	"github.com/almerlucke/muse/components/waveshaping"
	shaping "github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	adsrctrl "github.com/almerlucke/muse/ui/adsr"
	"github.com/almerlucke/muse/ui/theme"
	"github.com/almerlucke/muse/value"

	"github.com/almerlucke/muse/value/template"

	"github.com/almerlucke/muse/utils/notes"

	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/allpass"
	"github.com/almerlucke/muse/modules/chorus"
	"github.com/almerlucke/muse/modules/filters/moog"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/phasor"
	"github.com/almerlucke/muse/modules/player"
	"github.com/almerlucke/muse/modules/polyphony"
	"github.com/almerlucke/muse/modules/waveshaper"
)

type TestVoice struct {
	*muse.BasePatch
	ampEnv           *adsr.ADSR
	filterEnv        *adsr.ADSR
	phasor           *phasor.Phasor
	filter           *moog.Moog
	shaper           *shaping.SoftSyncTriangle
	ampEnvSetting    *adsrc.Setting
	filterEnvSetting *adsrc.Setting
}

func NewTestVoice(ampEnvSetting *adsrc.Setting, filterEnvSetting *adsrc.Setting) *TestVoice {
	testVoice := &TestVoice{
		BasePatch:        muse.NewPatch(0, 1),
		ampEnvSetting:    ampEnvSetting,
		filterEnvSetting: filterEnvSetting,
		shaper:           shaping.NewSoftSyncTriangle(1.25),
	}

	testVoice.SetSelf(testVoice)

	ampEnv := testVoice.AddModule(adsr.New(ampEnvSetting, adsrc.Duration, 1.0))
	filterEnv := testVoice.AddModule(adsr.New(filterEnvSetting, adsrc.Duration, 1.0))
	multiplier := testVoice.AddModule(functor.NewMult(2))
	filterEnvScaler := testVoice.AddModule(functor.NewScale(8000.0, 100.0))
	osc := testVoice.AddModule(phasor.New(140.0, 0.0))
	filter := testVoice.AddModule(moog.New(1400.0, 0.5, 1))
	shape := testVoice.AddModule(waveshaper.New(testVoice.shaper, 0, nil, nil))

	osc.Connect(0, shape, 0)
	shape.Connect(0, multiplier, 0)
	ampEnv.Connect(0, multiplier, 1)
	multiplier.Connect(0, filter, 0)
	filterEnv.Connect(0, filterEnvScaler, 0)
	filterEnvScaler.Connect(0, filter, 1)
	filter.Connect(0, testVoice, 0)

	testVoice.ampEnv = ampEnv.(*adsr.ADSR)
	testVoice.filterEnv = filterEnv.(*adsr.ADSR)
	testVoice.phasor = osc.(*phasor.Phasor)
	testVoice.filter = filter.(*moog.Moog)

	return testVoice
}

func (tv *TestVoice) IsActive() bool {
	return tv.ampEnv.IsActive()
}

func (tv *TestVoice) Note(duration float64, amplitude float64, message any, config *muse.Configuration) {
	msg := message.(map[string]any)

	tv.ampEnv.TriggerWithDuration(duration, amplitude)
	tv.filterEnv.TriggerWithDuration(duration, 1.0)
	tv.phasor.ReceiveMessage(msg["osc"])
}

func (tv *TestVoice) NoteOn(amplitude float64, message any, config *muse.Configuration) {
	msg := message.(map[string]any)

	tv.ampEnv.TriggerWithDuration(0, amplitude)
	tv.filterEnv.TriggerWithDuration(0, 1.0)
	tv.phasor.ReceiveMessage(msg["osc"])
}

func (tv *TestVoice) NoteOff() {
	tv.ampEnv.Release()
	tv.filterEnv.Release()
}

func (tv *TestVoice) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	if shaper, ok := content["shaper"].(float64); ok {
		tv.shaper.SetA1(shaper)
	}

	if filterFrequency, ok := content["filterFrequency"].(float64); ok {
		tv.filter.SetFrequency(filterFrequency)
	}

	if filterResonance, ok := content["filterResonance"].(float64); ok {
		tv.filter.SetResonance(filterResonance)
	}

	if filterDrive, ok := content["filterDrive"].(float64); ok {
		tv.filter.SetDrive(filterDrive)
	}

	return nil
}

func addDrumTrack(p muse.Patch, moduleName string, soundBuffer *io.SoundFile, tempo int, division int, lowSpeed float64, highSpeed float64, amp float64, steps value.Valuer[*swing.Step]) muse.Module {
	identifier := moduleName + "Speed"

	p.AddMessenger(stepper.NewStepper(swing.New(tempo, division, steps), []string{identifier}))

	p.AddMessenger(banger.NewTemplateGenerator([]string{moduleName}, template.Template{
		"speed": value.NewFunction(func() any { return rand.Float64()*(highSpeed-lowSpeed) + lowSpeed }),
		"bang":  true,
	}).MsgrNamed(identifier))

	return p.AddModule(player.New(soundBuffer, 1.0, amp, true).Named(moduleName))
}

func main() {
	root := muse.New(2)

	ampEnvSetting := adsrc.NewSetting(1.0, 5.0, 0.2, 37.0, 0.0, 1630.0)
	filterEnvSetting := adsrc.NewSetting(1.0, 5.0, 0.43, 50.0, 0.0, 1700.0)
	ampEnvControl := adsrctrl.NewADSRControl("Amplitude ADSR", ampEnvSetting)
	filterEnvControl := adsrctrl.NewADSRControl("Filter ADSR", filterEnvSetting)

	voices := []polyphony.Voice{}
	for i := 0; i < 20; i++ {
		voice := NewTestVoice(ampEnvSetting, filterEnvSetting)
		voices = append(voices, voice)
	}

	bpm := 80

	// milliPerBeat := 60000.0 / bpm

	poly := root.AddModule(polyphony.New(1, voices).Named("polyphony"))
	allpass := root.AddModule(allpass.New(50, 50, 0.3))

	sineTable := shaping.NewNormalizedSineTable(512)

	targetShaper := lfo.NewTarget("polyphony", shaping.NewSeries(sineTable, shaping.NewLinear(0.7, 1.0)), "shaper", template.Template{
		"command": "voice",
		"shaper":  template.NewParameter("shaper", nil),
	})

	targetFilter := lfo.NewTarget("polyphony", shaping.NewSeries(sineTable, shaping.NewLinear(0.1, 0.05)), "adsrDecayLevel", template.Template{
		"command":        "voice",
		"adsrDecayLevel": template.NewParameter("adsrDecayLevel", nil),
	})

	root.AddMessenger(lfo.NewLFO(0.03, []*lfo.Target{targetShaper}).MsgrNamed("lfo1"))
	root.AddMessenger(lfo.NewLFO(0.13, []*lfo.Target{targetFilter}).MsgrNamed("lfo2"))

	root.AddMessenger(banger.NewTemplateGenerator([]string{"polyphony"}, template.Template{
		"command":   "trigger",
		"duration":  value.NewSequence([]any{125.0, 125.0, 125.0, 250.0, 125.0, 250.0, 125.0, 125.0, 125.0, 250.0, 125.0}),
		"amplitude": value.NewSequence([]any{0.6, 0.3, 0.6, 0.5, 0.4, 0.3, 0.4, 0.5, 0.6, 0.4, 0.2}),
		"message": template.Template{
			"osc": template.Template{
				"frequency": value.NewTransform[any](value.NewSequence([]any{
					notes.Mtof(60), notes.Mtof(67), notes.Mtof(62), notes.Mtof(69), notes.Mtof(64), notes.Mtof(71),
					notes.Mtof(66), notes.Mtof(61), notes.Mtof(68), notes.Mtof(63), notes.Mtof(70), notes.Mtof(65),
				}),
					value.TFunc[any](func(v any) any { return v.(float64) / 2.0 })),
				"phase": 0.0,
			},
		},
	}).MsgrNamed("template1"))

	root.AddMessenger(banger.NewTemplateGenerator([]string{"polyphony"}, template.Template{
		"command":   "trigger",
		"duration":  value.NewSequence([]any{375.0, 500.0, 375.0, 1000.0, 375.0, 250.0}),
		"amplitude": value.NewSequence([]any{1.0, 1.0, 0.6, 0.6, 1.0, 1.0, 0.6, 1.0}),
		"message": template.Template{
			"osc": template.Template{
				"frequency": value.NewTransform[any](value.NewSequence([]any{
					notes.Mtof(67), notes.Mtof(62), notes.Mtof(69), notes.Mtof(64), notes.Mtof(71), notes.Mtof(66),
					notes.Mtof(61), notes.Mtof(68), notes.Mtof(63), notes.Mtof(70), notes.Mtof(65), notes.Mtof(72),
				}),
					value.TFunc[any](func(v any) any { return v.(float64) / 4.0 })),
				"phase": 0.375,
			},
		},
	}).MsgrNamed("template2"))

	root.AddMessenger(stepper.NewStepper(
		swing.New(bpm, 4, value.NewSequence(
			[]*swing.Step{{}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {SkipChance: 0.3},
				{}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {}, {SkipChance: 0.3}, {SkipChance: 0.3}},
		)),
		[]string{"template1"},
	))

	root.AddMessenger(stepper.NewStepper(
		swing.New(bpm, 2, value.NewSequence(
			[]*swing.Step{{Skip: true}, {}, {}, {Skip: true}, {Skip: true}, {}, {}, {SkipChance: 0.3}},
		)),
		[]string{"template2"},
	))

	hihatSound, _ := io.NewSoundFile("resources/drums/hihat/Cymatics - Humble Closed Hihat 1.wav")
	kickSound, _ := io.NewSoundFile("resources/drums/kick/Cymatics - Humble Triple Kick - E.wav")
	snareSound, _ := io.NewSoundFile("resources/drums/clap/Cymatics - Humble Stars Clap.wav")
	bassSound, _ := io.NewSoundFile("resources/drums/808/Cymatics - Humble 808 5 - G.wav")
	rideSound, _ := io.NewSoundFile("resources/drums/hihat/Cymatics - Humble Open Hihat 2.wav")
	waterSound, _ := io.NewSoundFile("resources/sounds/Cymatics - Orchid Live Recording - Waves.wav")
	swirlSound, _ := io.NewSoundFile("resources/sounds/Cymatics - Orchid KEYS Swirl (C).wav")
	vocalSound, _ := io.NewSoundFile("resources/sounds/Cymatics - Blurry Vocal - 80 BPM F Min.wav")

	hihatPlayer := addDrumTrack(root, "hihat", hihatSound, bpm, 8, 1.875, 2.125, 0.75, value.NewSequence([]*swing.Step{
		{}, {Shuffle: 0.1}, {SkipChance: 0.3, BurstChance: 1.0, NumBurst: 3}, {}, {Skip: true}, {Shuffle: 0.1}, {}, {SkipChance: 0.4}, {Skip: true}, {Skip: true},
	}))

	kickPlayer := addDrumTrack(root, "kick", kickSound, bpm, 4, 0.875, 1.125, 1.0, value.NewSequence([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {SkipChance: 0.4}, {}, {Skip: true}, {Skip: true},
	}))

	snarePlayer := addDrumTrack(root, "snare", snareSound, bpm, 2, 0.875, 1.125, 1.0, value.NewSequence([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {BurstChance: 0.2, NumBurst: 2}, {Skip: true}, {Skip: true}, {Skip: true}, {},
	}))

	bassPlayer := addDrumTrack(root, "bass", bassSound, bpm, 1, 0.875, 1.125, 1.0, value.NewSequence([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}))

	ridePlayer := addDrumTrack(root, "ride", rideSound, bpm, 2, 0.875, 1.25, 0.3, value.NewSequence([]*swing.Step{
		{Skip: true}, {Skip: true}, {BurstChance: 0.2, NumBurst: 4}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}))

	waterPlayer := addDrumTrack(root, "water", waterSound, int(float64(bpm)*0.125), 2, 0.875, 1.25, 0.5, value.NewSequence([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}))

	swirlPlayer := addDrumTrack(root, "swirl", swirlSound, int(float64(bpm)*0.5), 1.0, 0.875, 1.25, 0.2, value.NewSequence([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}))

	vocalPlayer := addDrumTrack(root, "vocal", vocalSound, int(float64(bpm)*0.125), 1.0, 0.975, 1.025, 0.2, value.NewSequence([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}))

	mult := root.AddModule(functor.NewAmp(0.5))

	chor1 := root.AddModule(chorus.New(false, 15, 10, 0.9, 1.3, 0.3, waveshaping.NewSineTable(512)))

	kickPlayer.Connect(0, mult, 0)
	hihatPlayer.Connect(0, mult, 0)
	snarePlayer.Connect(0, mult, 0)
	bassPlayer.Connect(0, mult, 0)
	ridePlayer.Connect(0, mult, 0)
	waterPlayer.Connect(0, mult, 0)
	swirlPlayer.Connect(0, mult, 0)
	vocalPlayer.Connect(0, mult, 0)

	poly.Connect(0, chor1, 0)
	chor1.Connect(0, allpass, 0)
	chor1.Connect(0, root, 0)
	mult.Connect(0, root, 0)
	mult.Connect(0, root, 1)
	allpass.Connect(0, root, 1)

	err := root.InitializeAudio()
	if err != nil {
		log.Fatalf("error opening portaudio stream, %v", err)
	}

	defer root.TerminateAudio()

	a := app.New()

	a.Settings().SetTheme(&theme.Theme{})

	w := a.NewWindow("Muse")

	w.Resize(fyne.Size{
		Width:  700,
		Height: 400,
	})

	w.SetContent(
		container.NewVBox(
			container.NewHBox(
				widget.NewButton("Start", func() {
					// env.SynthesizeToFile("/Users/almerlucke/Desktop/waterFlow.aiff", 240.0, env.Config.SampleRate, sndfile.SF_FORMAT_AIFF)
					root.StartAudio()
				}),
				widget.NewButton("Stop", func() {
					root.StopAudio()
				}),
				// widget.NewButton("Notes Off", func() {
				// 	poly.(*polyphony.Polyphony).AllNotesOff()
				// }),
			),
			container.NewHBox(
				ampEnvControl.UI(),
				filterEnvControl.UI(),
			),
		),
	)

	w.ShowAndRun()
}

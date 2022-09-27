package main

import (
	"log"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/gordonklaus/portaudio"

	"github.com/almerlucke/muse"

	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	shaping "github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/ui"
	adsrctrl "github.com/almerlucke/muse/ui/controls/adsr"
	"github.com/almerlucke/muse/ui/theme"
	"github.com/almerlucke/muse/value"

	"github.com/almerlucke/muse/value/template"

	"github.com/almerlucke/muse/utils/notes"

	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/allpass"
	"github.com/almerlucke/muse/modules/filters/moog"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/monitor"
	"github.com/almerlucke/muse/modules/phasor"
	"github.com/almerlucke/muse/modules/player"
	"github.com/almerlucke/muse/modules/polyphony"
	"github.com/almerlucke/muse/modules/waveshaper"
)

type TestVoice struct {
	*muse.BasePatch
	ampEnv             *adsr.ADSR
	filterEnv          *adsr.ADSR
	phasor             *phasor.Phasor
	filter             *moog.Moog
	shaper             *shaping.Chain
	ampStepProvider    adsrctrl.ADSRStepProvider
	filterStepProvider adsrctrl.ADSRStepProvider
}

func NewTestVoice(config *muse.Configuration, ampStepProvider adsrctrl.ADSRStepProvider, filterStepProvider adsrctrl.ADSRStepProvider) *TestVoice {
	testVoice := &TestVoice{
		BasePatch:          muse.NewPatch(0, 1, config, ""),
		ampStepProvider:    ampStepProvider,
		filterStepProvider: filterStepProvider,
		shaper:             shaping.NewSoftSyncTriangle(),
	}

	ampEnv := testVoice.AddModule(adsr.NewADSR(ampStepProvider.ADSRSteps(), adsrc.Absolute, adsrc.Duration, 1.0, config, "ampAdsr"))
	filterEnv := testVoice.AddModule(adsr.NewADSR(filterStepProvider.ADSRSteps(), adsrc.Absolute, adsrc.Duration, 1.0, config, "filterAdsr"))
	multiplier := testVoice.AddModule(functor.NewFunctor(2, functor.FunctorMult, config, ""))
	filterEnvScaler := testVoice.AddModule(functor.NewFunctor(1, func(in []float64) float64 { return in[0]*8000.0 + 100.0 }, config, ""))
	osc := testVoice.AddModule(phasor.NewPhasor(140.0, 0.0, config, "osc"))
	filter := testVoice.AddModule(moog.NewMoog(1400.0, 0.5, 1, config, "filter"))
	shape := testVoice.AddModule(waveshaper.NewWaveShaper(testVoice.shaper, 0, nil, nil, config, "shaper"))

	muse.Connect(osc, 0, shape, 0)
	muse.Connect(shape, 0, multiplier, 0)
	muse.Connect(ampEnv, 0, multiplier, 1)
	muse.Connect(multiplier, 0, filter, 0)
	muse.Connect(filterEnv, 0, filterEnvScaler, 0)
	muse.Connect(filterEnvScaler, 0, filter, 1)
	muse.Connect(filter, 0, testVoice, 0)

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

	tv.ampEnv.TriggerFull(duration, amplitude, tv.ampStepProvider.ADSRSteps(), adsrc.Absolute, adsrc.Duration)
	tv.filterEnv.TriggerFull(duration, 1.0, tv.filterStepProvider.ADSRSteps(), adsrc.Absolute, adsrc.Duration)
	tv.phasor.ReceiveMessage(msg["osc"])
}

func (tv *TestVoice) NoteOn(amplitude float64, message any, config *muse.Configuration) {
	msg := message.(map[string]any)

	tv.ampEnv.TriggerFull(0, amplitude, tv.ampStepProvider.ADSRSteps(), adsrc.Absolute, adsrc.Duration)
	tv.filterEnv.TriggerFull(0, 1.0, tv.filterStepProvider.ADSRSteps(), adsrc.Absolute, adsrc.Duration)
	tv.phasor.ReceiveMessage(msg["osc"])
}

func (tv *TestVoice) NoteOff() {
	tv.ampEnv.Release()
	tv.filterEnv.Release()
}

func (tv *TestVoice) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	if shaper, ok := content["shaper"].(float64); ok {
		tv.shaper.SetSoftSyncA1(shaper)
	}

	if adsrAttackDuration, ok := content["adsrAttackDuration"].(float64); ok {
		tv.filterStepProvider.(*adsrctrl.ADSRControl).SetAttackDuration(adsrAttackDuration)
	}

	if adsrAttackLevel, ok := content["adsrAttackLevel"].(float64); ok {
		tv.filterStepProvider.(*adsrctrl.ADSRControl).SetAttackLevel(adsrAttackLevel)
	}

	if adsrAttackShape, ok := content["adsrAttackShape"].(float64); ok {
		tv.filterStepProvider.(*adsrctrl.ADSRControl).SetAttackShape(adsrAttackShape)
	}

	if adsrDecayDuration, ok := content["adsrDecayDuration"].(float64); ok {
		tv.filterStepProvider.(*adsrctrl.ADSRControl).SetDecayDuration(adsrDecayDuration)
	}

	if adsrDecayLevel, ok := content["adsrDecayLevel"].(float64); ok {
		tv.filterStepProvider.(*adsrctrl.ADSRControl).SetDecayLevel(adsrDecayLevel)
	}

	if adsrDecayShape, ok := content["adsrDecayShape"].(float64); ok {
		tv.filterStepProvider.(*adsrctrl.ADSRControl).SetDecayShape(adsrDecayShape)
	}

	if adsrReleaseDuration, ok := content["adsrReleaseDuration"].(float64); ok {
		tv.filterStepProvider.(*adsrctrl.ADSRControl).SetReleaseDuration(adsrReleaseDuration)
	}

	if adsrReleaseShape, ok := content["adsrReleaseShape"].(float64); ok {
		tv.filterStepProvider.(*adsrctrl.ADSRControl).SetReleaseShape(adsrReleaseShape)
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

func addDrumTrack(env *muse.Environment, moduleName string, soundBuffer *io.SoundFileBuffer, tempo float64, division float64, lowSpeed float64, highSpeed float64, amp float64, steps value.Valuer[*swing.Step]) muse.Module {
	identifier := moduleName + "Speed"

	env.AddMessenger(stepper.NewStepper(swing.New(value.NewConst(tempo), value.NewConst(division), steps), []string{identifier}, ""))

	env.AddMessenger(banger.NewTemplateGenerator([]string{moduleName}, template.Template{
		"speed": value.NewFunction(func() any { return rand.Float64()*(highSpeed-lowSpeed) + lowSpeed }),
	}, identifier))

	return env.AddModule(player.NewPlayer(soundBuffer, 1.0, amp, true, env.Config, moduleName))
}

func main() {
	rand.Seed(time.Now().UnixNano())

	env := muse.NewEnvironment(2, 3*44100, 1024)

	ampEnvControl := adsrctrl.NewADSRControl("Amplitude ADSR")
	filterEnvControl := adsrctrl.NewADSRControl("Filter ADSR")

	ampEnvControl.SetAttackDuration(5)
	ampEnvControl.SetDecayDuration(37)
	ampEnvControl.SetDecayLevel(0.2)
	ampEnvControl.SetReleaseDuration(1630)
	ampEnvControl.SetReleaseShape(-0.35)

	filterEnvControl.SetAttackDuration(5)
	filterEnvControl.SetAttackLevel(0.43)
	filterEnvControl.SetDecayDuration(50.0)
	filterEnvControl.SetReleaseDuration(1700)

	monitor := monitor.NewMonitor(200, 100, env.Config)

	voices := []polyphony.Voice{}
	for i := 0; i < 20; i++ {
		voice := NewTestVoice(env.Config, ampEnvControl, filterEnvControl)
		voices = append(voices, voice)
	}

	bpm := 80.0

	// milliPerBeat := 60000.0 / bpm

	env.AddModule(monitor)
	poly := env.AddModule(polyphony.NewPolyphony(1, voices, env.Config, "polyphony"))
	allpass := env.AddModule(allpass.NewAllpass(50, 50, 0.3, env.Config, "allpass"))

	sineTable := shaping.NewNormalizedSineTable(512)

	targetShaper := lfo.NewTarget("polyphony", shaping.NewChain(sineTable, shaping.NewLinear(0.7, 1.0)), "shaper", template.Template{
		"command": "voice",
		"shaper":  template.NewParameter("shaper", nil),
	})

	targetFilter := lfo.NewTarget("polyphony", shaping.NewChain(sineTable, shaping.NewLinear(0.1, 0.05)), "adsrDecayLevel", template.Template{
		"command":        "voice",
		"adsrDecayLevel": template.NewParameter("adsrDecayLevel", nil),
	})

	env.AddMessenger(lfo.NewLFO(0.03, []*lfo.Target{targetShaper}, env.Config, "lfo1"))
	env.AddMessenger(lfo.NewLFO(0.13, []*lfo.Target{targetFilter}, env.Config, "lfo2"))

	env.AddMessenger(banger.NewTemplateGenerator([]string{"polyphony"}, template.Template{
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
	}, "template1"))

	env.AddMessenger(banger.NewTemplateGenerator([]string{"polyphony"}, template.Template{
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
	}, "template2"))

	env.AddMessenger(stepper.NewStepper(
		swing.New(value.NewConst(bpm), value.NewConst(4.0), value.NewSequence(
			[]*swing.Step{{}, {Skip: true}, {Shuffle: 0.2}, {Skip: true}, {}, {Skip: true}, {Shuffle: 0.2, ShuffleRand: 0.1}, {SkipFactor: 0.3},
				{}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Shuffle: 0.2}, {SkipFactor: 0.3}, {SkipFactor: 0.3}},
		)),
		[]string{"template1"}, "",
	))

	env.AddMessenger(stepper.NewStepper(
		swing.New(value.NewConst(bpm), value.NewConst(2.0), value.NewSequence(
			[]*swing.Step{{Skip: true}, {}, {Shuffle: 0.2}, {Skip: true}, {Skip: true}, {}, {Shuffle: 0.2, ShuffleRand: 0.1}, {SkipFactor: 0.3}},
		)),
		[]string{"template2"}, "",
	))

	hihatSound, _ := io.NewSoundFileBuffer("resources/drums/hihat/Cymatics - Humble Closed Hihat 1.wav")
	kickSound, _ := io.NewSoundFileBuffer("resources/drums/kick/Cymatics - Humble Triple Kick - E.wav")
	snareSound, _ := io.NewSoundFileBuffer("resources/drums/clap/Cymatics - Humble Stars Clap.wav")
	bassSound, _ := io.NewSoundFileBuffer("resources/drums/808/Cymatics - Humble 808 5 - G.wav")
	rideSound, _ := io.NewSoundFileBuffer("resources/drums/hihat/Cymatics - Humble Open Hihat 2.wav")
	waterSound, _ := io.NewSoundFileBuffer("resources/sounds/Cymatics - Orchid Live Recording - Waves.wav")
	swirlSound, _ := io.NewSoundFileBuffer("resources/sounds/Cymatics - Orchid KEYS Swirl (C).wav")
	vocalSound, _ := io.NewSoundFileBuffer("resources/sounds/Cymatics - Blurry Vocal - 80 BPM F Min.wav")

	hihatPlayer := addDrumTrack(env, "hihat", hihatSound, bpm, 8.0, 1.875, 2.125, 0.5, value.NewSequence([]*swing.Step{
		{}, {Shuffle: 0.3}, {Skip: true}, {Shuffle: 0.3, ShuffleRand: 0.2}, {Skip: true}, {Shuffle: 0.1}, {}, {SkipFactor: 0.4, Shuffle: 0.2}, {Skip: true}, {Skip: true},
	}),
	)

	kickPlayer := addDrumTrack(env, "kick", kickSound, bpm, 4.0, 0.875, 1.125, 1.0, value.NewSequence([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {SkipFactor: 0.4}, {Shuffle: 0.2}, {Skip: true}, {Skip: true},
	}),
	)

	snarePlayer := addDrumTrack(env, "snare", snareSound, bpm, 2.0, 0.875, 1.125, 1.0, value.NewSequence([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Shuffle: 0.1, ShuffleRand: 0.1}, {Skip: true}, {Skip: true}, {Skip: true}, {},
	}))

	bassPlayer := addDrumTrack(env, "bass", bassSound, bpm, 1.0, 0.875, 1.125, 1.0, value.NewSequence([]*swing.Step{
		{Shuffle: 0.2, ShuffleRand: 0.2}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}))

	ridePlayer := addDrumTrack(env, "ride", rideSound, bpm, 2.0, 0.875, 1.25, 0.3, value.NewSequence([]*swing.Step{
		{Skip: true}, {Skip: true}, {Shuffle: 0.2, ShuffleRand: 0.2}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}))

	waterPlayer := addDrumTrack(env, "water", waterSound, bpm*0.125, 2.0, 0.875, 1.25, 0.5, value.NewSequence([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}))

	swirlPlayer := addDrumTrack(env, "swirl", swirlSound, bpm*0.5, 1.0, 0.875, 1.25, 0.2, value.NewSequence([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}))

	vocalPlayer := addDrumTrack(env, "vocal", vocalSound, bpm*0.125, 1.0, 0.975, 1.025, 0.2, value.NewSequence([]*swing.Step{
		{}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true}, {Skip: true},
	}))

	mult := env.AddModule(functor.NewFunctor(1, func(v []float64) float64 { return v[0] * 0.5 }, env.Config, ""))

	muse.Connect(kickPlayer, 0, mult, 0)
	muse.Connect(hihatPlayer, 0, mult, 0)
	muse.Connect(snarePlayer, 0, mult, 0)
	muse.Connect(bassPlayer, 0, mult, 0)
	muse.Connect(ridePlayer, 0, mult, 0)
	muse.Connect(waterPlayer, 0, mult, 0)
	muse.Connect(swirlPlayer, 0, mult, 0)
	muse.Connect(vocalPlayer, 0, mult, 0)

	muse.Connect(poly, 0, allpass, 0)
	muse.Connect(poly, 0, monitor, 0)
	muse.Connect(poly, 0, env, 0)
	muse.Connect(mult, 0, env, 0)
	muse.Connect(mult, 0, env, 1)
	muse.Connect(allpass, 0, env, 1)

	portaudio.Initialize()
	defer portaudio.Terminate()

	stream, err := env.PortaudioStream()
	if err != nil {
		log.Fatalf("error opening portaudio stream, %v", err)
	}

	defer stream.Close()

	a := app.New()

	a.Settings().SetTheme(&theme.Theme{})

	w := a.NewWindow("Muse")

	w.Resize(fyne.Size{
		Width:  700,
		Height: 400,
	})

	// keyMap := map[string]float64{}

	// baseNote := 36

	// keyMap["`"] = utils.Mtof(baseNote + 0)
	// keyMap["Z"] = utils.Mtof(baseNote + 1)
	// keyMap["X"] = utils.Mtof(baseNote + 2)
	// keyMap["C"] = utils.Mtof(baseNote + 3)
	// keyMap["V"] = utils.Mtof(baseNote + 4)
	// keyMap["B"] = utils.Mtof(baseNote + 5)
	// keyMap["N"] = utils.Mtof(baseNote + 6)
	// keyMap["M"] = utils.Mtof(baseNote + 7)
	// keyMap[","] = utils.Mtof(baseNote + 8)
	// keyMap["."] = utils.Mtof(baseNote + 9)
	// keyMap["/"] = utils.Mtof(baseNote + 10)

	// keyMap["A"] = utils.Mtof(baseNote + 11)
	// keyMap["S"] = utils.Mtof(baseNote + 12)
	// keyMap["D"] = utils.Mtof(baseNote + 13)
	// keyMap["F"] = utils.Mtof(baseNote + 14)
	// keyMap["G"] = utils.Mtof(baseNote + 15)
	// keyMap["H"] = utils.Mtof(baseNote + 16)
	// keyMap["J"] = utils.Mtof(baseNote + 17)
	// keyMap["K"] = utils.Mtof(baseNote + 18)
	// keyMap["L"] = utils.Mtof(baseNote + 19)
	// keyMap[";"] = utils.Mtof(baseNote + 20)
	// keyMap["'"] = utils.Mtof(baseNote + 21)
	// keyMap["\\"] = utils.Mtof(baseNote + 22)

	// keyMap["Q"] = utils.Mtof(baseNote + 23)
	// keyMap["W"] = utils.Mtof(baseNote + 24)
	// keyMap["E"] = utils.Mtof(baseNote + 25)
	// keyMap["R"] = utils.Mtof(baseNote + 26)
	// keyMap["T"] = utils.Mtof(baseNote + 27)
	// keyMap["Y"] = utils.Mtof(baseNote + 28)
	// keyMap["U"] = utils.Mtof(baseNote + 29)
	// keyMap["I"] = utils.Mtof(baseNote + 30)
	// keyMap["O"] = utils.Mtof(baseNote + 31)
	// keyMap["P"] = utils.Mtof(baseNote + 32)
	// keyMap["["] = utils.Mtof(baseNote + 33)
	// keyMap["]"] = utils.Mtof(baseNote + 34)

	// keyMap["1"] = utils.Mtof(baseNote + 35)
	// keyMap["2"] = utils.Mtof(baseNote + 36)
	// keyMap["3"] = utils.Mtof(baseNote + 37)
	// keyMap["4"] = utils.Mtof(baseNote + 38)
	// keyMap["5"] = utils.Mtof(baseNote + 39)
	// keyMap["6"] = utils.Mtof(baseNote + 40)
	// keyMap["7"] = utils.Mtof(baseNote + 41)
	// keyMap["8"] = utils.Mtof(baseNote + 42)
	// keyMap["9"] = utils.Mtof(baseNote + 43)
	// keyMap["0"] = utils.Mtof(baseNote + 44)
	// keyMap["-"] = utils.Mtof(baseNote + 45)
	// keyMap["="] = utils.Mtof(baseNote + 46)

	// if deskCanvas, ok := w.Canvas().(desktop.Canvas); ok {
	// 	deskCanvas.SetOnKeyDown(func(k *fyne.KeyEvent) {
	// 		if f, ok := keyMap[string(k.Name)]; ok {
	// 			poly.ReceiveMessage(map[string]any{
	// 				"command":   "trigger",
	// 				"noteOn":    string(k.Name),
	// 				"amplitude": 1.0,
	// 				"message": map[string]any{
	// 					"osc": map[string]any{
	// 						"frequency": f,
	// 					},
	// 				},
	// 			})
	// 		}
	// 	})

	// 	deskCanvas.SetOnKeyUp(func(k *fyne.KeyEvent) {
	// 		if _, ok := keyMap[string(k.Name)]; ok {
	// 			poly.ReceiveMessage(map[string]any{
	// 				"command": "trigger",
	// 				"noteOff": string(k.Name),
	// 			})
	// 		}
	// 	})
	// }

	w.SetContent(
		container.NewVBox(
			container.NewHBox(
				widget.NewButton("Start", func() {
					// env.SynthesizeToFile("/Users/almerlucke/Desktop/waterFlow.aiff", 240.0, env.Config.SampleRate, sndfile.SF_FORMAT_AIFF)
					stream.Start()
				}),
				widget.NewButton("Stop", func() {
					stream.Stop()
				}),
				// widget.NewButton("Notes Off", func() {
				// 	poly.(*polyphony.Polyphony).AllNotesOff()
				// }),
			),
			container.NewHBox(
				widget.NewCard("Monitor", "", fyne.NewContainerWithLayout(ui.NewFixedSizeLayout(fyne.NewSize(200, 100)), monitor.UI())),
			),
			container.NewHBox(
				ampEnvControl.UI(),
				filterEnvControl.UI(),
			),
		),
	)

	w.ShowAndRun()
}

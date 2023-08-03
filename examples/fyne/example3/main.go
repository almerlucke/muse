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
	shaping "github.com/almerlucke/muse/components/waveshaping"
	adsrctrl "github.com/almerlucke/muse/ui/adsr"

	"github.com/almerlucke/muse/io"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"

	"github.com/almerlucke/muse/ui/theme"

	"github.com/almerlucke/muse/value"
	"github.com/almerlucke/muse/value/template"

	"github.com/almerlucke/muse/utils/notes"

	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/allpass"
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
	shaper           *shaping.PulseWidthMod
	ampEnvSetting    *adsrc.Setting
	filterEnvSetting *adsrc.Setting
}

func NewTestVoice(ampEnvSetting *adsrc.Setting, filterEnvSetting *adsrc.Setting) *TestVoice {
	testVoice := &TestVoice{
		BasePatch:        muse.NewPatch(0, 1),
		ampEnvSetting:    ampEnvSetting,
		filterEnvSetting: filterEnvSetting,
		shaper:           shaping.NewPulseWidthMod(),
	}

	testVoice.SetSelf(testVoice)

	ampEnv := testVoice.AddModule(adsr.New(ampEnvSetting, adsrc.Duration, 1.0))
	filterEnv := testVoice.AddModule(adsr.New(filterEnvSetting, adsrc.Duration, 1.0))
	multiplier := testVoice.AddModule(functor.NewMult(2))
	filterEnvScaler := testVoice.AddModule(functor.NewScale(5000.0, 30.0))
	osc := testVoice.AddModule(phasor.New(140.0, 0.0))
	filter := testVoice.AddModule(moog.New(1400.0, 0.8, 1.5))
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

	tv.ampEnv.TriggerFull(duration, amplitude, tv.ampEnvSetting, adsrc.Duration)
	tv.filterEnv.TriggerFull(duration, 1.0, tv.filterEnvSetting, adsrc.Duration)
	tv.phasor.ReceiveMessage(msg["osc"])
}

func (tv *TestVoice) NoteOn(amplitude float64, message any, config *muse.Configuration) {
	msg := message.(map[string]any)

	tv.ampEnv.TriggerFull(0, amplitude, tv.ampEnvSetting, adsrc.NoteOff)
	tv.filterEnv.TriggerFull(0, 1.0, tv.filterEnvSetting, adsrc.NoteOff)
	tv.phasor.ReceiveMessage(msg["osc"])
}

func (tv *TestVoice) NoteOff() {
	tv.ampEnv.Release()
	tv.filterEnv.Release()
}

func (tv *TestVoice) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	if shaper, ok := content["shaper"].(float64); ok {
		tv.shaper.SetWidth(shaper)
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

func addDrumTrack(p muse.Patch, moduleName string, soundFile *io.SoundFile, tempo int, division int, lowSpeed float64, highSpeed float64, steps value.Valuer[*swing.Step]) muse.Module {
	identifier := moduleName + "Speed"

	p.AddMessenger(stepper.NewStepper(swing.New(tempo, division, steps), []string{identifier}))

	p.AddMessenger(banger.NewTemplateGenerator([]string{moduleName}, template.Template{
		"speed": value.NewFunction(func() any { return rand.Float64()*(highSpeed-lowSpeed) + lowSpeed }),
		"bang":  true,
	}).MsgrNamed(identifier))

	return p.AddModule(player.New(soundFile, 1.0, 1.0, true).Named(moduleName))
}

type Nums []float64

func main() {
	root := muse.New(2)

	ampEnvSetting := adsrc.NewSetting(1.0, 5.0, 0.2, 5.0, 0.0, 1500.0)
	filterEnvSetting := adsrc.NewSetting(1.0, 5.0, 0.2, 5.0, 0.0, 1500.0)
	ampEnvControl := adsrctrl.NewADSRControl("Amplitude ADSR", ampEnvSetting)
	filterEnvControl := adsrctrl.NewADSRControl("Filter ADSR", filterEnvSetting)

	voices := []polyphony.Voice{}
	for i := 0; i < 40; i++ {
		voice := NewTestVoice(ampEnvSetting, filterEnvSetting)
		voices = append(voices, voice)
	}

	bpm := 80

	milliPerBeat := 60000.0 / float64(bpm)

	hihatSound, _ := io.NewSoundFile("resources/drums/hihat/Cymatics - Humble Closed Hihat 1.wav")
	kickSound, _ := io.NewSoundFile("resources/drums/kick/Cymatics - Humble Friday Kick - E.wav")
	snareSound, _ := io.NewSoundFile("resources/drums/snare/Cymatics - Humble Adequate Snare - E.wav")

	hihatPlayer := addDrumTrack(root, "hihat", hihatSound, bpm, 8, 0.875, 1.125, value.NewAnd([]value.Valuer[*swing.Step]{
		value.NewRepeat[*swing.Step](value.NewSequenceNC([]*swing.Step{
			{}, {Shuffle: 0.3}, {Skip: true}, {Shuffle: 0.3, ShuffleRand: 0.2}, {Skip: true}, {Shuffle: 0.1}, {}, {SkipChance: 0.4, Shuffle: 0.2}, {Skip: true}, {Skip: true},
		}), 2, 3),
		value.NewRepeat[*swing.Step](value.NewSequenceNC([]*swing.Step{
			{}, {Shuffle: 0.3}, {Skip: true}, {Skip: true}, {Skip: true}, {Shuffle: 0.3, ShuffleRand: 0.2}, {Skip: true}, {Shuffle: 0.1}, {SkipChance: 0.4}, {SkipChance: 0.4}, {SkipChance: 0.4, Shuffle: 0.2}, {Skip: true}, {Skip: true},
		}), 1, 2),
	}, true))

	kickPlayer := addDrumTrack(root, "kick", kickSound, bpm, 4, 0.875, 1.125, value.NewAnd([]value.Valuer[*swing.Step]{
		value.NewRepeat[*swing.Step](value.NewSequenceNC([]*swing.Step{
			{}, {Skip: true}, {Skip: true}, {Skip: true}, {}, {Skip: true}, {Skip: true}, {SkipChance: 0.4}, {Shuffle: 0.2}, {Skip: true}, {Skip: true},
		}), 2, 3),
		value.NewRepeat[*swing.Step](value.NewSequenceNC([]*swing.Step{
			{}, {Skip: true}, {Shuffle: 0.2}, {Skip: true}, {SkipChance: 0.4}, {Skip: true}, {Skip: true}, {SkipChance: 0.4}, {Shuffle: 0.2}, {Skip: true}, {Skip: true}, {Skip: true},
		}), 1, 2),
	}, true))

	snarePlayer := addDrumTrack(root, "snare", snareSound, bpm, 2, 0.875, 1.125, value.NewSequence([]*swing.Step{
		{Skip: true}, {Skip: true}, {Skip: true}, {Shuffle: 0.1, ShuffleRand: 0.1},
	}))

	mult := root.AddModule(functor.NewAmp(0.3))
	poly := root.AddModule(polyphony.New(1, voices).Named("polyphony"))
	allpass := root.AddModule(allpass.New(milliPerBeat*1.5, milliPerBeat*1.5, 0.3))

	sineTable := shaping.NewNormalizedSineTable(512)

	targetShaper := lfo.NewTarget("polyphony", shaping.NewSeries(sineTable, shaping.NewLinear(0.8, 0.1)), "shaper", template.Template{
		"command": "voice",
		"shaper":  template.NewParameter("shaper", nil),
	})

	targetFilter := lfo.NewTarget("polyphony", shaping.NewSeries(sineTable, shaping.NewLinear(0.4, 0.1)), "adsrDecayLevel", template.Template{
		"command":        "voice",
		"adsrDecayLevel": template.NewParameter("adsrDecayLevel", nil),
	})

	root.AddMessenger(lfo.NewLFO(0.05, []*lfo.Target{targetShaper}).MsgrNamed("lfo1"))
	root.AddMessenger(lfo.NewLFO(0.1, []*lfo.Target{targetFilter}).MsgrNamed("lfo2"))

	root.AddMessenger(banger.NewTemplateGenerator([]string{"polyphony"}, template.Template{
		"command":   "trigger",
		"duration":  value.NewSequence([]any{Nums{125.0, 300.0}, Nums{125.0, 400.0}, Nums{125.0, 500.0}, Nums{250.0, 300.0}, Nums{250.0, 400.0}}),
		"amplitude": value.NewConst[any](1.0),
		"message": template.Template{
			"osc": template.Template{
				"frequency": value.NewSequence([]any{
					notes.Mtofs(60, 48), notes.Mtofs(67, 53), notes.Mtofs(65, 60), notes.Mtofs(64, 48), notes.Mtofs(60, 48),
					notes.Mtofs(67, 53), notes.Mtofs(62, 60), notes.Mtofs(62, 48), notes.Mtofs(64, 48), notes.Mtofs(65, 53),
					notes.Mtofs(69, 60), notes.Mtofs(72, 48),
				}),
				"phase": 0.0,
			},
		},
	}).MsgrNamed("template1"))

	root.AddMessenger(stepper.NewStepper(
		swing.New(40, 2, value.NewSequence([]*swing.Step{
			{}, {}, {}, {},
		})),
		[]string{"template1"},
	))

	kickPlayer.Connect(0, mult, 0)
	hihatPlayer.Connect(0, mult, 0)
	snarePlayer.Connect(0, mult, 0)
	mult.Connect(0, root, 0)
	mult.Connect(0, root, 1)
	poly.Connect(0, allpass, 0)
	poly.Connect(0, root, 0)
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

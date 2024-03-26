package main

import (
	"github.com/almerlucke/genny/float/shape/shapers/emulations/supersaw"
	"github.com/almerlucke/genny/float/shape/shapers/linear"
	"github.com/almerlucke/genny/float/shape/shapers/lookup"
	"github.com/almerlucke/genny/float/shape/shapers/series"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/almerlucke/muse"

	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	"github.com/almerlucke/muse/messengers/lfo"
	adsrctrl "github.com/almerlucke/muse/ui/adsr"
	"github.com/almerlucke/muse/ui/theme"
	"github.com/almerlucke/muse/utils/notes"

	"github.com/almerlucke/muse/value/template"

	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/allpass"
	"github.com/almerlucke/muse/modules/filters/moog"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/phasor"
	"github.com/almerlucke/muse/modules/polyphony"
	"github.com/almerlucke/muse/modules/waveshaper"
)

type TestVoice struct {
	*muse.BasePatch
	ampEnv           *adsr.ADSR
	filterEnv        *adsr.ADSR
	phasor           *phasor.Phasor
	filter           *moog.Moog
	superSaw         *supersaw.SuperSaw
	ampEnvSetting    *adsrc.Setting
	filterEnvSetting *adsrc.Setting
}

func NewTestVoice(ampEnvSetting *adsrc.Setting, filterEnvSetting *adsrc.Setting) *TestVoice {
	testVoice := &TestVoice{
		BasePatch:        muse.NewPatch(0, 1),
		ampEnvSetting:    ampEnvSetting,
		filterEnvSetting: filterEnvSetting,
		superSaw:         supersaw.New(1.5, 0.25, 0.88),
	}

	testVoice.SetSelf(testVoice)

	ampEnv := testVoice.AddModule(adsr.New(ampEnvSetting, adsrc.Duration, 1.0))
	filterEnv := testVoice.AddModule(adsr.New(filterEnvSetting, adsrc.Duration, 1.0))
	multiplier := testVoice.AddModule(functor.NewMult(2))
	filterEnvScaler := testVoice.AddModule(functor.NewScale(5000.0, 100.0))
	osc := testVoice.AddModule(phasor.New(140.0, 0.0))
	filter := testVoice.AddModule(moog.New(1400.0, 0.7, 1.0))
	shape := testVoice.AddModule(waveshaper.New(testVoice.superSaw, 0, nil, nil))

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
	// STUB
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

	if superSawM1, ok := content["superSawM1"].(float64); ok {
		tv.superSaw.SetM1(superSawM1)
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

func main() {
	root := muse.New(2)

	ampEnvSetting := adsrc.NewSetting(1.0, 5.0, 0.2, 5.0, 0.0, 1500.0)
	filterEnvSetting := adsrc.NewSetting(1.0, 5.0, 0.2, 5.0, 0.0, 1500.0)
	ampEnvControl := adsrctrl.NewADSRControl("Amplitude ADSR", ampEnvSetting)
	filterEnvControl := adsrctrl.NewADSRControl("Filter ADSR", filterEnvSetting)

	voices := []polyphony.Voice{}
	for i := 0; i < 20; i++ {
		voice := NewTestVoice(ampEnvSetting, filterEnvSetting)
		voices = append(voices, voice)
	}

	poly := polyphony.New(1, voices).Named("polyphony").AddTo(root)
	allpass := root.AddModule(allpass.New(50, 50, 0.3))

	sineTable := lookup.NewNormalizedSineTable(512)

	targetSuperSaw := lfo.NewTarget("polyphony", series.New(sineTable, linear.New(0.15, 0.1)), "superSawM1", template.Template{
		"command":    "voice",
		"superSawM1": template.NewParameter("superSawM1", nil),
	})

	targetFilter := lfo.NewTarget("polyphony", series.New(sineTable, linear.New(6000.0, 800.0)), "frequency", template.Template{
		"command":         "voice",
		"filterFrequency": template.NewParameter("frequency", nil),
	})

	root.AddMessenger(lfo.NewLFO(0.23, []*lfo.Target{targetSuperSaw}).MsgrNamed("lfo1"))
	root.AddMessenger(lfo.NewLFO(0.13, []*lfo.Target{targetFilter}).MsgrNamed("lfo2"))

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

	keyMap := map[string]float64{}

	baseNote := 36

	keyMap["`"] = notes.Mtof(baseNote + 0)
	keyMap["Z"] = notes.Mtof(baseNote + 1)
	keyMap["X"] = notes.Mtof(baseNote + 2)
	keyMap["C"] = notes.Mtof(baseNote + 3)
	keyMap["V"] = notes.Mtof(baseNote + 4)
	keyMap["B"] = notes.Mtof(baseNote + 5)
	keyMap["N"] = notes.Mtof(baseNote + 6)
	keyMap["M"] = notes.Mtof(baseNote + 7)
	keyMap[","] = notes.Mtof(baseNote + 8)
	keyMap["."] = notes.Mtof(baseNote + 9)
	keyMap["/"] = notes.Mtof(baseNote + 10)

	keyMap["A"] = notes.Mtof(baseNote + 11)
	keyMap["S"] = notes.Mtof(baseNote + 12)
	keyMap["D"] = notes.Mtof(baseNote + 13)
	keyMap["F"] = notes.Mtof(baseNote + 14)
	keyMap["G"] = notes.Mtof(baseNote + 15)
	keyMap["H"] = notes.Mtof(baseNote + 16)
	keyMap["J"] = notes.Mtof(baseNote + 17)
	keyMap["K"] = notes.Mtof(baseNote + 18)
	keyMap["L"] = notes.Mtof(baseNote + 19)
	keyMap[";"] = notes.Mtof(baseNote + 20)
	keyMap["'"] = notes.Mtof(baseNote + 21)
	keyMap["\\"] = notes.Mtof(baseNote + 22)

	keyMap["Q"] = notes.Mtof(baseNote + 23)
	keyMap["W"] = notes.Mtof(baseNote + 24)
	keyMap["E"] = notes.Mtof(baseNote + 25)
	keyMap["R"] = notes.Mtof(baseNote + 26)
	keyMap["T"] = notes.Mtof(baseNote + 27)
	keyMap["Y"] = notes.Mtof(baseNote + 28)
	keyMap["U"] = notes.Mtof(baseNote + 29)
	keyMap["I"] = notes.Mtof(baseNote + 30)
	keyMap["O"] = notes.Mtof(baseNote + 31)
	keyMap["P"] = notes.Mtof(baseNote + 32)
	keyMap["["] = notes.Mtof(baseNote + 33)
	keyMap["]"] = notes.Mtof(baseNote + 34)

	keyMap["1"] = notes.Mtof(baseNote + 35)
	keyMap["2"] = notes.Mtof(baseNote + 36)
	keyMap["3"] = notes.Mtof(baseNote + 37)
	keyMap["4"] = notes.Mtof(baseNote + 38)
	keyMap["5"] = notes.Mtof(baseNote + 39)
	keyMap["6"] = notes.Mtof(baseNote + 40)
	keyMap["7"] = notes.Mtof(baseNote + 41)
	keyMap["8"] = notes.Mtof(baseNote + 42)
	keyMap["9"] = notes.Mtof(baseNote + 43)
	keyMap["0"] = notes.Mtof(baseNote + 44)
	keyMap["-"] = notes.Mtof(baseNote + 45)
	keyMap["="] = notes.Mtof(baseNote + 46)

	if deskCanvas, ok := w.Canvas().(desktop.Canvas); ok {
		deskCanvas.SetOnKeyDown(func(k *fyne.KeyEvent) {
			if f, ok := keyMap[string(k.Name)]; ok {
				poly.ReceiveMessage(map[string]any{
					"command":   "trigger",
					"noteOn":    string(k.Name),
					"amplitude": 1.0,
					"message": map[string]any{
						"osc": map[string]any{
							"frequency": f,
						},
					},
				})
			}
		})

		deskCanvas.SetOnKeyUp(func(k *fyne.KeyEvent) {
			if _, ok := keyMap[string(k.Name)]; ok {
				poly.ReceiveMessage(map[string]any{
					"command": "trigger",
					"noteOff": string(k.Name),
				})
			}
		})
	}

	w.SetContent(
		container.NewVBox(
			container.NewHBox(
				widget.NewButton("Start", func() {
					root.StartAudio()
				}),
				widget.NewButton("Stop", func() {
					root.StopAudio()
				}),
				widget.NewButton("Notes Off", func() {
					poly.(*polyphony.Polyphony).AllNotesOff()
				}),
			),
			container.NewHBox(
				ampEnvControl.UI(),
				filterEnvControl.UI(),
			),
		),
	)

	w.ShowAndRun()
}

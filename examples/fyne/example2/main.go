package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/gordonklaus/portaudio"

	"github.com/almerlucke/muse"

	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	shaping "github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/messengers/lfo"
	adsrctrl "github.com/almerlucke/muse/ui/controls/adsr"
	museTheme "github.com/almerlucke/muse/ui/theme"
	"github.com/almerlucke/muse/utils"

	"github.com/almerlucke/muse/values/template"

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
	ampEnv             *adsr.ADSR
	filterEnv          *adsr.ADSR
	phasor             *phasor.Phasor
	filter             *moog.Moog
	superSaw           *shaping.Chain
	ampStepProvider    adsrctrl.ADSRStepProvider
	filterStepProvider adsrctrl.ADSRStepProvider
}

func NewTestVoice(config *muse.Configuration, ampStepProvider adsrctrl.ADSRStepProvider, filterStepProvider adsrctrl.ADSRStepProvider) *TestVoice {
	testVoice := &TestVoice{
		BasePatch:          muse.NewPatch(0, 1, config, ""),
		ampStepProvider:    ampStepProvider,
		filterStepProvider: filterStepProvider,
		superSaw:           shaping.NewSuperSaw(),
	}

	ampEnv := testVoice.AddModule(adsr.NewADSR(ampStepProvider.ADSRSteps(), adsrc.Absolute, adsrc.Duration, 1.0, config, "ampAdsr"))
	filterEnv := testVoice.AddModule(adsr.NewADSR(filterStepProvider.ADSRSteps(), adsrc.Absolute, adsrc.Duration, 1.0, config, "filterAdsr"))
	multiplier := testVoice.AddModule(functor.NewFunctor(2, functor.FunctorMult, config, ""))
	filterEnvScaler := testVoice.AddModule(functor.NewFunctor(1, func(in []float64) float64 { return in[0]*5000.0 + 100.0 }, config, ""))
	osc := testVoice.AddModule(phasor.NewPhasor(140.0, 0.0, config, "osc"))
	filter := testVoice.AddModule(moog.NewMoog(1400.0, 0.7, 1.0, config, "filter"))
	shape := testVoice.AddModule(waveshaper.NewWaveShaper(testVoice.superSaw, 0, nil, nil, config, "shaper"))

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
	// STUB
}

func (tv *TestVoice) NoteOn(amplitude float64, message any, config *muse.Configuration) {
	msg := message.(map[string]any)

	tv.ampEnv.TriggerFull(0, amplitude, tv.ampStepProvider.ADSRSteps(), adsrc.Absolute, adsrc.NoteOff)
	tv.filterEnv.TriggerFull(0, 1.0, tv.filterStepProvider.ADSRSteps(), adsrc.Absolute, adsrc.NoteOff)
	tv.phasor.ReceiveMessage(msg["osc"])
}

func (tv *TestVoice) NoteOff() {
	tv.ampEnv.Release()
	tv.filterEnv.Release()
}

func (tv *TestVoice) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	if superSawM1, ok := content["superSawM1"].(float64); ok {
		tv.superSaw.SetSuperSawM1(superSawM1)
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
	env := muse.NewEnvironment(2, 44100, 512)

	ampEnvControl := adsrctrl.NewADSRControl("Amplitude ADSR")
	filterEnvControl := adsrctrl.NewADSRControl("Filter ADSR")

	voices := []polyphony.Voice{}
	for i := 0; i < 20; i++ {
		voice := NewTestVoice(env.Config, ampEnvControl, filterEnvControl)
		voices = append(voices, voice)
	}

	poly := env.AddModule(polyphony.NewPolyphony(1, voices, env.Config, "polyphony"))
	allpass := env.AddModule(allpass.NewAllpass(50, 50, 0.3, env.Config, "allpass"))

	sineTable := shaping.NewNormalizedSineTable(512)

	targetSuperSaw := lfo.NewTarget("polyphony", shaping.NewChain(sineTable, shaping.NewLinear(0.15, 0.1)), "superSawM1", template.Template{
		"command":    "voice",
		"superSawM1": template.NewParameter("superSawM1", nil),
	})

	targetFilter := lfo.NewTarget("polyphony", shaping.NewChain(sineTable, shaping.NewLinear(6000.0, 800.0)), "frequency", template.Template{
		"command":         "voice",
		"filterFrequency": template.NewParameter("frequency", nil),
	})

	env.AddMessenger(lfo.NewLFO(0.23, []*lfo.Target{targetSuperSaw}, env.Config, "lfo1"))
	env.AddMessenger(lfo.NewLFO(0.13, []*lfo.Target{targetFilter}, env.Config, "lfo2"))

	muse.Connect(poly, 0, allpass, 0)
	muse.Connect(poly, 0, env, 0)
	muse.Connect(allpass, 0, env, 1)

	portaudio.Initialize()
	defer portaudio.Terminate()

	stream, err := env.PortaudioStream()
	if err != nil {
		log.Fatalf("error opening portaudio stream, %v", err)
	}

	defer stream.Close()

	a := app.New()

	// Theme

	appTheme := &museTheme.Theme{}

	// theme.LightTheme() appTheme

	a.Settings().SetTheme(appTheme)

	w := a.NewWindow("Muse")

	w.Resize(fyne.Size{
		Width:  700,
		Height: 400,
	})

	keyMap := map[string]float64{}

	baseNote := 36

	keyMap["`"] = utils.Mtof(baseNote + 0)
	keyMap["Z"] = utils.Mtof(baseNote + 1)
	keyMap["X"] = utils.Mtof(baseNote + 2)
	keyMap["C"] = utils.Mtof(baseNote + 3)
	keyMap["V"] = utils.Mtof(baseNote + 4)
	keyMap["B"] = utils.Mtof(baseNote + 5)
	keyMap["N"] = utils.Mtof(baseNote + 6)
	keyMap["M"] = utils.Mtof(baseNote + 7)
	keyMap[","] = utils.Mtof(baseNote + 8)
	keyMap["."] = utils.Mtof(baseNote + 9)
	keyMap["/"] = utils.Mtof(baseNote + 10)

	keyMap["A"] = utils.Mtof(baseNote + 11)
	keyMap["S"] = utils.Mtof(baseNote + 12)
	keyMap["D"] = utils.Mtof(baseNote + 13)
	keyMap["F"] = utils.Mtof(baseNote + 14)
	keyMap["G"] = utils.Mtof(baseNote + 15)
	keyMap["H"] = utils.Mtof(baseNote + 16)
	keyMap["J"] = utils.Mtof(baseNote + 17)
	keyMap["K"] = utils.Mtof(baseNote + 18)
	keyMap["L"] = utils.Mtof(baseNote + 19)
	keyMap[";"] = utils.Mtof(baseNote + 20)
	keyMap["'"] = utils.Mtof(baseNote + 21)
	keyMap["\\"] = utils.Mtof(baseNote + 22)

	keyMap["Q"] = utils.Mtof(baseNote + 23)
	keyMap["W"] = utils.Mtof(baseNote + 24)
	keyMap["E"] = utils.Mtof(baseNote + 25)
	keyMap["R"] = utils.Mtof(baseNote + 26)
	keyMap["T"] = utils.Mtof(baseNote + 27)
	keyMap["Y"] = utils.Mtof(baseNote + 28)
	keyMap["U"] = utils.Mtof(baseNote + 29)
	keyMap["I"] = utils.Mtof(baseNote + 30)
	keyMap["O"] = utils.Mtof(baseNote + 31)
	keyMap["P"] = utils.Mtof(baseNote + 32)
	keyMap["["] = utils.Mtof(baseNote + 33)
	keyMap["]"] = utils.Mtof(baseNote + 34)

	keyMap["1"] = utils.Mtof(baseNote + 35)
	keyMap["2"] = utils.Mtof(baseNote + 36)
	keyMap["3"] = utils.Mtof(baseNote + 37)
	keyMap["4"] = utils.Mtof(baseNote + 38)
	keyMap["5"] = utils.Mtof(baseNote + 39)
	keyMap["6"] = utils.Mtof(baseNote + 40)
	keyMap["7"] = utils.Mtof(baseNote + 41)
	keyMap["8"] = utils.Mtof(baseNote + 42)
	keyMap["9"] = utils.Mtof(baseNote + 43)
	keyMap["0"] = utils.Mtof(baseNote + 44)
	keyMap["-"] = utils.Mtof(baseNote + 45)
	keyMap["="] = utils.Mtof(baseNote + 46)

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
					stream.Start()
				}),
				widget.NewButton("Stop", func() {
					stream.Stop()
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

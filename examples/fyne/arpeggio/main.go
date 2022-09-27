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
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	adsrctrl "github.com/almerlucke/muse/ui/controls/adsr"
	"github.com/almerlucke/muse/ui/theme"
	"github.com/almerlucke/muse/value"

	"github.com/almerlucke/muse/value/arpeggio"
	"github.com/almerlucke/muse/value/template"

	"github.com/almerlucke/muse/utils/notes"

	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/filters/moog"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/phasor"
	"github.com/almerlucke/muse/modules/polyphony"
	"github.com/almerlucke/muse/modules/waveshaper"
)

type TestVoice struct {
	*muse.BasePatch
	ampEnv          *adsr.ADSR
	filterEnv       *adsr.ADSR
	phasor          *phasor.Phasor
	filter          *moog.Moog
	shaper          shaping.Shaper
	ampStepProvider adsrctrl.ADSRStepProvider
}

func NewTestVoice(config *muse.Configuration, ampStepProvider adsrctrl.ADSRStepProvider) *TestVoice {
	testVoice := &TestVoice{
		BasePatch:       muse.NewPatch(0, 1, config, ""),
		ampStepProvider: ampStepProvider,
		shaper:          shaping.NewSineTable(512),
	}

	ampEnv := testVoice.AddModule(adsr.NewADSR(ampStepProvider.ADSRSteps(), adsrc.Absolute, adsrc.Duration, 1.0, config, "ampAdsr"))
	multiplier := testVoice.AddModule(functor.NewFunctor(2, functor.FunctorMult, config, ""))
	osc := testVoice.AddModule(phasor.NewPhasor(140.0, 0.0, config, "osc"))
	shape := testVoice.AddModule(waveshaper.NewWaveShaper(testVoice.shaper, 0, nil, nil, config, "shaper"))

	muse.Connect(osc, 0, shape, 0)
	muse.Connect(shape, 0, multiplier, 0)
	muse.Connect(ampEnv, 0, multiplier, 1)
	muse.Connect(multiplier, 0, testVoice, 0)

	testVoice.ampEnv = ampEnv.(*adsr.ADSR)
	testVoice.phasor = osc.(*phasor.Phasor)

	return testVoice
}

func (tv *TestVoice) IsActive() bool {
	return tv.ampEnv.IsActive()
}

func (tv *TestVoice) Note(duration float64, amplitude float64, message any, config *muse.Configuration) {
	msg := message.(map[string]any)

	tv.ampEnv.TriggerFull(duration, amplitude, tv.ampStepProvider.ADSRSteps(), adsrc.Absolute, adsrc.Duration)
	tv.phasor.ReceiveMessage(msg["osc"])
}

func (tv *TestVoice) NoteOn(amplitude float64, message any, config *muse.Configuration) {
	// STUB
}

func (tv *TestVoice) NoteOff() {
	tv.ampEnv.Release()
	tv.filterEnv.Release()
}

func (tv *TestVoice) ReceiveMessage(msg any) []*muse.Message {
	// content := msg.(map[string]any)
	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	env := muse.NewEnvironment(1, 44100, 1024)

	ampEnvControl := adsrctrl.NewADSRControl("Amplitude ADSR")

	voices := []polyphony.Voice{}
	for i := 0; i < 20; i++ {
		voice := NewTestVoice(env.Config, ampEnvControl)
		voices = append(voices, voice)
	}

	bpm := 100.0

	// milliPerBeat := 60000.0 / bpm

	poly := env.AddModule(polyphony.NewPolyphony(1, voices, env.Config, "polyphony"))

	octave := notes.O4

	env.AddMessenger(banger.NewTemplateGenerator([]string{"polyphony"}, template.Template{
		"command":   "trigger",
		"duration":  value.NewSequence([]any{375.0}),
		"amplitude": value.NewConst[any](0.5),
		"message": template.Template{
			"osc": template.Template{
				"frequency": value.NewAnd(
					[]value.Valuer[any]{
						value.NewRepeat[any](arpeggio.NewArpeggiatorNC(notes.GMajor.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 2, 2),
						value.NewRepeat[any](arpeggio.NewArpeggiatorNC(notes.DMajor.FreqAny(octave), arpeggio.Alternate, arpeggio.Inclusive, false), 2, 2),
						value.NewRepeat[any](arpeggio.NewArpeggiatorNC(notes.CMajor.FreqAny(octave), arpeggio.Converge, arpeggio.None, false), 1, 1),
						value.NewRepeat[any](arpeggio.NewArpeggiatorNC(notes.GMajor.FreqAny(octave), arpeggio.Converge, arpeggio.None, true), 2, 2),
						value.NewRepeat[any](arpeggio.NewArpeggiatorNC(notes.AMinor.FreqAny(octave), arpeggio.Random, arpeggio.Exclusive, false), 1, 1),
						value.NewRepeat[any](arpeggio.NewArpeggiatorNC(notes.DMajor7.FreqAny(octave), arpeggio.Up, arpeggio.Inclusive, true), 2, 2),
						value.NewRepeat[any](arpeggio.NewArpeggiatorNC(notes.GMajor.FreqAny(octave), arpeggio.Alternate, arpeggio.Exclusive, false), 1, 1),
						value.NewRepeat[any](arpeggio.NewArpeggiatorNC(notes.DMajor.FreqAny(octave), arpeggio.Converge, arpeggio.Inclusive, true), 2, 2),

						// // Row 1
						// notes.GMajor.Freq(octave), notes.DMajor.Freq(octave), notes.CMajor.Freq(octave), notes.GMajor.Freq(octave),
						// notes.AMinor.Freq(octave), notes.DMajor7.Freq(octave), notes.GMajor.Freq(octave), notes.DMajor.Freq(octave),
						// // Row 2
						// notes.GMajor.Freq(octave), notes.AMajor7.Freq(octave), notes.DMajor.Freq(octave), notes.EMinor.Freq(octave),
						// notes.AMajor7.Freq(octave), notes.DMajor.Freq(octave),
						// // Row 3
						// notes.GMajor.Freq(octave), notes.DMajor.Freq(octave), notes.CMajor.Freq(octave), notes.GMajor.Freq(octave),
						// notes.AMinor.Freq(octave), notes.DMajor7.Freq(octave), notes.GMajor.Freq(octave), notes.DMajor.Freq(octave),
						// // Row 4
						// notes.GMajor.Freq(octave), notes.AMajor7.Freq(octave), notes.DMajor.Freq(octave), notes.BMinor.Freq(octave),
						// notes.AMinor.Freq(octave), notes.DMajor7.Freq(octave), notes.GMajor.Freq(octave),
						// // Row 5
						// notes.CMajor.Freq(octave), notes.DMajor7.Freq(octave), notes.GMajor.Freq(octave), notes.EMinor.Freq(octave),
						// notes.CMajor.Freq(octave), notes.DMajor7.Freq(octave), notes.GMajor.Freq(octave),
					}, true),
			},
		},
	}, "template"))

	env.AddMessenger(stepper.NewStepper(
		swing.New(value.NewConst(bpm), value.NewConst(2.0), value.NewSequence(
			[]*swing.Step{
				{},
			},
		)),
		[]string{"template"}, "",
	))

	muse.Connect(poly, 0, env, 0)

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

	w.SetContent(
		container.NewVBox(
			container.NewHBox(
				widget.NewButton("Start", func() {
					stream.Start()
				}),
				widget.NewButton("Stop", func() {
					stream.Stop()
				}),
			),
			container.NewHBox(
				ampEnvControl.UI(),
			),
		),
	)

	w.ShowAndRun()
}
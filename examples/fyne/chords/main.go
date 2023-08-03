package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/almerlucke/muse"

	adsrc "github.com/almerlucke/muse/components/envelopes/adsr"
	shaping "github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/modules"
	adsrctrl "github.com/almerlucke/muse/ui/adsr"
	"github.com/almerlucke/muse/ui/theme"
	"github.com/almerlucke/muse/value"

	"github.com/almerlucke/muse/value/template"

	"github.com/almerlucke/muse/utils/notes"

	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/filters/moog"
	"github.com/almerlucke/muse/modules/phasor"
	"github.com/almerlucke/muse/modules/polyphony"
	"github.com/almerlucke/muse/modules/waveshaper"
)

type TestVoice struct {
	*muse.BasePatch
	ampEnv        *adsr.ADSR
	filterEnv     *adsr.ADSR
	phasor        *phasor.Phasor
	filter        *moog.Moog
	shaper        shaping.Shaper
	ampEnvSetting *adsrc.Setting
}

func NewTestVoice(ampEnvSetting *adsrc.Setting) *TestVoice {
	testVoice := &TestVoice{
		BasePatch:     muse.NewPatch(0, 1),
		ampEnvSetting: ampEnvSetting,
		shaper:        shaping.NewSineTable(512),
	}

	testVoice.SetSelf(testVoice)

	ampEnv := adsr.New(ampEnvSetting, adsrc.Duration, 1.0).AddTo(testVoice)
	osc := phasor.New(140.0, 0.0).AddTo(testVoice)
	shape := waveshaper.New(testVoice.shaper, 0, nil, nil).AddTo(testVoice).In(osc)
	mult := modules.Mult(shape, ampEnv).AddTo(testVoice)

	testVoice.In(mult)

	testVoice.ampEnv = ampEnv.(*adsr.ADSR)
	testVoice.phasor = osc.(*phasor.Phasor)

	return testVoice
}

func (tv *TestVoice) IsActive() bool {
	return tv.ampEnv.IsActive()
}

func (tv *TestVoice) Note(duration float64, amplitude float64, message any, config *muse.Configuration) {
	msg := message.(map[string]any)

	tv.ampEnv.TriggerFull(duration, amplitude, tv.ampEnvSetting, adsrc.Duration)
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
	root := muse.New(1)

	ampEnvSetting := adsrc.NewSetting(1.0, 5.0, 0.3, 5.0, 0.0, 1300.0)
	ampEnvControl := adsrctrl.NewADSRControl("Amplitude ADSR", ampEnvSetting)

	voices := []polyphony.Voice{}
	for i := 0; i < 20; i++ {
		voice := NewTestVoice(ampEnvSetting)
		voices = append(voices, voice)
	}

	bpm := 200

	// milliPerBeat := 60000.0 / bpm

	poly := polyphony.New(1, voices).Named("polyphony").AddTo(root)

	octave := notes.O3

	root.AddMessenger(banger.NewTemplateGenerator([]string{"polyphony"}, template.Template{
		"command": "trigger",
		"duration": value.NewSequence([]any{
			750.0, 750.0, 750.0, 750.0, 375.0, 375.0, 750.0, 750.0,
			750.0, 750.0, 750.0, 750.0, 750.0, 750.0, 750.0,
			750.0, 750.0, 750.0, 750.0, 375.0, 375.0, 750.0, 750.0,
			750.0, 750.0, 750.0, 750.0, 750.0, 750.0, 750.0,
			750.0, 750.0, 750.0, 750.0, 750.0, 750.0, 750.0,
		}),
		"amplitude": value.NewConst[any](1.0),
		"message": template.Template{
			"osc": template.Template{
				"frequency": value.NewSequence([]any{
					// Row 1
					notes.GMajor.Freq(octave), notes.DMajor.Freq(octave), notes.CMajor.Freq(octave), notes.GMajor.Freq(octave),
					notes.AMinor.Freq(octave), notes.DMajor7.Freq(octave), notes.GMajor.Freq(octave), notes.DMajor.Freq(octave),
					// Row 2
					notes.GMajor.Freq(octave), notes.AMajor7.Freq(octave), notes.DMajor.Freq(octave), notes.EMinor.Freq(octave),
					notes.AMajor7.Freq(octave), notes.DMajor.Freq(octave),
					// Row 3
					notes.GMajor.Freq(octave), notes.DMajor.Freq(octave), notes.CMajor.Freq(octave), notes.GMajor.Freq(octave),
					notes.AMinor.Freq(octave), notes.DMajor7.Freq(octave), notes.GMajor.Freq(octave), notes.DMajor.Freq(octave),
					// Row 4
					notes.GMajor.Freq(octave), notes.AMajor7.Freq(octave), notes.DMajor.Freq(octave), notes.BMinor.Freq(octave),
					notes.AMinor.Freq(octave), notes.DMajor7.Freq(octave), notes.GMajor.Freq(octave),
					// Row 5
					notes.CMajor.Freq(octave), notes.DMajor7.Freq(octave), notes.GMajor.Freq(octave), notes.EMinor.Freq(octave),
					notes.CMajor.Freq(octave), notes.DMajor7.Freq(octave), notes.GMajor.Freq(octave),
				}),
			},
		},
	}).MsgrNamed("template"))

	root.AddMessenger(stepper.NewStepper(
		swing.New(bpm, 1, value.NewSequence(
			[]*swing.Step{
				// Row 1
				// G              D                 C                 G                 Am  D7  G                 D
				{}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {}, {}, {Skip: true}, {}, {Skip: true},
				// Row 2
				// G              A7                 D                                  Em                A7                D
				{}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true},
				// Row 3
				// G              D                 C                 G                 Am  D7  G                 D
				{}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {}, {}, {Skip: true}, {}, {Skip: true},
				// Row 4
				// G              A7                D                 BM                Am                D7                G
				{}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true},
				// Row 5
				// C              D7                G                 Em                 C                 D7               G
				{}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true}, {}, {Skip: true},
			},
		)),
		[]string{"template"},
	))

	poly.Connect(0, root, 0)

	err := root.InitializeAudio()
	if err != nil {
		log.Fatalf("error opening audio stream, %v", err)
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
					root.StartAudio()
				}),
				widget.NewButton("Stop", func() {
					root.StopAudio()
				}),
			),
			container.NewHBox(
				ampEnvControl.UI(),
			),
		),
	)

	w.ShowAndRun()
}

package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/almerlucke/genny/and"
	"github.com/almerlucke/genny/arpeggio"
	"github.com/almerlucke/genny/constant"
	adsrc "github.com/almerlucke/genny/float/envelopes/adsr"
	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/float/shape/shapers/emulations/jp8000"
	"github.com/almerlucke/genny/repeat"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/template"
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/modules/adsr"
	"github.com/almerlucke/muse/modules/effects/chorus"
	"github.com/almerlucke/muse/modules/filters/moog"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/phasor"
	"github.com/almerlucke/muse/modules/polyphony"
	"github.com/almerlucke/muse/modules/waveshaper"
	adsrctrl "github.com/almerlucke/muse/ui/adsr"
	"github.com/almerlucke/muse/ui/theme"
	"github.com/almerlucke/muse/utils/notes"
	"log"
)

type TestVoice struct {
	*muse.BasePatch
	ampEnv        *adsr.ADSR
	filterEnv     *adsr.ADSR
	phasor        *phasor.Phasor
	filter        *moog.Moog
	shaper        shape.Shaper
	ampEnvSetting *adsrc.Setting
}

func NewTestVoice(ampEnvSetting *adsrc.Setting) *TestVoice {
	testVoice := &TestVoice{
		BasePatch:     muse.NewPatch(0, 1),
		ampEnvSetting: ampEnvSetting,
		shaper:        jp8000.NewTriMod(0.3),
	}

	ampEnv := testVoice.AddModule(adsr.New(ampEnvSetting, adsrc.Duration, 1.0))
	multiplier := testVoice.AddModule(functor.NewMult(2))
	osc := testVoice.AddModule(phasor.New(140.0, 0.0))
	shaper := testVoice.AddModule(waveshaper.New(testVoice.shaper, 0, nil, nil))

	osc.Connect(0, shaper, 0)
	shaper.Connect(0, multiplier, 0)
	ampEnv.Connect(0, multiplier, 1)
	multiplier.Connect(0, testVoice, 0)

	testVoice.ampEnv = ampEnv.(*adsr.ADSR)
	testVoice.phasor = osc.(*phasor.Phasor)

	return testVoice
}

func (tv *TestVoice) IsActive() bool {
	return tv.ampEnv.IsActive()
}

func (tv *TestVoice) Note(duration float64, amplitude float64, message any, config *muse.Configuration) {
	msg := message.(map[string]any)

	tv.ampEnv.Clear()
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
	root := muse.New(2)

	ampEnvSetting := adsrc.NewSetting(1.0, 5.0, 0.2, 5.0, 0.0, 1300.0)
	ampEnvControl := adsrctrl.NewADSRControl("Amplitude ADSR", ampEnvSetting)

	voices1 := []polyphony.Voice{}
	for i := 0; i < 20; i++ {
		var voice polyphony.Voice

		voice = NewTestVoice(ampEnvSetting)

		voices1 = append(voices1, voice)
	}

	bpm := 100

	// milliPerBeat := 60000.0 / bpm

	poly1 := polyphony.New(1, voices1).Named("polyphony1").AddTo(root)
	chor := chorus.New(0.24, 0.4, 0.45, 0.2, 1.0, 0.5, nil).AddTo(root)

	octave := notes.O4

	root.AddMessenger(banger.NewTemplateBang([]string{"polyphony1"}, template.Template{
		"command":   "trigger",
		"duration":  constant.New(375.0),
		"amplitude": constant.New(0.7),
		"message": template.Template{
			"osc": template.Template{
				"frequency": and.NewLoop[float64](
					// Row 1
					repeat.NewRand[float64](arpeggio.New(notes.GMajor.Freq(octave), arpeggio.Up, arpeggio.Exclusive, false), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.DMajor.Freq(octave), arpeggio.Alternate, arpeggio.None, false), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.CMajor.Freq(octave), arpeggio.Converge, arpeggio.None, false), 1, 3),
					repeat.NewRand[float64](arpeggio.New(notes.GMajor.Freq(octave), arpeggio.Converge, arpeggio.None, true), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.AMinor.Freq(octave), arpeggio.Random, arpeggio.Exclusive, false), 1, 3),
					repeat.NewRand[float64](arpeggio.New(notes.DMajor7.Freq(octave), arpeggio.Up, arpeggio.Inclusive, true), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.GMajor.Freq(octave), arpeggio.Alternate, arpeggio.None, false), 1, 3),
					repeat.NewRand[float64](arpeggio.New(notes.DMajor.Freq(octave), arpeggio.Converge, arpeggio.Inclusive, true), 2, 4),
					// Row 2
					repeat.NewRand[float64](arpeggio.New(notes.GMajor.Freq(octave), arpeggio.Up, arpeggio.Inclusive, false), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.AMajor7.Freq(octave), arpeggio.Alternate, arpeggio.None, false), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.DMajor.Freq(octave), arpeggio.Converge, arpeggio.Exclusive, true), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.EMinor.Freq(octave), arpeggio.Converge, arpeggio.Inclusive, false), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.AMajor7.Freq(octave), arpeggio.Up, arpeggio.Inclusive, true), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.DMajor.Freq(octave), arpeggio.Random, arpeggio.Exclusive, false), 2, 4),
					// Row 3
					repeat.NewRand[float64](arpeggio.New(notes.GMajor.Freq(octave), arpeggio.Alternate, arpeggio.None, false), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.DMajor.Freq(octave), arpeggio.Up, arpeggio.Inclusive, false), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.CMajor.Freq(octave), arpeggio.Converge, arpeggio.Inclusive, false), 1, 3),
					repeat.NewRand[float64](arpeggio.New(notes.GMajor.Freq(octave), arpeggio.Converge, arpeggio.Exclusive, true), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.AMinor.Freq(octave), arpeggio.Random, arpeggio.Exclusive, false), 1, 3),
					repeat.NewRand[float64](arpeggio.New(notes.DMajor7.Freq(octave), arpeggio.Alternate, arpeggio.Inclusive, true), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.GMajor.Freq(octave), arpeggio.Up, arpeggio.None, false), 1, 3),
					repeat.NewRand[float64](arpeggio.New(notes.DMajor.Freq(octave), arpeggio.Converge, arpeggio.Inclusive, true), 2, 4),
					// Row 4
					repeat.NewRand[float64](arpeggio.New(notes.GMajor.Freq(octave), arpeggio.Up, arpeggio.Inclusive, false), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.AMajor7.Freq(octave), arpeggio.Alternate, arpeggio.None, false), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.DMajor.Freq(octave), arpeggio.Converge, arpeggio.Exclusive, true), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.BMinor.Freq(octave), arpeggio.Converge, arpeggio.Inclusive, false), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.AMinor.Freq(octave), arpeggio.Up, arpeggio.Inclusive, true), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.DMajor7.Freq(octave), arpeggio.Random, arpeggio.Exclusive, false), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.GMajor.Freq(octave), arpeggio.Up, arpeggio.Inclusive, false), 2, 4),
					// Row 5
					repeat.NewRand[float64](arpeggio.New(notes.CMajor.Freq(octave), arpeggio.Up, arpeggio.Exclusive, false), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.DMajor7.Freq(octave), arpeggio.Alternate, arpeggio.None, false), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.GMajor.Freq(octave), arpeggio.Converge, arpeggio.None, false), 1, 3),
					repeat.NewRand[float64](arpeggio.New(notes.EMinor.Freq(octave), arpeggio.Converge, arpeggio.None, true), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.CMajor.Freq(octave), arpeggio.Random, arpeggio.Exclusive, false), 1, 3),
					repeat.NewRand[float64](arpeggio.New(notes.DMajor7.Freq(octave), arpeggio.Up, arpeggio.Inclusive, true), 2, 4),
					repeat.NewRand[float64](arpeggio.New(notes.GMajor.Freq(octave), arpeggio.Alternate, arpeggio.None, false), 1, 3),
				),
			},
		},
	}).MsgrNamed("template1"))

	root.AddMessenger(stepper.NewStepper(
		swing.New(bpm, 2, sequence.NewLoop(
			[]*swing.Step{
				{}, {}, {}, {}, {}, {}, {}, {Shuffle: 0.1}, {}, {}, {}, {}, {}, {}, {}, {Shuffle: 0.1, ShuffleRand: 0.05},
			}...,
		)),
		[]string{"template1"},
	))

	poly1.Connect(0, chor, 0)
	chor.Connect(0, root, 0)
	chor.Connect(1, root, 1)

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
					_ = root.StartAudio()
				}),
				widget.NewButton("Stop", func() {
					_ = root.StopAudio()
				}),
			),
			container.NewHBox(
				ampEnvControl.UI(),
			),
		),
	)

	w.ShowAndRun()
}

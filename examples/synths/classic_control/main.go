package main

import (
	"log"
	"math/rand"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/envelopes/adsr"
	"github.com/almerlucke/muse/control"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/modules/allpass"
	"github.com/almerlucke/muse/modules/chorus"
	"github.com/almerlucke/muse/modules/functor"
	"github.com/almerlucke/muse/modules/polyphony"
	"github.com/almerlucke/muse/synths/classic"
	"github.com/almerlucke/muse/ui/theme"
	"github.com/almerlucke/muse/utils/notes"
	"github.com/almerlucke/muse/value"
	"github.com/almerlucke/muse/value/arpeggio"
	"github.com/almerlucke/muse/value/template"
	"github.com/gordonklaus/portaudio"
)

type ClassicSynth struct {
	*muse.BasePatch
	*control.Collection
	ampEnv    *adsr.BasicStepProvider
	filterEnv *adsr.BasicStepProvider
	poly      *polyphony.Polyphony
	chorus1   *chorus.Chorus
	chorus2   *chorus.Chorus
}

func NewClassicSynth(bpm float64, config *muse.Configuration) *ClassicSynth {
	synth := &ClassicSynth{
		BasePatch:  muse.NewPatch(0, 2, config, "synth"),
		Collection: control.NewCollection(),
	}

	synth.AddReceiver(synth, "synth")

	ampEnv := adsr.NewBasicStepProvider()
	ampEnv.Steps[0] = adsr.Step{Level: 1.0, Duration: 25.0}
	ampEnv.Steps[1] = adsr.Step{Level: 0.3, Duration: 80.0}
	ampEnv.Steps[3] = adsr.Step{Duration: 2000.0}

	filterEnv := adsr.NewBasicStepProvider()
	filterEnv.Steps[0] = adsr.Step{Level: 0.9, Duration: 25.0}
	filterEnv.Steps[1] = adsr.Step{Level: 0.5, Duration: 80.0}
	filterEnv.Steps[3] = adsr.Step{Duration: 2000.0}

	synth.ampEnv = ampEnv
	synth.filterEnv = filterEnv
	synth.poly = classic.NewSynth(20, ampEnv, filterEnv, config, "poly")
	synth.chorus1 = chorus.NewChorus(false, 15, 10, 0.3, 1.42, 0.5, nil, config, "chorus1")
	synth.chorus2 = chorus.NewChorus(false, 15, 10, 0.31, 1.43, 0.55, nil, config, "chorus2")

	synth.AddModule(synth.poly)
	synth.AddModule(synth.chorus1)
	synth.AddModule(synth.chorus2)

	synthAmp1 := synth.AddModule(functor.NewFunctor(1, func(v []float64) float64 { return v[0] * 0.85 }, config))
	synthAmp2 := synth.AddModule(functor.NewFunctor(1, func(v []float64) float64 { return v[0] * 0.85 }, config))
	allpass1 := synth.AddModule(allpass.NewAllpass(2500.0, 60000/bpm*1.666, 0.5, config, "allpass"))
	allpass2 := synth.AddModule(allpass.NewAllpass(2500.0, 60000/bpm*1.75, 0.4, config, "allpass"))
	allpassAmp1 := synth.AddModule(functor.NewFunctor(1, func(v []float64) float64 { return v[0] * 0.5 }, config))
	allpassAmp2 := synth.AddModule(functor.NewFunctor(1, func(v []float64) float64 { return v[0] * 0.5 }, config))

	muse.Connect(synth.poly, 0, synthAmp1, 0)
	muse.Connect(synth.poly, 1, synthAmp2, 0)
	muse.Connect(synthAmp1, 0, synth.chorus1, 0)
	muse.Connect(synthAmp2, 0, synth.chorus2, 0)
	muse.Connect(synthAmp1, 0, allpass1, 0)
	muse.Connect(synthAmp2, 0, allpass2, 0)
	muse.Connect(allpass1, 0, allpassAmp1, 0)
	muse.Connect(allpass2, 0, allpassAmp2, 0)
	muse.Connect(allpassAmp1, 0, synth.chorus1, 0)
	muse.Connect(allpassAmp2, 0, synth.chorus2, 0)
	muse.Connect(synth.chorus1, 0, synth, 0)
	muse.Connect(synth.chorus2, 0, synth, 1)

	synth.AddControl(control.NewBaseFloatControl("voice.filterFcMin", "Filter", "Filter Frequency Min", 50.0, 8000.0, 1.0, 50.0))
	synth.AddControl(control.NewBaseFloatControl("voice.filterFcMax", "Filter", "Filter Frequency Max", 50.0, 8000.0, 1.0, 8000.0))
	synth.AddControl(control.NewBaseFloatControl("voice.filterResonance", "Filter", "Resonance", 0.0, 1.0, 0.01, 0.7))
	synth.AddControl(control.NewBaseFloatControl("voice.osc1Mix", "Mixer", "Osc1 Mix", 0.0, 1.0, 0.01, 0.6))
	synth.AddControl(control.NewBaseFloatControl("voice.osc2Mix", "Mixer", "Osc2 Mix", 0.0, 1.0, 0.01, 0.35))
	synth.AddControl(control.NewBaseFloatControl("voice.noiseMix", "Mixer", "Noise Mix", 0.0, 1.0, 0.01, 0.05))
	synth.AddControl(control.NewBaseFloatControl("voice.osc1PulseWidth", "Osc1", "Pulse Width", 0.0, 1.0, 0.01, 0.5))
	synth.AddControl(control.NewBaseFloatControl("voice.osc1SineMix", "Osc1", "Sine Mix", 0.0, 1.0, 0.01, 0.0))
	synth.AddControl(control.NewBaseFloatControl("voice.osc1SawMix", "Osc1", "Saw Mix", 0.0, 1.0, 0.01, 0.0))
	synth.AddControl(control.NewBaseFloatControl("voice.osc1PulseMix", "Osc1", "Pulse Mix", 0.0, 1.0, 0.01, 1.0))
	synth.AddControl(control.NewBaseFloatControl("voice.osc1TriMix", "Osc1", "Tri Mix", 0.0, 1.0, 0.01, 0.0))
	synth.AddControl(control.NewBaseFloatControl("voice.osc2PulseWidth", "Osc2", "Pulse Width", 0.0, 1.0, 0.01, 0.5))
	synth.AddControl(control.NewBaseFloatControl("voice.osc2SineMix", "Osc2", "Sine Mix", 0.0, 1.0, 0.01, 0.0))
	synth.AddControl(control.NewBaseFloatControl("voice.osc2SawMix", "Osc2", "Saw Mix", 0.0, 1.0, 0.01, 0.0))
	synth.AddControl(control.NewBaseFloatControl("voice.osc2PulseMix", "Osc2", "Pulse Mix", 0.0, 1.0, 0.01, 1.0))
	synth.AddControl(control.NewBaseFloatControl("voice.osc2TriMix", "Osc2", "Tri Mix", 0.0, 1.0, 0.01, 0.0))
	synth.AddControl(control.NewBaseFloatControl("voice.osc2Tuning", "Osc2", "Tuning", 0.125, 8.0, 0.01, 2.0))
	synth.AddControl(control.NewBaseFloatControl("voice.pan", "Pan", "Pan", 0.0, 1.0, 0.01, 0.5))

	return synth
}

func (cs *ClassicSynth) AddControl(ctrl control.Control) {
	cs.Collection.AddControl(ctrl)
	ctrl.AddListener(cs)
}

func (cs *ClassicSynth) ControlChanged(ctrl control.Control, oldValue any, newValue any, setter any) {
	id := ctrl.Identifier()
	components := strings.Split(id, ".")

	if components[0] == "voice" {
		cs.poly.ReceiveMessage(map[string]any{
			"command":     "voice",
			components[1]: newValue,
		})
	}
}

func (cs *ClassicSynth) ReceiveMessage(msg any) []*muse.Message {
	content := msg.(map[string]any)

	for k, v := range content {
		ctrl := cs.ControlById(k)
		if ctrl != nil {
			if ctrl.Type() == control.Float {
				ctrl.(*control.BaseFloatControl).Set(v.(float64), nil)
			}
		}
	}

	return nil
}

func noteSequence(octave notes.Note) value.Valuer[any] {
	return value.NewAnd(
		[]value.Valuer[any]{
			// Row 1
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor.FreqAny(octave), arpeggio.Alternate, arpeggio.None, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.CMajor.FreqAny(octave), arpeggio.Converge, arpeggio.None, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Converge, arpeggio.None, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.AMinor.FreqAny(octave), arpeggio.Random, arpeggio.Exclusive, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor7.FreqAny(octave), arpeggio.Up, arpeggio.Inclusive, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Alternate, arpeggio.None, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor.FreqAny(octave), arpeggio.Converge, arpeggio.Inclusive, true), 2, 2),
			// Row 2
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Up, arpeggio.Inclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.AMajor7.FreqAny(octave), arpeggio.Alternate, arpeggio.None, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor.FreqAny(octave), arpeggio.Converge, arpeggio.Exclusive, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.EMinor.FreqAny(octave), arpeggio.Converge, arpeggio.Inclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.AMajor7.FreqAny(octave), arpeggio.Up, arpeggio.Inclusive, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor.FreqAny(octave), arpeggio.Random, arpeggio.Exclusive, false), 2, 2),
			// Row 3
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor.FreqAny(octave), arpeggio.Alternate, arpeggio.None, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.CMajor.FreqAny(octave), arpeggio.Converge, arpeggio.None, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Converge, arpeggio.None, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.AMinor.FreqAny(octave), arpeggio.Random, arpeggio.Exclusive, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor7.FreqAny(octave), arpeggio.Up, arpeggio.Inclusive, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Alternate, arpeggio.None, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor.FreqAny(octave), arpeggio.Converge, arpeggio.Inclusive, true), 2, 2),
			// Row 4
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Up, arpeggio.Inclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.AMajor7.FreqAny(octave), arpeggio.Alternate, arpeggio.None, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor.FreqAny(octave), arpeggio.Converge, arpeggio.Exclusive, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.BMinor.FreqAny(octave), arpeggio.Converge, arpeggio.Inclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.AMinor.FreqAny(octave), arpeggio.Up, arpeggio.Inclusive, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor7.FreqAny(octave), arpeggio.Random, arpeggio.Exclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Up, arpeggio.Inclusive, false), 2, 2),
			// Row 5
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.CMajor.FreqAny(octave), arpeggio.Up, arpeggio.Exclusive, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor7.FreqAny(octave), arpeggio.Alternate, arpeggio.None, false), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Converge, arpeggio.None, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.EMinor.FreqAny(octave), arpeggio.Converge, arpeggio.None, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.CMajor.FreqAny(octave), arpeggio.Random, arpeggio.Exclusive, false), 1, 1),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.DMajor7.FreqAny(octave), arpeggio.Up, arpeggio.Inclusive, true), 2, 2),
			value.NewRepeat[any](arpeggio.NewArpeggioNC(notes.GMajor.FreqAny(octave), arpeggio.Alternate, arpeggio.None, false), 1, 1),
		}, true)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	env := muse.NewEnvironment(2, 44100.0, 512)

	bpm := 100.0
	synth := NewClassicSynth(bpm, env.Config)

	env.AddModule(synth)

	muse.Connect(synth, 0, env, 0)
	muse.Connect(synth, 1, env, 1)

	synth.AddMessenger(banger.NewTemplateGenerator([]string{"poly"}, template.Template{
		"command":   "trigger",
		"duration":  value.NewSequence([]any{375.0, 750.0, 1000.0, 250.0, 250.0, 375.0, 750.0}),
		"amplitude": value.NewConst[any](1.0),
		"message": template.Template{
			"osc1SineMix":  value.NewSequence([]any{1.0, 0.7, 0.5, 0.3, 0.1, 0.0, 0.1, 0.2, 0.3, 0.5, 0.7}),
			"osc1SawMix":   value.NewSequence([]any{0.0, 0.1, 0.3, 0.5, 0.7, 1.0, 0.7, 0.5, 0.3, 0.2, 0.1}),
			"osc1PulseMix": value.NewSequence([]any{0.2, 0.5, 0.7, 1.0, 0.7, 0.5, 0.2, 0.1, 0.0}),
			"osc1TriMix":   value.NewSequence([]any{0.7, 1.0, 1.0, 0.7, 0.3, 0.1, 0.0, 0.0, 0.0}),
			"osc2SineMix":  value.NewSequence([]any{0.7, 1.0, 1.0, 0.7, 0.3, 0.1, 0.0, 0.0, 0.0}),
			"osc2SawMix":   value.NewSequence([]any{0.2, 0.5, 0.7, 1.0, 0.7, 0.5, 0.2, 0.1, 0.0}),
			"osc2PulseMix": value.NewSequence([]any{0.0, 0.1, 0.3, 0.5, 0.7, 1.0, 0.7, 0.5, 0.3, 0.2, 0.1}),
			"osc2TriMix":   value.NewSequence([]any{1.0, 0.7, 0.5, 0.3, 0.1, 0.0, 0.1, 0.2, 0.3, 0.5, 0.7}),
			"frequency":    noteSequence(notes.O3),
		},
	}, "control"))

	synth.AddMessenger(stepper.NewStepper(
		swing.New(value.NewConst(bpm), value.NewConst(2.0),
			value.NewSequence([]*swing.Step{{}, {Skip: true}}),
		),
		[]string{"control"}, "",
	))

	synth.AddMessenger(lfo.NewBasicLFO(0.14, 0.7, 0.15, []string{"synth"}, env.Config, "val", template.Template{
		"voice.osc1PulseWidth": template.NewParameter("val", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.103, 0.7, 0.15, []string{"synth"}, env.Config, "val", template.Template{
		"voice.osc2PulseWidth": template.NewParameter("val", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.085, 0.6, 0.25, []string{"synth"}, env.Config, "val", template.Template{
		"voice.filterResonance": template.NewParameter("val", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.115, 0.06, 4.0, []string{"synth"}, env.Config, "val", template.Template{
		"voice.osc2Tuning": template.NewParameter("val", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.0367, 0.1, 0.01, []string{"synth"}, env.Config, "val", template.Template{
		"voice.noiseMix": template.NewParameter("val", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.0567, 0.4, 0.3, []string{"synth"}, env.Config, "val", template.Template{
		"voice.osc1Mix": template.NewParameter("val", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.0667, 0.4, 0.2, []string{"synth"}, env.Config, "val", template.Template{
		"voice.osc2Mix": template.NewParameter("val", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.1067, 0.3, 0.35, []string{"synth"}, env.Config, "val", template.Template{
		"voice.pan": template.NewParameter("val", nil),
	}))

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/classic_synth.aiff", 240.0, env.Config.SampleRate, true, sndfile.SF_FORMAT_AIFF)

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

	filterFcMaxControl := synth.ControlById("voice.filterFcMax").(control.FloatControl)
	fcMaxBinding := binding.NewFloat()
	filterFcMaxControl.AddListener(control.NewChangeCallback(func(ctrl control.Control, oldValue any, newValue any, setter any) {
		if setter != fcMaxBinding {
			fcMaxBinding.Set(newValue.(float64))
		}
	}))
	fcMaxBinding.AddListener(binding.NewDataListener(func() {
		v, err := fcMaxBinding.Get()
		if err == nil {
			filterFcMaxControl.Set(v, fcMaxBinding)
		}
	}))
	fcMaxSlider := widget.NewSliderWithData(50.0, 8000.0, fcMaxBinding)
	fcMaxSlider.Step = 1.0

	w.SetContent(
		container.NewVBox(
			container.NewHBox(
				widget.NewButton("Start", func() {
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
				fcMaxSlider,
			),
		),
	)

	w.ShowAndRun()
}

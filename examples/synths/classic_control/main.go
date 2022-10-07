package main

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"time"

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
	"github.com/almerlucke/muse/utils/notes"
	"github.com/almerlucke/muse/value"
	"github.com/almerlucke/muse/value/arpeggio"
	"github.com/almerlucke/muse/value/template"
	"github.com/gordonklaus/portaudio"
)

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

type ClassicSynth struct {
	*muse.BasePatch
	controls  []control.Control
	ampEnv    *adsr.BasicStepProvider
	filterEnv *adsr.BasicStepProvider
	poly      *polyphony.Polyphony
	chorus1   *chorus.Chorus
	chorus2   *chorus.Chorus
}

func NewClassicSynth(bpm float64, config *muse.Configuration) *ClassicSynth {
	synth := &ClassicSynth{
		BasePatch: muse.NewPatch(0, 2, config, "classic"),
	}

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

	return synth
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

	synth.AddMessenger(lfo.NewBasicLFO(0.14, 0.7, 0.15, []string{"poly"}, env.Config, "pw", template.Template{
		"command":        "voice",
		"osc1PulseWidth": template.NewParameter("pw", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.103, 0.7, 0.15, []string{"poly"}, env.Config, "pw", template.Template{
		"command":        "voice",
		"osc2PulseWidth": template.NewParameter("pw", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.085, 0.6, 0.25, []string{"poly"}, env.Config, "resonance", template.Template{
		"command":         "voice",
		"filterResonance": template.NewParameter("resonance", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.115, 0.06, 4.0, []string{"poly"}, env.Config, "tuning", template.Template{
		"command":    "voice",
		"osc2Tuning": template.NewParameter("tuning", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.0367, 0.1, 0.01, []string{"poly"}, env.Config, "noise", template.Template{
		"command":  "voice",
		"noiseMix": template.NewParameter("noise", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.0567, 0.4, 0.3, []string{"poly"}, env.Config, "noise", template.Template{
		"command": "voice",
		"osc1Mix": template.NewParameter("noise", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.0667, 0.4, 0.2, []string{"poly"}, env.Config, "noise", template.Template{
		"command": "voice",
		"osc2Mix": template.NewParameter("noise", nil),
	}))

	synth.AddMessenger(lfo.NewBasicLFO(0.1067, 0.3, 0.35, []string{"poly"}, env.Config, "pan", template.Template{
		"command": "voice",
		"pan":     template.NewParameter("pan", nil),
	}))

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/classic_synth.aiff", 240.0, env.Config.SampleRate, true, sndfile.SF_FORMAT_AIFF)

	portaudio.Initialize()
	defer portaudio.Terminate()

	stream, err := env.PortaudioStream()
	if err != nil {
		log.Fatalf("error opening portaudio stream, %v", err)
	}

	defer stream.Close()

	stream.Start()

	reader := bufio.NewReader(os.Stdin)

	reader.ReadString('\n')
}

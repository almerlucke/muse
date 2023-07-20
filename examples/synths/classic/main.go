package main

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/envelopes/adsr"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/messengers/triggers/stepper/swing"
	"github.com/almerlucke/muse/modules/allpass"
	"github.com/almerlucke/muse/modules/chorus"
	"github.com/almerlucke/muse/modules/functor"
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

func main() {
	rand.Seed(time.Now().UnixNano())

	env := muse.NewEnvironment(2, 44100.0, 512)

	ampEnv := adsr.NewBasicStepProvider()
	ampEnv.Steps[0] = adsr.Step{Level: 1.0, Duration: 25.0}
	ampEnv.Steps[1] = adsr.Step{Level: 0.3, Duration: 80.0}
	ampEnv.Steps[3] = adsr.Step{Duration: 2000.0}

	filterEnv := adsr.NewBasicStepProvider()
	filterEnv.Steps[0] = adsr.Step{Level: 0.9, Duration: 25.0}
	filterEnv.Steps[1] = adsr.Step{Level: 0.5, Duration: 80.0}
	filterEnv.Steps[3] = adsr.Step{Duration: 2000.0}

	bpm := 100

	synth := classic.NewSynth(20, ampEnv, filterEnv, env.Config).Named("poly").Add(env)
	synthAmp1 := env.AddModule(functor.NewFunctor(1, func(v []float64) float64 { return v[0] * 0.85 }, env.Config))
	synthAmp2 := env.AddModule(functor.NewFunctor(1, func(v []float64) float64 { return v[0] * 0.85 }, env.Config))
	allpass1 := env.AddModule(allpass.NewAllpass(2500.0, 60000.0/float64(bpm)*1.666, 0.5, env.Config, "allpass"))
	allpass2 := env.AddModule(allpass.NewAllpass(2500.0, 60000.0/float64(bpm)*1.75, 0.4, env.Config, "allpass"))
	allpassAmp1 := env.AddModule(functor.NewFunctor(1, func(v []float64) float64 { return v[0] * 0.5 }, env.Config))
	allpassAmp2 := env.AddModule(functor.NewFunctor(1, func(v []float64) float64 { return v[0] * 0.5 }, env.Config))
	chor1 := env.AddModule(chorus.NewChorus(false, 15, 10, 0.3, 1.42, 0.5, nil, env.Config, "chorus1"))
	chor2 := env.AddModule(chorus.NewChorus(false, 15, 10, 0.31, 1.43, 0.55, nil, env.Config, "chorus2"))

	env.AddMessenger(banger.NewTemplateGenerator([]string{"poly"}, template.Template{
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

	env.AddMessenger(stepper.NewStepper(
		swing.New(bpm, 2,
			value.NewSequence([]*swing.Step{{}, {Skip: true}}),
		),
		[]string{"control"}, "",
	))

	env.AddMessenger(lfo.NewBasicLFO(0.14, 0.7, 0.15, []string{"poly"}, env.Config, "pw", template.Template{
		"command":        "voice",
		"osc1PulseWidth": template.NewParameter("pw", nil),
	}))

	env.AddMessenger(lfo.NewBasicLFO(0.103, 0.7, 0.15, []string{"poly"}, env.Config, "pw", template.Template{
		"command":        "voice",
		"osc2PulseWidth": template.NewParameter("pw", nil),
	}))

	env.AddMessenger(lfo.NewBasicLFO(0.085, 0.6, 0.25, []string{"poly"}, env.Config, "resonance", template.Template{
		"command":         "voice",
		"filterResonance": template.NewParameter("resonance", nil),
	}))

	env.AddMessenger(lfo.NewBasicLFO(0.115, 0.06, 4.0, []string{"poly"}, env.Config, "tuning", template.Template{
		"command":    "voice",
		"osc2Tuning": template.NewParameter("tuning", nil),
	}))

	env.AddMessenger(lfo.NewBasicLFO(0.0367, 0.1, 0.01, []string{"poly"}, env.Config, "noise", template.Template{
		"command":  "voice",
		"noiseMix": template.NewParameter("noise", nil),
	}))

	env.AddMessenger(lfo.NewBasicLFO(0.0567, 0.4, 0.3, []string{"poly"}, env.Config, "noise", template.Template{
		"command": "voice",
		"osc1Mix": template.NewParameter("noise", nil),
	}))

	env.AddMessenger(lfo.NewBasicLFO(0.0667, 0.4, 0.2, []string{"poly"}, env.Config, "noise", template.Template{
		"command": "voice",
		"osc2Mix": template.NewParameter("noise", nil),
	}))

	env.AddMessenger(lfo.NewBasicLFO(0.1067, 0.3, 0.35, []string{"poly"}, env.Config, "pan", template.Template{
		"command": "voice",
		"pan":     template.NewParameter("pan", nil),
	}))

	synth.Connect(0, synthAmp1, 0)
	synth.Connect(1, synthAmp2, 0)
	synthAmp1.Connect(0, chor1, 0)
	synthAmp2.Connect(0, chor2, 0)
	synthAmp1.Connect(0, allpass1, 0)
	synthAmp2.Connect(0, allpass2, 0)
	allpass1.Connect(0, allpassAmp1, 0)
	allpass2.Connect(0, allpassAmp2, 0)
	allpassAmp1.Connect(0, chor1, 0)
	allpassAmp2.Connect(0, chor2, 0)
	chor1.Connect(0, env, 0)
	chor2.Connect(0, env, 1)

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

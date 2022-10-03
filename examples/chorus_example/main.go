package main

import (
	"bufio"
	"log"
	"os"

	"github.com/almerlucke/muse"
	shaping "github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/lfo"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/modules/blosc"
	"github.com/almerlucke/muse/modules/chorus"
	"github.com/almerlucke/muse/modules/filters/moog"
	"github.com/almerlucke/muse/utils/notes"
	"github.com/almerlucke/muse/value"
	"github.com/almerlucke/muse/value/template"
	"github.com/gordonklaus/portaudio"
	"github.com/mkb218/gosndfile/sndfile"
)

func makeLFO(speed float64, targets []string, shaper shaping.Shaper, config *muse.Configuration, param string, templ template.Template) *lfo.LFO {
	ts := make([]*lfo.Target, len(targets))
	for i, target := range targets {
		ts[i] = lfo.NewTarget(target, shaper, param, templ)
	}

	return lfo.NewLFO(speed, ts, config, "")
}

func main() {
	env := muse.NewEnvironment(2, 44100, 128)

	env.AddMessenger(banger.NewTemplateGenerator([]string{"osc"}, template.Template{
		"frequency": value.NewSequence([]any{notes.A3.Freq(), notes.C3.Freq(), notes.E3.Freq()}),
	}, "sequencer"))

	env.AddMessenger(banger.NewTemplateGenerator([]string{"osc2"}, template.Template{
		"frequency": value.NewSequence([]any{notes.A2.Freq(), notes.C2.Freq(), notes.E2.Freq()}),
	}, "sequencer2"))

	env.AddMessenger(stepper.NewStepper(
		stepper.NewValueStepper(value.NewSequence([]float64{250, -125, 250, 250, -125, 125, -125, 250})),
		[]string{"sequencer", "sequencer2"}, "",
	))

	sineTable := shaping.NewNormalizedSineTable(128)
	oscScale := shaping.NewLinear(0.7, 0.1)
	mixScale := shaping.NewLinear(1.0, 0.0)

	env.AddMessenger(makeLFO(0.24, []string{"osc"}, shaping.NewChain(sineTable, oscScale), env.Config, "pw", template.Template{
		"pulseWidth": template.NewParameter("pw", nil),
	}))
	env.AddMessenger(makeLFO(0.13, []string{"osc2"}, shaping.NewChain(sineTable, oscScale), env.Config, "pw", template.Template{
		"pulseWidth": template.NewParameter("pw", nil),
	}))
	env.AddMessenger(makeLFO(0.16, []string{"filter"}, shaping.NewChain(sineTable, shaping.NewLinear(8500.0, 1300.0)), env.Config, "freq", template.Template{
		"frequency": template.NewParameter("freq", nil),
	}))

	env.AddMessenger(makeLFO(0.15, []string{"osc"}, shaping.NewChain(sineTable, mixScale), env.Config, "mix", template.Template{
		"mix1": template.NewParameter("mix", nil),
	}))
	env.AddMessenger(makeLFO(0.17, []string{"osc"}, shaping.NewChain(sineTable, mixScale), env.Config, "mix", template.Template{
		"mix2": template.NewParameter("mix", nil),
	}))
	env.AddMessenger(makeLFO(0.19, []string{"osc"}, shaping.NewChain(sineTable, mixScale), env.Config, "mix", template.Template{
		"mix3": template.NewParameter("mix", nil),
	}))
	env.AddMessenger(makeLFO(0.21, []string{"osc"}, shaping.NewChain(sineTable, mixScale), env.Config, "mix", template.Template{
		"mix4": template.NewParameter("mix", nil),
	}))

	env.AddMessenger(makeLFO(0.25, []string{"osc2"}, shaping.NewChain(sineTable, mixScale), env.Config, "mix", template.Template{
		"mix1": template.NewParameter("mix", nil),
	}))
	env.AddMessenger(makeLFO(0.37, []string{"osc2"}, shaping.NewChain(sineTable, mixScale), env.Config, "mix", template.Template{
		"mix2": template.NewParameter("mix", nil),
	}))
	env.AddMessenger(makeLFO(0.09, []string{"osc2"}, shaping.NewChain(sineTable, mixScale), env.Config, "mix", template.Template{
		"mix3": template.NewParameter("mix", nil),
	}))
	env.AddMessenger(makeLFO(0.11, []string{"osc2"}, shaping.NewChain(sineTable, mixScale), env.Config, "mix", template.Template{
		"mix4": template.NewParameter("mix", nil),
	}))

	osc := env.AddModule(blosc.NewOscX(100.0, 0.0, 0.2, [4]float64{0.1, 0.1, 0.4, 0.1}, env.Config, "osc"))
	osc2 := env.AddModule(blosc.NewOscX(100.0, 0.5, 0.2, [4]float64{0.1, 0.1, 0.4, 0.1}, env.Config, "osc2"))
	filter := env.AddModule(moog.NewMoog(300.0, 0.63, 0.7, env.Config, "filter"))
	ch := env.AddModule(chorus.NewChorus(true, 15, 10, 0.4, 1.6, 0.6, shaping.NewSineTable(512.0), env.Config, "chorus"))

	muse.Connect(osc, 4, filter, 0)
	muse.Connect(osc2, 4, filter, 0)
	muse.Connect(filter, 0, ch, 0)
	muse.Connect(ch, 0, env, 0)
	muse.Connect(ch, 0, env, 1)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/test.aiff", 25.0, env.Config.SampleRate, true, sndfile.SF_FORMAT_AIFF)

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

	// env := muse.NewEnvironment(1, 44100, 128)

	// env.AddMessenger(banger.NewTemplateGenerator([]string{"osc"}, template.Template{
	// 	"frequency": value.NewSequence([]any{notes.A4.Freq(), notes.C4.Freq(), notes.E4.Freq()}),
	// }, "sequencer"))

	// env.AddMessenger(stepper.NewStepper(
	// 	stepper.NewValueStepper(value.NewSequence([]float64{-250, 250, 250, 250, -250, 500, -250})),
	// 	[]string{"sequencer", "adsr"}, "",
	// ))

	// steps := []adsrc.Step{
	// 	{Level: 1.0, DurationRatio: 0.2, Shape: 0.1},
	// 	{Level: 0.4, DurationRatio: 0.2, Shape: -0.1},
	// 	{DurationRatio: 0.1},
	// 	{DurationRatio: 0.4, Shape: -0.1},
	// }

	// adsrEnv := env.AddModule(adsr.NewADSR(steps, adsrc.Ratio, adsrc.Automatic, 1.0, env.Config, "adsr"))
	// mult := env.AddModule(functor.NewFunctor(2, functor.FunctorMult, env.Config, ""))
	// osc := env.AddModule(blosc.NewBloscModule(100.0, 0.0, 1.0, env.Config, "osc"))
	// ch := env.AddModule(chorus.NewChorus(waveshaping.NewSineTable(512), 0.9, 2.4, env.Config, "chorus"))

	// muse.Connect(osc, 1, mult, 0)
	// muse.Connect(adsrEnv, 0, mult, 1)
	// muse.Connect(mult, 0, ch, 0)
	// muse.Connect(ch, 0, env, 0)

	// portaudio.Initialize()
	// defer portaudio.Terminate()

	// stream, err := env.PortaudioStream()
	// if err != nil {
	// 	log.Fatalf("error opening portaudio stream, %v", err)
	// }

	// defer stream.Close()

	// stream.Start()

	// reader := bufio.NewReader(os.Stdin)

	// reader.ReadString('\n')
}

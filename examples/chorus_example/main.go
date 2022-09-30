package main

import (
	"bufio"
	"log"
	"os"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/components/waveshaping"
	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/modules/blosc"
	"github.com/almerlucke/muse/modules/chorus"
	"github.com/almerlucke/muse/utils/notes"
	"github.com/almerlucke/muse/value"
	"github.com/almerlucke/muse/value/template"
	"github.com/gordonklaus/portaudio"
)

func main() {
	env := muse.NewEnvironment(1, 44100, 128)

	env.AddMessenger(banger.NewTemplateGenerator([]string{"osc"}, template.Template{
		"frequency": value.NewSequence([]any{notes.A4.Freq(), notes.C4.Freq(), notes.E4.Freq()}),
	}, "sequencer"))

	env.AddMessenger(stepper.NewStepper(
		stepper.NewValueStepper(value.NewSequence([]float64{-250, 250, 250, 250, -250, 500, -250})),
		[]string{"sequencer"}, "",
	))

	// adsrEnv := env.AddModule(adsr.NewADSR(steps, adsrc.Ratio, adsrc.Automatic, 1.0, env.Config, "adsr"))
	// mult := env.AddModule(functor.NewFunctor(2, functor.FunctorMult, env.Config, ""))
	osc := env.AddModule(blosc.NewBloscModule(100.0, 0.0, 1.0, env.Config, "osc"))
	ch := env.AddModule(chorus.NewChorus(false, 15, 10, 0.7, 1.6, 0.3, waveshaping.NewSineTable(512.0), env.Config, "chorus"))

	muse.Connect(osc, 0, ch, 0)
	muse.Connect(ch, 0, env, 0)

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

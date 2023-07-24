package main

import (
	"bufio"
	"log"
	"os"

	"github.com/almerlucke/muse"
	"github.com/gordonklaus/portaudio"

	"github.com/almerlucke/muse/utils"
	"github.com/almerlucke/muse/value"

	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/modules/osc"
)

func main() {
	env := muse.NewEnvironment(1)

	sequence := value.NewSequence(utils.ReadJSONNull[[][]*muse.Message]("examples/osc/sequence1.json"))

	banger.NewValueGenerator(sequence).MsgrNamed("sequencer").MsgrAdd(env)

	stepper.NewStepper(
		stepper.NewValueStepProvider(value.NewSequence([]float64{250, -125, 250, 250, -125, 125, -125, 250})),
		[]string{"sequencer"},
	).MsgrAdd(env)

	osc := osc.New(100.0, 0.0).Add(env)

	env.In(osc, 3)

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/test.aiff", 10.0, env.Config.SampleRate, false, sndfile.SF_FORMAT_AIFF)

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

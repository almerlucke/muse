package main

import (
	"github.com/almerlucke/muse"

	"github.com/almerlucke/muse/utils"
	"github.com/almerlucke/muse/value"

	"github.com/almerlucke/muse/messengers/banger"
	"github.com/almerlucke/muse/messengers/triggers/stepper"
	"github.com/almerlucke/muse/modules/osc"
)

func main() {
	root := muse.New(1)

	sequence := value.NewSequence(utils.ReadJSONNull[[][]*muse.Message]("examples/osc/sequence1.json"))

	banger.NewValueGenerator(sequence).MsgrNamed("sequencer").MsgrAdd(root)

	stepper.NewStepper(
		stepper.NewValueStepProvider(value.NewSequence([]float64{250, -125, 250, 250, -125, 125, -125, 250})),
		[]string{"sequencer"},
	).MsgrAdd(root)

	osc := osc.New(100.0, 0.0).Add(root)

	root.In(osc, 3)

	root.RenderLive()

	// env.SynthesizeToFile("/Users/almerlucke/Desktop/test.aiff", 10.0, env.Config.SampleRate, false, sndfile.SF_FORMAT_AIFF)
}

package main

import (
	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/modules"
)

func main() {
	env := muse.NewEnvironment(2, 44100, 512)

	osc := modules.NewBloscModule(300.0, 0.0, 1.0, "blosc")

	env.AddModule(osc)

	muse.Connect(osc, 0, env, 0)
	muse.Connect(osc, 1, env, 1)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/test.aiff", 4.0)

	// env.PrepareBuffers()

	// swr, err := muse.OpenSoundWriter("/Users/almerlucke/Desktop/test.aiff", 1, int32(env.Config.SampleRate), true)
	// if err != nil {
	// 	log.Fatalf("error opening sound writer: %v", err)
	// }

	// defer swr.Close()

	// numSeconds := 4.0

	// framesToProduce := int64(env.Config.SampleRate * numSeconds)

	// for framesToProduce > 0 {
	// 	env.Produce()

	// 	if framesToProduce > int64(env.Config.BufferSize) {
	// 		swr.WriteSamples(env.OutputAtIndex(0).Buffer)
	// 		framesToProduce -= int64(env.Config.BufferSize)
	// 	} else {
	// 		swr.WriteSamples(env.OutputAtIndex(0).Buffer[:framesToProduce])
	// 		framesToProduce = 0
	// 	}
	// }
}

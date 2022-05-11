package main

import (
	"log"
	"os"

	"github.com/almerlucke/muse"
	"github.com/almerlucke/muse/messengers"
	"github.com/almerlucke/muse/modules/blosc"
)

func main() {
	env := muse.NewEnvironment(2, 44100, 128)

	data, err := os.ReadFile("examples/blosc_example/events.json")
	if err != nil {
		log.Fatalf("error reading events: %v", err)
	}

	msgr, err := messengers.NewSchedulerWithJSONData(data)
	if err != nil {
		log.Fatalf("error unmarshalling json events: %v", err)
	}

	osc1 := blosc.NewBloscModule(100.0, 0.0, 0.1, env.Config, "blosc1")
	osc2 := blosc.NewBloscModule(400.0, 0.0, 1.0, env.Config, "blosc2")

	env.AddMessenger(msgr)
	env.AddModule(osc1)
	env.AddModule(osc2)

	muse.Connect(osc1, 0, osc2, 1)
	muse.Connect(osc2, 0, env, 0)

	env.SynthesizeToFile("/Users/almerlucke/Desktop/test.aiff", 5.0)
}

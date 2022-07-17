package main

import (
	"log"

	params "github.com/almerlucke/muse/parameters"
)

/*
"duration":  duration,
			"amplitude": amplitude,
			"message":   map[string]any{"osc": map[string]any{"frequency": frequency}},

			250, 0.4, 200.0),
		NewVoiceMessage("voicePlayer", 125, 1.0, 400.0),
		NewVoiceMessage("voicePlayer", 500, 0.6, 500.0),
		NewVoiceMessage("voicePlayer", 125, 1.0, 600.0),
		NewVoiceMessage("voicePlayer", 250, 0.5, 100.0),
		NewVoiceMessage("voicePlayer", 250, 0.5, 50.0),
		NewVoiceMessage("voicePlayer", 750, 1.0, 50.0),
		NewVoiceMessage("voicePlayer", 500, 0.3, 100.0),
		NewVoiceMessage("voicePlayer", 375, 1.0, 250.0),
		NewVoiceMessage("voicePlayer", 250, 0.7, 750.0),
		NewVoiceMessage("voicePlayer", 250,
*/

func main() {
	p := params.Prototype{
		"duration":  params.NewSequence([]any{250.0, 500.0, 125.0, 250.0, 250.0, 750.0, 500.0, 375.0, 250.0, 250.0}),
		"amplitude": params.NewSequence([]any{1.0, 0.6, 1.0, 0.5, 0.5, 1.0, 0.3, 1.0, 0.7, 1.0}),
		"message": params.Prototype{
			"osc": params.Prototype{
				"frequency": params.NewSequence([]any{400.0, 500.0, 600.0, 100.0, 50.0, 50.0, 100.0, 250.0, 750.0}),
				"phase":     params.NewConst(0.0),
			},
		},
	}

	m := p.Map()
	log.Printf("d: %v", m.F("duration"))
	log.Printf("f: %v", m.M("message").M("osc").F("frequency"))
	log.Printf("p: %v", m.M("message").M("osc").F("phase"))
	m = p.Map()
	log.Printf("d: %v", m.F("duration"))
	log.Printf("f: %v", m.M("message").M("osc").F("frequency"))
	log.Printf("p: %v", m.M("message").M("osc").F("phase"))
	m = p.Map()
	log.Printf("d: %v", m.F("duration"))
	log.Printf("f: %v", m.M("message").M("osc").F("frequency"))
	log.Printf("p: %v", m.M("message").M("osc").F("phase"))
}

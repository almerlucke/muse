package main

import (
	"log"

	params "github.com/almerlucke/muse/parameters"
)

func main() {
	p := params.Prototype{
		"osc": params.Prototype{
			"frequency": params.NewSequence([]any{1.0, 2.0, 3.0, 4.0}),
			"phase":     params.NewConst(0.1),
		},
	}

	m := p.Map()
	log.Printf("f: %v", m.M("osc").F("frequency"))
	log.Printf("p: %v", m.M("osc").F("phase"))
	m = p.Map()
	log.Printf("f: %v", m.M("osc").F("frequency"))
	log.Printf("p: %v", m.M("osc").F("phase"))
	m = p.Map()
	log.Printf("f: %v", m.M("osc").F("frequency"))
	log.Printf("p: %v", m.M("osc").F("phase"))
}

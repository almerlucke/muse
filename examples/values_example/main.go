package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/almerlucke/muse/values"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	patt := values.NewRepeat[float64](values.NewPattern(
		[]values.Generator[float64]{values.NewSequence([]float64{1.0, 2.0, 3.0}, false), values.NewSequence([]float64{4.0, 5.0, 6.0}, false)},
		false,
	), 2, 3)

	for !patt.Finished() {
		v := patt.Next()
		log.Printf("v: %v", v)
	}

	p := values.MapPrototype{
		"duration":  values.NewSequence([]any{250.0, 500.0, 125.0, 250.0, 250.0, 750.0, 500.0, 375.0, 250.0, 250.0}, true),
		"amplitude": values.NewSequence([]any{1.0, 0.6, 1.0, 0.5, 0.5, 1.0, 0.3, 1.0, 0.7, 1.0}, true),
		"message": values.MapPrototype{
			"osc": values.MapPrototype{
				"frequency": values.NewSequence([]any{400.0, 500.0, 600.0, 100.0, 50.0, 50.0, 100.0, 250.0, 750.0}, true),
				"phase":     values.NewConst[any](0.0),
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

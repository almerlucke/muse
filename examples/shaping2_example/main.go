package main

import (
	"log"

	"github.com/almerlucke/muse/components/shaping"
	val "github.com/almerlucke/muse/values"
)

func main() {
	table := shaping.NewSineTable(128)
	phase := 0.0
	inc := 1.0 / 200.0

	for i := 0; i < 1000; i++ {
		v := table.Shape(phase)
		log.Printf("v: %v", v)
		phase += inc
		for phase >= 1.0 {
			phase -= 1.0
		}
		for phase < 0.0 {
			phase += 1.0
		}
	}

	placeholder := val.NewPlaceholder("testHolder")
	m := val.Prototype{"hola": val.Prototype{"test1": 2, "test": placeholder}, "hola2": placeholder}

	log.Printf("m: %v", m.Map([]*val.Replacement{val.NewReplacement("testHolder", 13.0)}))
}

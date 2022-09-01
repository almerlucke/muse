package main

import (
	"log"

	"github.com/almerlucke/muse/components/shaping"
	"github.com/almerlucke/muse/values"
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

	placeholder := values.NewPlaceholder("testHolder")
	m := values.MapPrototype{"hola": values.MapPrototype{"test1": 2, "test": placeholder}, "hola2": placeholder}

	log.Printf("m: %v", m.Map([]string{"testHolder"}, []any{13.0}))
}

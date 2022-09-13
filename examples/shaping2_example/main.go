package main

import (
	"log"
)

func clip(x float64) float64 {
	for x >= 1.0 {
		x -= 1.0
	}
	for x < 0.0 {
		x += 1.0
	}

	return x
}

func main() {
	// slope := shaping.NewParallelF(1.0, func(v1 float64, v2 float64) float64 { return v1 * v2 }, shaping.NewChain(
	// 	shaping.NewAdd(),
	// 	shaping.NewPulse(0.5),
	// ), shaping.NewThru(),
	// )

	inc1 := 1.0 / 200.0
	w := 0.35
	phase1 := 0.0
	phase2 := w

	for i := 0; i < 1000; i++ {
		p2 := clip(phase2 + w)
		log.Printf("p1 %f - p2 %f - v %f", phase1, phase2, phase1-p2+w)

		phase1 = clip(phase1 + inc1)
		phase2 = clip(phase2 + inc1)
	}
}

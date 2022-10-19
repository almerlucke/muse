package main

import (
	"log"

	"github.com/almerlucke/muse/components/waveshaping"
)

func main() {
	hw := waveshaping.NewHanningWindow(512.0)

	for i := 0; i < 512; i++ {
		log.Printf("wave: %v", hw.Shape(float64(i)/512.0))
	}
}

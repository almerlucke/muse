package utils

import "math"

type Factory[T any] interface {
	New() T
}

func Mtof(midiNote int) float64 {
	return math.Pow(2, float64(midiNote-69)/12.0) * 440.0
}

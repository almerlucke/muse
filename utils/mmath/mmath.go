package mmath

import "math"

func Quantize(v float64, q float64) float64 {
	return math.Round(v/q) * q
}

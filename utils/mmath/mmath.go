package mmath

import "math"

func Quantize(v float64, q float64) float64 {
	return math.Round(v/q) * q
}

func Limit(x float64, min float64, max float64) float64 {
	if x < min {
		return min
	}

	if x > max {
		return max
	}

	return x
}

func Db2Rap(dB float64) float64 {
	return math.Exp(dB * math.Log10E / 20.0)
}

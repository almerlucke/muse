package mmath

import "math"

func Quantize(v float64, q float64) float64 {
	return math.Round(v/q) * q
}

func PolyBlep(t float64, dt float64) float64 {
	if t < dt {
		t /= dt
		return t + t - t*t - 1.0
	} else if t > 1.0-dt {
		t = (t - 1.0) / dt
		return t*t + t + t + 1.0
	}

	return 0.0
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

package float

import "math"

var DenormGuard = 1e-15

var SampleEpsilon = 1.0e-10

func Equal(x1 float64, x2 float64) bool {
	return math.Abs(x1-x2) < SampleEpsilon
}

func ZeroIfSmall(x float64) float64 {
	if math.Abs(x) < SampleEpsilon {
		return 0.0
	}

	return x
}

package butterworth

import "math"

const (
	buddaQScale = 6.0
)

type Butterworth struct {
	coef0, coef1, coef2, coef3             float64
	history1, history2, history3, history4 float64
	gain                                   float64
}

func (bwc *Butterworth) Set(fc float64, q float64, fs float64) {
	t0 := 4.0 * fs * fs
	t1 := 8.0 * fs * fs
	t2 := 2.0 * fs
	t3 := math.Pi / fs

	wp := t2 * math.Tan(t3*fc)

	q *= buddaQScale
	q += 1.0

	b1 := (0.765367 / q) / wp
	b2 := 1.0 / (wp * wp)
	bdTmp := t0*b2 + 1.0
	bd := 1.0 / (bdTmp + t2*b1)

	bwc.gain = bd
	bwc.coef2 = 2.0 - t1*b2
	bwc.coef0 = bwc.coef2 * bd
	bwc.coef1 = (bdTmp - t2*b1) * bd

	b1 = (1.847759 / q) / wp
	bd = 1.0 / (bdTmp + t2*b1)

	bwc.gain *= bd
	bwc.coef2 *= bd
	bwc.coef3 = (bdTmp - t2*b1) * bd
}

func (bwc *Butterworth) Process(input float64) float64 {
	output := input * bwc.gain

	output -= bwc.history1 * bwc.coef0
	newHist := output - bwc.history2*bwc.coef1

	output = newHist + bwc.history1*2.0
	output += bwc.history2

	bwc.history2 = bwc.history1
	bwc.history1 = newHist

	output -= bwc.history3 * bwc.coef2
	newHist = output - bwc.history4*bwc.coef3

	output = newHist + bwc.history3*2.0
	output += bwc.history4

	bwc.history4 = bwc.history3
	bwc.history3 = newHist

	return output
}

package rbj

import "math"

type FilterType int

const (
	Lowpass FilterType = iota
	Highpass
	BandpassCSG
	BandpassCZPG
	Notch
	Allpass
	Peaking
	Lowshelf
	Highshelf
)

type Filter struct {
	b0a0 float64
	b1a0 float64
	b2a0 float64
	a1a0 float64
	a2a0 float64
	ou1  float64
	ou2  float64
	in1  float64
	in2  float64

	FilterType   FilterType
	Frequency    float64
	Q            float64
	DbGain       float64
	QIsBandwidth bool
}

func NewFilter(filterType FilterType, frequency float64, q float64, dbGain float64, qIsBandwidth bool, sampleRate float64) *Filter {
	filter := &Filter{
		FilterType:   filterType,
		Frequency:    frequency,
		Q:            q,
		DbGain:       dbGain,
		QIsBandwidth: qIsBandwidth,
	}

	filter.Update(sampleRate)

	return filter
}

func (r *Filter) Process(in0 float64) float64 {
	yn := r.b0a0*in0 + r.b1a0*r.in1 + r.b2a0*r.in2 - r.a1a0*r.ou1 - r.a2a0*r.ou2

	r.in2 = r.in1
	r.in1 = in0
	r.ou2 = r.ou1
	r.ou1 = yn

	return yn
}

func (r *Filter) Update(sr float64) {
	var alpha, a0, a1, a2, b0, b1, b2 float64

	omega := 2.0 * math.Pi * r.Frequency / sr
	tsin := math.Sin(omega)
	tcos := math.Cos(omega)

	if r.QIsBandwidth {
		alpha = tsin * math.Sinh(math.Log(2.0)/2.0*r.Q*omega/tsin)
	} else {
		alpha = tsin / (2.0 * r.Q)
	}

	A := math.Pow(10.0, r.DbGain/40.0)
	beta := math.Sqrt(A) / r.Q

	switch r.FilterType {
	case Peaking:
		b0 = 1.0 + alpha*A
		b1 = -2.0 * tcos
		b2 = 1.0 - alpha*A
		a0 = 1.0 + alpha/A
		a1 = -2.0 * tcos
		a2 = 1.0 - alpha/A
	case Lowshelf:
		b0 = A * ((A + 1.0) - (A-1.0)*tcos + beta*tsin)
		b1 = 2.0 * A * ((A - 1.0) - (A+1.0)*tcos)
		b2 = A * ((A + 1.0) - (A-1.0)*tcos - beta*tsin)
		a0 = (A + 1.0) + (A-1.0)*tcos + beta*tsin
		a1 = -2.0 * ((A - 1.0) + (A+1.0)*tcos)
		a2 = (A + 1.0) + (A-1.0)*tcos - beta*tsin
	case Highshelf:
		b0 = A * ((A + 1.0) + (A-1.0)*tcos + beta*tsin)
		b1 = -2.0 * A * ((A - 1.0) + (A+1.0)*tcos)
		b2 = A * ((A + 1.0) + (A-1.0)*tcos - beta*tsin)
		a0 = (A + 1.0) - (A-1.0)*tcos + beta*tsin
		a1 = 2.0 * ((A - 1.0) - (A+1.0)*tcos)
		a2 = (A + 1.0) - (A-1.0)*tcos - beta*tsin
	case Lowpass:
		b0 = (1.0 - tcos) / 2.0
		b1 = 1.0 - tcos
		b2 = (1.0 - tcos) / 2.0
		a0 = 1.0 + alpha
		a1 = -2.0 * tcos
		a2 = 1.0 - alpha
	case Highpass:
		b0 = (1.0 + tcos) / 2.0
		b1 = -(1.0 + tcos)
		b2 = (1.0 + tcos) / 2.0
		a0 = 1.0 + alpha
		a1 = -2.0 * tcos
		a2 = 1.0 - alpha
	case BandpassCSG:
		b0 = tsin / 2.0
		b1 = 0.0
		b2 = -tsin / 2
		a0 = 1.0 + alpha
		a1 = -2.0 * tcos
		a2 = 1.0 - alpha
	case BandpassCZPG:
		b0 = alpha
		b1 = 0.0
		b2 = -alpha
		a0 = 1.0 + alpha
		a1 = -2.0 * tcos
		a2 = 1.0 - alpha
	case Notch:
		b0 = 1.0
		b1 = -2.0 * tcos
		b2 = 1.0
		a0 = 1.0 + alpha
		a1 = -2.0 * tcos
		a2 = 1.0 - alpha
	case Allpass:
		b0 = 1.0 - alpha
		b1 = -2.0 * tcos
		b2 = 1.0 + alpha
		a0 = 1.0 + alpha
		a1 = -2.0 * tcos
		a2 = 1.0 - alpha
	}

	// set filter coeffs
	r.b0a0 = b0 / a0
	r.b1a0 = b1 / a0
	r.b2a0 = b2 / a0
	r.a1a0 = a1 / a0
	r.a2a0 = a2 / a0
}

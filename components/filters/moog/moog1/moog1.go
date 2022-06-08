package moog1

import (
	"math"
)

type OnePole struct {
	x0 float64
	x1 float64
	g  float64
}

func (op *OnePole) Update(x float64) float64 {
	y := (x*0.7692307692+op.x0*0.2307692308-op.x1)*op.g + op.x1
	op.x0 = x
	op.x1 = y
	return y
}

type Moog1c struct {
	poles      [4]OnePole
	coefsMix   [5]float64
	y1         float64
	saturation float64
	comp       float64
	res        float64
}

func NewMoog1c(fc float64, res float64, saturation float64, fs float64) *Moog1c {
	m := &Moog1c{
		comp: 0.5,
	}

	m.coefsMix[4] = 1.0
	m.Set(fc, res, saturation, fs)

	return m
}

func (m *Moog1c) Set(fc float64, res float64, saturation float64, fs float64) {
	g := 1.0 - math.Exp(-2.0*math.Pi*fc/fs)
	for index := range m.poles {
		m.poles[index].g = g
	}

	m.res = res
	m.saturation = saturation
}

func (m *Moog1c) Process(x float64) float64 {
	a := x - (x*m.comp+math.Tanh(m.y1*m.saturation))*m.res*4.0
	b := m.poles[0].Update(a)
	c := m.poles[1].Update(b)
	d := m.poles[2].Update(c)
	e := m.poles[3].Update(d)
	m.y1 = e
	return m.coefsMix[0]*a + m.coefsMix[1]*b + m.coefsMix[2]*c + m.coefsMix[3]*d + m.coefsMix[4]*e
}

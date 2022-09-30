package waveshaping

import (
	"math"
)

type Shaper interface {
	Shape(float64) float64
}

type ShapeFunction func(float64) float64

type Chain struct {
	Shapers []Shaper
}

func (c *Chain) Shape(x float64) float64 {
	for _, shaper := range c.Shapers {
		x = shaper.Shape(x)
	}

	return x
}

func NewChain(shapers ...Shaper) *Chain {
	return &Chain{
		Shapers: shapers,
	}
}

type Thru struct {
}

func (t *Thru) Shape(x float64) float64 {
	return x
}

func NewThru() *Thru {
	return &Thru{}
}

type LookupTable []float64

func (lt LookupTable) Shape(x float64) float64 {
	nf := x * float64(len(lt)-1)
	n1 := int(nf)
	n2 := n1 + 1
	fr := nf - float64(n1)
	return lt[n1] + fr*(lt[n2]-lt[n1])
}

func NewSineTable(n int) LookupTable {
	table := make(LookupTable, n)
	phase := 0.0
	inc := 2.0 * math.Pi / float64(n-1)

	for i := 0; i < n; i++ {
		table[i] = math.Sin(phase)
		phase += inc
	}

	return table
}

func NewNormalizedSineTable(n int) LookupTable {
	table := make(LookupTable, n)
	phase := 0.0
	inc := 2.0 * math.Pi / float64(n-1)

	for i := 0; i < n; i++ {
		table[i] = math.Sin(phase)*0.5 + 0.5
		phase += inc
	}

	return table
}

// func NewHarmonicsTable(n int, harmonics int) LookupTable {
// 	table := make(LookupTable, n)

// 	for i := 0; i < n; i++ {
// 		inc := float64(i) / float64(n-1)
// 		acc := 0.0
// 		for j := 1; j <= harmonics; j++ {
// 			acc += math.Sin(2.0 * math.Pi * inc * float64(j) * 2.0)
// 		}
// 		table[i] = acc
// 	}

// 	return table
// }

type ParallelFunction func(float64, float64) float64

type Parallel struct {
	Shapers  []Shaper
	Function ParallelFunction
	Start    float64
}

func (p *Parallel) Shape(x float64) float64 {
	v := p.Start

	for _, shaper := range p.Shapers {
		v = p.Function(shaper.Shape(x), v)
	}

	return v
}

func NewParallel(shapers ...Shaper) *Parallel {
	return &Parallel{
		Shapers:  shapers,
		Function: func(x float64, v float64) float64 { return x + v },
	}
}

func NewParallelF(start float64, function ParallelFunction, shapers ...Shaper) *Parallel {
	return &Parallel{
		Start:    start,
		Shapers:  shapers,
		Function: function,
	}
}

type Linear struct {
	Scale float64
	Shift float64
}

func (l *Linear) Shape(x float64) float64 {
	return x*l.Scale + l.Shift
}

func NewLinear(scale float64, shift float64) *Linear {
	return &Linear{
		Scale: scale,
		Shift: shift,
	}
}

func NewBipolar() *Linear {
	return &Linear{Scale: 2.0, Shift: -1.0}
}

func NewUnipolar() *Linear {
	return &Linear{Scale: 0.5, Shift: 0.5}
}

type Any struct {
	F ShapeFunction
}

func (a *Any) Shape(signal float64) float64 {
	return a.F(signal)
}

func NewAny(f ShapeFunction) *Any {
	return &Any{F: f}
}

func NewMod1() *Any {
	return &Any{F: func(signal float64) float64 { return math.Mod(signal, 1.0) }}
}

func NewAbs() *Any {
	return &Any{F: math.Abs}
}

func NewSin() *Any {
	return &Any{F: func(signal float64) float64 { return math.Sin(signal * 2.0 * math.Pi) }}
}

func NewTri() *Any {
	return &Any{F: func(x float64) float64 {
		if x < 0.5 {
			return 2.0 * x
		} else {
			return 2.0 - 2.0*x
		}
	}}
}

type Mod struct {
	M float64
}

func (m *Mod) Shape(signal float64) float64 {
	return math.Mod(signal, m.M)
}

func NewMod(m float64) *Mod {
	return &Mod{M: m}
}

type Mult struct {
	M float64
}

func (m *Mult) Shape(x float64) float64 {
	return m.M * x
}

func NewMult(m float64) *Mult {
	return &Mult{M: m}
}

type Add struct {
	A float64
}

func (a *Add) Shape(x float64) float64 {
	return x + a.A
}

func NewAdd(a float64) *Add {
	return &Add{A: a}
}

type Pulse struct {
	W float64
}

func (p *Pulse) Shape(signal float64) float64 {
	if signal < p.W {
		return 1.0
	}

	return 0.0
}

func NewPulse(w float64) *Pulse {
	return &Pulse{
		W: w,
	}
}

type Ripple struct {
	M float64
}

func (r *Ripple) Shape(signal float64) float64 {
	return signal + math.Mod(signal, r.M)
}

func NewRipple(m float64) *Ripple {
	return &Ripple{M: m}
}

func NewMinimoogVoyagerSawtooth() *Chain {
	return NewChain(
		NewLinear(0.25, 0.0),
		NewAny(func(s float64) float64 { return math.Sin(2.0 * math.Pi * s) }),
		NewBipolar(),
	)
}

func NewHardSync() *Chain {
	return NewChain(
		NewLinear(2.5, 0.0),
		NewMod1(),
		NewBipolar(),
	)
}

func (c *Chain) SetHardSyncA1(a1 float64) {
	c.Shapers[0].(*Linear).Scale = a1
}

func NewSoftSyncTriangle() *Chain {
	return NewChain(
		NewBipolar(),
		NewAbs(),
		NewLinear(1.25, 0.0),
		NewMod1(),
		NewTri(),
		NewBipolar(),
	)
}

func (c *Chain) SetSoftSyncA1(a1 float64) {
	c.Shapers[2].(*Linear).Scale = a1
}

func NewJP8000triMod() *Chain {
	return NewChain(
		NewBipolar(),
		NewAbs(),
		NewLinear(2.0, -1.0),
		NewMod1(),
		NewMult(0.3),
		NewAny(func(x float64) float64 { return 2.0 * (x - math.Ceil(x-0.5)) }),
	)
}

func (c *Chain) SetJP8000Mod(m float64) {
	c.Shapers[4].(*Mult).M = m
}

func NewPulseWidthMod() *Chain {
	return NewChain(
		NewLinear(1.25, 0.0),
		NewMod1(),
		NewPulse(0.4),
		NewBipolar(),
	)
}

func (c *Chain) SetPulseWidthA1(a1 float64) {
	c.Shapers[0].(*Linear).Scale = a1
}

func (c *Chain) SetPulseWidthW(w float64) {
	c.Shapers[2].(*Pulse).W = w
}

func NewSuperSaw() *Chain {
	return NewChain(
		NewLinear(1.5, 0.0),
		NewParallel(NewMod(0.25), NewMod(0.88)),
		NewAny(math.Sin),
		NewBipolar(),
	)
}

func (c *Chain) SetSuperSawA1(a1 float64) {
	c.Shapers[0].(*Linear).Scale = a1
}

func (c *Chain) SetSuperSawM1(m1 float64) {
	c.Shapers[1].(*Parallel).Shapers[0].(*Mod).M = m1
}

func (c *Chain) SetSuperSawM2(m2 float64) {
	c.Shapers[1].(*Parallel).Shapers[1].(*Mod).M = m2
}

// // sin(2.0 * PI * x(n){1 + g pulse [x(n) – 1, w]})
// func NewVarSlopeSin() *Chain {
// 	return NewChain(
// 		NewAdd(-1.0),
// 		NewPulse(0.5),
// 		NewAdd(1.0),
// 		NewSin(),
// 	)
// }

// func (c *Chain) SetVarSlopeSinW(w float64) {
// 	c.Shapers[1].(*Pulse).W = w
// }

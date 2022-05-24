package phaseshaping

import (
	"math"
)

type Shaper interface {
	Shape(float64) float64
}

type ShapeFunction func(float64) float64

type PhaseGenerator interface {
	SetPhase(float64)
	SetFrequency(float64, float64)
	Tick() float64
}

type Phasor struct {
	Phase float64
	Delta float64
}

func (ph *Phasor) SetPhase(phase float64) {
	ph.Phase = phase
}

func (ph *Phasor) SetFrequency(f float64, sr float64) {
	ph.Delta = f / sr
}

func (ph *Phasor) Tick() float64 {
	out := ph.Phase
	ph.Phase += ph.Delta
	ph.Phase = math.Mod(ph.Phase, 1.0)
	return out
}

type VarTri struct {
	Phasor
	W float64 // 0 <= - < 1
}

func (vtri *VarTri) SetPhase(phase float64) {
	vtri.Phase = phase
}

func (vtri *VarTri) SetFrequency(f float64, sr float64) {
	vtri.Delta = f / sr
}

func (vtri *VarTri) Tick() float64 {
	g1 := 2.0*vtri.Phase - 1.0
	g2 := vtri.Phase - vtri.W

	if g2 < 0.0 {
		g2 += 1.0
	}

	g2 = 2.0*g2 - 1.0

	vtri.Phase += vtri.Delta
	vtri.Phase = math.Mod(vtri.Phase, 1.0)

	return 1.0/(8.0*(vtri.W-vtri.W*vtri.W))*(g1*g1-g2*g2) + 0.5
}

type PhaseDistortion struct {
	Generator PhaseGenerator
	Shapers   []Shaper
}

func NewPhaseDistortion(generator PhaseGenerator) *PhaseDistortion {
	return &PhaseDistortion{
		Generator: generator,
		Shapers:   []Shaper{},
	}
}

func (pd *PhaseDistortion) Tick() float64 {
	out := pd.Generator.Tick()

	for _, shaper := range pd.Shapers {
		out = shaper.Shape(out)
	}

	return out
}

type LinearShape struct {
	Scale float64
	Shift float64
}

func NewLinear(scale float64, shift float64) *LinearShape {
	return &LinearShape{
		Scale: scale,
		Shift: shift,
	}
}

func NewBipolar() *LinearShape {
	return &LinearShape{Scale: 2.0, Shift: -1.0}
}

func NewUnipolar() *LinearShape {
	return &LinearShape{Scale: 0.5, Shift: 0.5}
}

func (l *LinearShape) Shape(x float64) float64 {
	return x*l.Scale + l.Shift
}

type FuncShape struct {
	f ShapeFunction
}

func (f *FuncShape) Shape(signal float64) float64 {
	return f.f(signal)
}

func NewFunction(f ShapeFunction) *FuncShape {
	return &FuncShape{f: f}
}

func NewMod1() *FuncShape {
	return &FuncShape{f: func(signal float64) float64 { return math.Mod(signal, 1.0) }}
}

func NewAbs() *FuncShape {
	return &FuncShape{f: math.Abs}
}

func NewTri() *FuncShape {
	return &FuncShape{f: func(x float64) float64 {
		if x < 0.5 {
			return 2.0 * x
		} else {
			return 2.0 - 2.0*x
		}
	}}
}

type ModM struct {
	M float64
}

func NewModM(m float64) *ModM {
	return &ModM{M: m}
}

func (m *ModM) Shape(signal float64) float64 {
	return math.Mod(signal, m.M)
}

type Mult struct {
	M float64
}

func NewMult(m float64) *Mult {
	return &Mult{M: m}
}

func (m *Mult) Shape(x float64) float64 {
	return m.M * x
}

type Pulse struct {
	W float64
}

func NewPulse(w float64) *Pulse {
	return &Pulse{
		W: w,
	}
}

func (p *Pulse) Shape(signal float64) float64 {
	if signal < p.W {
		return 1.0
	}

	return 0.0
}

type Ripple struct {
	M float64
}

func NewRipple(m float64) *Ripple {
	return &Ripple{M: m}
}

func (r *Ripple) Shape(signal float64) float64 {
	return signal + math.Mod(signal, r.M)
}

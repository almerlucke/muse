package waveshaping

import (
	"math"

	"github.com/almerlucke/muse/components/generator"
)

type Shaper interface {
	Shape(float64) float64
}

type GeneratorWrapper struct {
	generator generator.Generator
	shapers   []Shaper
	outVector []float64
}

func NewGeneratorWrapper(generator generator.Generator, shapers []Shaper) *GeneratorWrapper {
	return &GeneratorWrapper{
		generator: generator,
		shapers:   shapers,
		outVector: make([]float64, generator.NumDimensions()),
	}
}

func (gw *GeneratorWrapper) NumDimensions() int {
	return gw.generator.NumDimensions()
}

func (gw *GeneratorWrapper) Generate() []float64 {
	vec := gw.generator.Generate()
	for i, v := range vec {
		gw.outVector[i] = gw.shapers[i].Shape(v)
	}

	return gw.outVector
}

type Series struct {
	Shapers []Shaper
}

func (s *Series) Shape(x float64) float64 {
	for _, shaper := range s.Shapers {
		x = shaper.Shape(x)
	}

	return x
}

func NewSeries(shapers ...Shaper) *Series {
	return &Series{
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

// Make sure lookup table always contains one more sample so we do not need to wrap
type LookupTable []float64

func (lt LookupTable) Shape(x float64) float64 {
	nf := x * float64(len(lt)-1)
	n1 := int(nf)
	n2 := n1 + 1
	fr := nf - float64(n1)
	return lt[n1] + fr*(lt[n2]-lt[n1])
}

func NewSineTable(n int) LookupTable {
	table := make(LookupTable, n+1)
	phase := 0.0
	inc := 2.0 * math.Pi / float64(n)

	for i := 0; i < n; i++ {
		table[i] = math.Sin(phase)
		phase += inc
	}

	table[n] = table[0]

	return table
}

func NewNormalizedSineTable(n int) LookupTable {
	table := make(LookupTable, n+1)
	phase := 0.0
	inc := 2.0 * math.Pi / float64(n)

	for i := 0; i < n; i++ {
		table[i] = math.Sin(phase)*0.5 + 0.5
		phase += inc
	}

	table[n] = table[0]

	return table
}

func NewHanningWindow(n int) LookupTable {
	table := make(LookupTable, n+1)
	phase := 0.0
	inc := 2.0 * math.Pi / float64(n)

	for i := 0; i < n; i++ {
		table[i] = 0.5 - 0.5*math.Cos(phase)
		phase += inc
	}

	table[n] = table[0]

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

type Function struct {
	F func(float64) float64
}

func (f *Function) Shape(signal float64) float64 {
	return f.F(signal)
}

func NewFunction(f func(float64) float64) *Function {
	return &Function{F: f}
}

func NewMod1() *Function {
	return &Function{F: func(signal float64) float64 { return math.Mod(signal, 1.0) }}
}

func NewAbs() *Function {
	return &Function{F: math.Abs}
}

func NewSin() *Function {
	return &Function{F: func(signal float64) float64 { return math.Sin(signal * 2.0 * math.Pi) }}
}

func NewTri() *Function {
	return &Function{F: func(x float64) float64 {
		if x < 0.5 {
			return 2.0 * x
		} else {
			return 2.0 - 2.0*x
		}
	}}
}

type Mirror struct {
	Bottom float64
	Top    float64
}

func NewMirror(bottom float64, top float64) *Mirror {
	return &Mirror{
		Bottom: bottom,
		Top:    top,
	}
}

func (mirror *Mirror) Shape(x float64) float64 {
	for true {
		if x > mirror.Top {
			x = mirror.Top*2 - x
		} else if x < mirror.Bottom {
			x = mirror.Bottom*2 - x
		} else {
			break
		}
	}

	return x
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

type MinimoogVoyagerSawtooth struct {
	*Series
}

func NewMinimoogVoyagerSawtooth() *MinimoogVoyagerSawtooth {
	return &MinimoogVoyagerSawtooth{
		Series: NewSeries(
			NewLinear(0.25, 0.0),
			NewSin(),
			NewBipolar(),
		),
	}
}

type HardSync struct {
	*Series
}

func NewHardSync(a1 float64) *HardSync {
	return &HardSync{
		Series: NewSeries(
			NewLinear(a1, 0.0),
			NewMod1(),
			NewBipolar(),
		),
	}
}

func (hs *HardSync) SetA1(a1 float64) {
	hs.Shapers[0].(*Linear).Scale = a1
}

type SoftSyncTriangle struct {
	*Series
}

func NewSoftSyncTriangle(a1 float64) *SoftSyncTriangle {
	return &SoftSyncTriangle{
		Series: NewSeries(
			NewBipolar(),
			NewAbs(),
			NewLinear(a1, 0.0), // a=1.25
			NewMod1(),
			NewTri(),
			NewBipolar(),
		),
	}
}

func (sst *SoftSyncTriangle) SetA1(a1 float64) {
	sst.Shapers[2].(*Linear).Scale = a1
}

type JP8000triMod struct {
	*Series
}

func NewJP8000triMod(m float64) *JP8000triMod {
	return &JP8000triMod{
		Series: NewSeries(
			NewBipolar(),
			NewAbs(),
			NewLinear(2.0, -1.0),
			NewMod1(),
			NewMult(m), // m=0.3
			NewFunction(func(x float64) float64 { return 2.0 * (x - math.Ceil(x-0.5)) }),
		),
	}
}

func (jp *JP8000triMod) SetMod(m float64) {
	jp.Shapers[4].(*Mult).M = m
}

type PulseWidthMod struct {
	*Series
}

func NewPulseWidthMod() *PulseWidthMod {
	return &PulseWidthMod{
		Series: NewSeries(
			NewLinear(1.25, 0.0),
			NewMod1(),
			NewPulse(0.4),
			NewBipolar(),
		),
	}
}

func (pwm *PulseWidthMod) SetA1(a1 float64) {
	pwm.Shapers[0].(*Linear).Scale = a1
}

func (pwm *PulseWidthMod) SetWidth(w float64) {
	pwm.Shapers[2].(*Pulse).W = w
}

type SuperSaw struct {
	*Series
}

func NewSuperSaw(a1 float64, m1 float64, m2 float64) *SuperSaw {
	return &SuperSaw{
		Series: NewSeries(
			NewLinear(a1, 0.0),                  // a1=1.5
			NewParallel(NewMod(m1), NewMod(m2)), // m1=025, m2=0.88
			NewFunction(math.Sin),
			NewBipolar(),
		),
	}
}

func (ss *SuperSaw) SetA1(a1 float64) {
	ss.Shapers[0].(*Linear).Scale = a1
}

func (ss *SuperSaw) SetM1(m1 float64) {
	ss.Shapers[1].(*Parallel).Shapers[0].(*Mod).M = m1
}

func (ss *SuperSaw) SetM2(m2 float64) {
	ss.Shapers[1].(*Parallel).Shapers[1].(*Mod).M = m2
}

type Chebyshev struct {
	harmonics   map[int]float64
	maxHarmonic int
}

func NewChebyshev(harmonics map[int]float64) *Chebyshev {
	cheby := &Chebyshev{}

	cheby.SetHarmonics(harmonics)

	return cheby
}

func (cheby *Chebyshev) SetHarmonics(harmonics map[int]float64) {
	cheby.harmonics = harmonics

	maxHarmonic := 0

	for k, _ := range cheby.harmonics {
		if k > maxHarmonic {
			maxHarmonic = k
		}
	}

	cheby.maxHarmonic = maxHarmonic
}

func (cheby *Chebyshev) Shape(signal float64) float64 {
	var t0, t1, t2, mix float64

	t0 = 1
	t1 = signal

	for harmonic := 1; harmonic <= cheby.maxHarmonic; harmonic++ {
		if magnitude, ok := cheby.harmonics[harmonic]; ok {
			mix += magnitude * t1
		}
		t2 = 2.0*signal*t1 - t0
		t0 = t1
		t1 = t2
	}

	return mix
}

type Switch struct {
	Shapers []Shaper
	Index   int
}

func NewSwitch(index int, shapers ...Shaper) *Switch {
	return &Switch{
		Shapers: shapers,
		Index:   index,
	}
}

func (s *Switch) Shape(x float64) float64 {
	return s.Shapers[s.Index].Shape(x)
}

func (s *Switch) Selected() Shaper {
	return s.Shapers[s.Index]
}

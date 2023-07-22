package interpolator

import "github.com/almerlucke/muse/components/generator"

type Method int

const (
	Linear Method = iota
	Cubic
)

type interpolationHistory [4]float64

func (h interpolationHistory) interpolateLinear(t float64) float64 {
	return h[0] + (h[1]-h[0])*t
}

func (h interpolationHistory) interpolateCubic(t float64) float64 {
	t2 := t * t
	t3 := t * t2
	return (2*t3-3*t2+1)*h[1] + (t3-2*t2+t)*h[0] + (-2*t3+3*t2)*h[2] + (t3-t2)*h[3]
}

type Interpolator struct {
	generator     generator.Generator
	dt            float64
	t             float64
	method        Method
	numDimensions int
	history       []interpolationHistory
	outVector     []float64
}

func New(generator generator.Generator, method Method, dt float64) *Interpolator {
	numDimensions := generator.NumDimensions()

	interpol := &Interpolator{
		numDimensions: numDimensions,
		generator:     generator,
		dt:            dt,
		method:        method,
		history:       make([]interpolationHistory, numDimensions),
		outVector:     make([]float64, numDimensions),
	}

	for dim := 0; dim < numDimensions; dim++ {
		interpol.history[dim] = interpolationHistory{}
	}

	interpol.initialize()

	return interpol
}

func (inter *Interpolator) NumDimensions() int {
	return inter.generator.NumDimensions()
}

func (inter *Interpolator) SetDelta(dt float64) {
	inter.dt = dt
}

func (inter *Interpolator) initialize() {
	switch inter.method {
	case Linear:
		for dim, v := range inter.generator.Generate() {
			inter.history[dim][0] = v
		}
		for dim, v := range inter.generator.Generate() {
			inter.history[dim][1] = v
		}
	case Cubic:
		for dim, v := range inter.generator.Generate() {
			inter.history[dim][0] = v
			inter.history[dim][1] = v
		}
		for dim, v := range inter.generator.Generate() {
			inter.history[dim][2] = v
		}
		for dim, v := range inter.generator.Generate() {
			inter.history[dim][3] = v
		}
	}
}

func (inter *Interpolator) updateHistory() {
	switch inter.method {
	case Linear:
		for dim, v := range inter.generator.Generate() {
			inter.history[dim][0] = inter.history[dim][1]
			inter.history[dim][1] = v
		}
	case Cubic:
		for dim, v := range inter.generator.Generate() {
			inter.history[dim][0] = inter.history[dim][1]
			inter.history[dim][1] = inter.history[dim][2]
			inter.history[dim][2] = inter.history[dim][3]
			inter.history[dim][3] = v
		}
	}
}

func (inter *Interpolator) interpolate(t float64) []float64 {
	switch inter.method {
	case Linear:
		for dim := 0; dim < inter.numDimensions; dim++ {
			inter.outVector[dim] = inter.history[dim].interpolateLinear(t)
		}
	case Cubic:
		for dim := 0; dim < inter.numDimensions; dim++ {
			inter.outVector[dim] = inter.history[dim].interpolateCubic(t)
		}
	}

	return inter.outVector
}

func (inter *Interpolator) Generate() []float64 {
	out := inter.interpolate(inter.t)

	inter.t += inter.dt

	if inter.t >= 1.0 {
		inter.t -= 1.0
		inter.updateHistory()
	}

	return out
}

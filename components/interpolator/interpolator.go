package interpolator

import "github.com/almerlucke/muse/components/generator"

type Method int

const (
	Linear Method = iota
	Cubic
)

type Interpolator struct {
	generator    generator.Generator
	numCycles    int
	currentCycle int
	method       Method
	history      [4]float64
}

func NewInterpolator(generator generator.Generator, method Method, numCycles int) *Interpolator {
	interpol := &Interpolator{
		generator: generator,
		numCycles: numCycles,
		method:    method,
	}

	interpol.initialize()

	return interpol
}

func (inter *Interpolator) SetNumCycles(n int) {
	inter.numCycles = n
	if inter.currentCycle >= n {
		inter.currentCycle = 0
	}
}

func (inter *Interpolator) initialize() {
	switch inter.method {
	case Linear:
		inter.history[0] = inter.generator.Tick()
		inter.history[1] = inter.generator.Tick()
	case Cubic:
		inter.history[1] = inter.generator.Tick()
		inter.history[2] = inter.generator.Tick()
		inter.history[3] = inter.generator.Tick()
		inter.history[0] = inter.history[1]
	}
}

func (inter *Interpolator) updateHistory() {
	switch inter.method {
	case Linear:
		inter.history[0] = inter.history[1]
		inter.history[1] = inter.generator.Tick()
	case Cubic:
		inter.history[0] = inter.history[1]
		inter.history[1] = inter.history[2]
		inter.history[2] = inter.history[3]
		inter.history[3] = inter.generator.Tick()
	}
}

func (inter *Interpolator) interpolate(t float64) float64 {
	out := 0.0
	switch inter.method {
	case Linear:
		out = inter.history[0] + (inter.history[1]-inter.history[0])*t
	case Cubic:
		t2 := t * t
		t3 := t * t2
		out = (2*t3-3*t2+1)*inter.history[1] + (t3-2*t2+t)*inter.history[0] + (-2*t3+3*t2)*inter.history[2] + (t3-t2)*inter.history[3]
	}

	return out
}

func (inter *Interpolator) Tick() float64 {
	if inter.currentCycle >= inter.numCycles {
		inter.currentCycle = 0
		inter.updateHistory()
	}

	t := float64(inter.currentCycle) / float64(inter.numCycles)
	inter.currentCycle++

	return inter.interpolate(t)
}

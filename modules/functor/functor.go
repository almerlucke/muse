package functor

import "github.com/almerlucke/muse"

type FunctorFunction func([]float64) float64

type Functor struct {
	*muse.BaseModule
	f     FunctorFunction
	inVec []float64
}

func NewFunctor(numInputs int, f FunctorFunction, config *muse.Configuration) *Functor {
	return &Functor{
		BaseModule: muse.NewBaseModule(numInputs, 1, config, ""),
		f:          f,
		inVec:      make([]float64, numInputs),
	}
}

func NewMult(numInputs int, config *muse.Configuration) *Functor {
	return NewFunctor(numInputs, FunctorMult, config)
}

func NewScale(scale float64, offset float64, config *muse.Configuration) *Functor {
	return NewFunctor(1, FunctorScale(scale, offset), config)
}

func NewAmp(amp float64, config *muse.Configuration) *Functor {
	return NewFunctor(1, FunctorScale(amp, 0), config)
}

func NewBetween(min float64, max float64, config *muse.Configuration) *Functor {
	return NewFunctor(1, FunctorBetween(min, max), config)
}

func FunctorMult(vec []float64) float64 {
	if len(vec) == 0 {
		return 0
	}

	mult := 1.0

	for _, v := range vec {
		mult *= v
	}

	return mult
}

func FunctorScale(scale float64, offset float64) FunctorFunction {
	return func(v []float64) float64 {
		return v[0]*scale + offset
	}
}

func FunctorBetween(min float64, max float64) FunctorFunction {
	if min > max {
		tmp := max
		max = min
		min = tmp
	}
	return func(v []float64) float64 {
		return v[0]*(max-min) + min
	}
}

func (f *Functor) Synthesize() bool {
	if !f.BaseModule.Synthesize() {
		return false
	}

	out := f.Outputs[0].Buffer

	for i := 0; i < f.Config.BufferSize; i++ {
		for ii, input := range f.Inputs {
			f.inVec[ii] = input.Buffer[i]
		}

		out[i] = f.f(f.inVec)
	}

	return true
}

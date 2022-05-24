package functor

import "github.com/almerlucke/muse"

type FunctorFunction func([]float64) float64

type Functor struct {
	*muse.BaseModule
	f     FunctorFunction
	inVec []float64
}

func NewFunctor(numInputs int, f FunctorFunction, config *muse.Configuration, identifier string) *Functor {
	return &Functor{
		BaseModule: muse.NewBaseModule(numInputs, 1, config, identifier),
		f:          f,
		inVec:      make([]float64, numInputs),
	}
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

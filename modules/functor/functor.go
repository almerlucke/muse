package functor

import "github.com/almerlucke/muse"

type Function func([]float64) float64

type Functor struct {
	*muse.BaseModule
	f     Function
	inVec []float64
}

func New(numInputs int, f Function) *Functor {
	fctr := &Functor{
		BaseModule: muse.NewBaseModule(numInputs, 1),
		f:          f,
		inVec:      make([]float64, numInputs),
	}

	fctr.SetSelf(fctr)

	return fctr
}

func NewMult(numInputs int) *Functor {
	return New(numInputs, Mult)
}

func NewScale(scale float64, offset float64) *Functor {
	return New(1, Scale(scale, offset))
}

func NewAmp(amp float64) *Functor {
	return New(1, Scale(amp, 0))
}

func NewBetween(min float64, max float64) *Functor {
	return New(1, Between(min, max))
}

func Mult(vec []float64) float64 {
	if len(vec) == 0 {
		return 0
	}

	mult := 1.0

	for _, v := range vec {
		mult *= v
	}

	return mult
}

func Scale(scale float64, offset float64) Function {
	return func(v []float64) float64 {
		return v[0]*scale + offset
	}
}

func Between(min float64, max float64) Function {
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

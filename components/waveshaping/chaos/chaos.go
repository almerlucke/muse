package chaos

import "math"

type VerhulstFunc func(float64) float64

type Verhulst struct {
	A float64
	F VerhulstFunc
}

func NewVerhulst(a float64) *Verhulst {
	return &Verhulst{A: a, F: Iter1}
}

func NewVerhulstWithFunc(a float64, f VerhulstFunc) *Verhulst {
	return &Verhulst{A: a, F: f}
}

func (v *Verhulst) Shape(x float64) float64 {
	return v.A * v.F(x)
}

func Iter1(x float64) float64 {
	return x * (1.0 - x)
}

func Iter1a(x float64) float64 {
	return x * (1.0 - x*x)
}

func Iter1b(x float64) float64 {
	return x * math.Sin(x)
}

func Iter1c(x float64) float64 {
	return x * math.Exp(x)
}

func Iter1d(x float64) float64 {
	t := 1.0 + x*x
	return x / (t * t)
}

func Iter1e(x float64) float64 {
	return x * math.Sqrt(1.0-x)
}

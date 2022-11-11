package chaos

type Verhulst struct {
	A float64
}

func NewVerhulst(a float64) *Verhulst {
	return &Verhulst{A: a}
}

func (v *Verhulst) Shape(x float64) float64 {
	return v.A * x * (1.0 - x)
}

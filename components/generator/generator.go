package generator

type Generator interface {
	NumDimensions() int
	Generate() []float64
}

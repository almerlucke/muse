package generator

type Generator interface {
	NumDimensions() int
	Tick() []float64
}

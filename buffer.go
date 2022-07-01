package muse

type Buffer []float64

func (b Buffer) Clear() {
	for i := 0; i < len(b); i++ {
		b[i] = 0
	}
}

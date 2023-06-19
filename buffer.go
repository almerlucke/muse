package muse

type Buffer []float64

func (b Buffer) Clear() {
	for i := 0; i < len(b); i++ {
		b[i] = 0
	}
}

func (b Buffer) Lookup(pos float64) float64 {
	integer1 := int(pos)
	fraction := pos - float64(integer1)
	integer2 := (integer1 + 1) % len(b)
	sample1 := b[integer1]
	sample2 := b[integer2]
	return sample1 + (sample2-sample1)*fraction
}

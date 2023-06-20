package buffer

type Buffer []float64

func (b Buffer) Clear() {
	for i := 0; i < len(b); i++ {
		b[i] = 0
	}
}

func (b Buffer) Lookup(pos float64, wrap bool) float64 {
	l := len(b)
	integer1 := int(pos)
	fraction := pos - float64(integer1)

	var integer2 int

	if wrap {
		integer2 = (integer1 + 1) % l
	} else {
		integer2 = integer1 + 1
		if integer2 >= l {
			integer2 = l - 1
		}
	}

	sample1 := b[integer1]
	sample2 := b[integer2]
	return sample1 + (sample2-sample1)*fraction
}

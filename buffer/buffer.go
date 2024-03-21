package buffer

type Buffer []float64

func (b Buffer) Clear() {
	for i := 0; i < len(b); i++ {
		b[i] = 0
	}
}

func (b Buffer) Lookup(pos float64, wrap bool) float64 {
	l := len(b)
	i1 := int(pos)
	i2 := i1 + 1
	fr := pos - float64(i1)

	if wrap {
		i2 = i2 % l
	} else {
		if i2 >= l {
			i2 = l - 1
		}
	}

	s1 := b[i1]

	return s1 + (b[i2]-s1)*fr
}

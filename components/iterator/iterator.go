package iterator

type Updater interface {
	Update([]float64)
}

type Iterator struct {
	updater   Updater
	values    []float64
	outVector []float64
}

func NewIterator(initialValues []float64, updater Updater) *Iterator {
	iter := &Iterator{
		values:    make([]float64, len(initialValues)),
		updater:   updater,
		outVector: make([]float64, len(initialValues)),
	}

	copy(iter.values, initialValues)

	return iter
}

func (iter *Iterator) NumDimensions() int {
	return len(iter.values)
}

func (iter *Iterator) SetValues(vs []float64) {
	copy(iter.values, vs)
}

func (iter *Iterator) Tick() []float64 {
	copy(iter.outVector, iter.values)
	iter.updater.Update(iter.values)
	return iter.outVector
}

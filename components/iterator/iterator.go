package iterator

import "github.com/almerlucke/muse/components/waveshaping"

type Iterator struct {
	Shaper waveshaping.Shaper
	Value  float64
}

func NewIterator(shaper waveshaping.Shaper, initValue float64) *Iterator {
	return &Iterator{
		Shaper: shaper,
		Value:  initValue,
	}
}

func (iter *Iterator) SetValue(v float64) {
	iter.Value = v
}

func (iter *Iterator) Tick() float64 {
	v := iter.Value
	iter.Value = iter.Shaper.Shape(v)
	return v
}

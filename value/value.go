package value

import (
	"math/rand"
)

type Valuer[T any] interface {
	Value() T
	Continuous() bool
	Reset()
	Finished() bool
}

type Transformer[T any] interface {
	Transform(T) T
}

func Generate[T any](val Valuer[T], n int) []T {
	sequence := []T{}

	for i := 0; i < n; i++ {
		sequence = append(sequence, val.Value())
		if val.Finished() {
			val.Reset()
		}
	}

	return sequence
}

type Const[T any] struct {
	value T
}

func NewConst[T any](c T) *Const[T] {
	return &Const[T]{value: c}
}

func (c *Const[T]) Get() T {
	return c.value
}

func (c *Const[T]) Set(v T) {
	c.value = v
}

func (c *Const[T]) Value() T {
	return c.value
}

func (c *Const[T]) Continuous() bool {
	return true
}

func (c *Const[T]) Reset() {}

func (c *Const[T]) Finished() bool {
	return false
}

type Sequence[T any] struct {
	values     []T
	index      int
	continuous bool
}

func NewSequence[T any](values []T) *Sequence[T] {
	return &Sequence[T]{values: values, continuous: true}
}

func NewSequenceNC[T any](values []T) *Sequence[T] {
	return &Sequence[T]{values: values, continuous: false}
}

func (s *Sequence[T]) Get() []T {
	return s.values
}

func (s *Sequence[T]) Set(v []T) {
	s.values = v
	if s.index >= len(v) {
		s.index = 0
	}
}

func (s *Sequence[T]) Value() T {
	var v T

	if s.index < len(s.values) {
		v = s.values[s.index]
		s.index++
		if s.index == len(s.values) && s.continuous {
			s.index = 0
		}
	} else if s.continuous {
		s.index = 0
		return s.Value()
	} else {
		v = s.values[s.index-1]
	}

	return v
}

func (s *Sequence[T]) Continuous() bool {
	return s.continuous
}

func (s *Sequence[T]) Reset() {
	s.index = 0
}

func (s *Sequence[T]) Finished() bool {
	return !s.continuous && s.index == len(s.values)
}

func (s *Sequence[T]) Randomize() {
	rand.Shuffle(len(s.values), func(i, j int) {
		s.values[i], s.values[j] = s.values[j], s.values[i]
	})
}

type Function[T any] struct {
	f func() T
}

func NewFunction[T any](f func() T) *Function[T] {
	return &Function[T]{f: f}
}

func (f *Function[T]) Value() T {
	return f.f()
}

func (f *Function[T]) Continuous() bool {
	return true
}

func (f *Function[T]) Reset() {}

func (f *Function[T]) Finished() bool {
	return false
}

// Repeat can make a continous generator non-continous by only repeating Next() n times,
// or can extend a non-continuous generator by repeating it n times
type Repeat[T any] struct {
	value Valuer[T]
	min   int
	max   int
	n     int
}

func NewRepeat[T any](value Valuer[T], min int, max int) *Repeat[T] {
	return &Repeat[T]{
		value: value,
		min:   min,
		max:   max,
		n:     min + rand.Intn((max-min)+1),
	}
}

func (r *Repeat[T]) Value() T {
	var v T

	if r.n > 0 {
		v = r.value.Value()
		if r.value.Continuous() {
			r.n--
		} else if r.value.Finished() {
			r.n--
			if r.n > 0 {
				r.value.Reset()
			}
		}
	}

	return v
}

func (r *Repeat[T]) Continuous() bool {
	return false
}

func (r *Repeat[T]) Reset() {
	r.value.Reset()
	r.n = r.min + rand.Intn((r.max-r.min)+1)
}

func (r *Repeat[T]) Finished() bool {
	return r.n == 0
}

// And is a meta-sequence, a sequence of Generators
type And[T any] struct {
	values     []Valuer[T]
	current    Valuer[T]
	index      int
	continuous bool
}

func NewAnd[T any](values []Valuer[T], continuous bool) *And[T] {
	return &And[T]{
		values:     values,
		continuous: continuous,
	}
}

func (a *And[T]) Value() T {
	var v T

	if a.index == len(a.values) {
		return v
	}

	if a.current == nil {
		a.current = a.values[a.index]
	}

	if a.current.Continuous() {
		v = a.current.Value()
		a.index++
		if a.index == len(a.values) && a.continuous {
			a.Reset()
		}
		if a.index != len(a.values) {
			a.current = a.values[a.index]
		}
	} else if !a.current.Finished() {
		v = a.current.Value()
		if a.current.Finished() {
			a.index++
			if a.index == len(a.values) && a.continuous {
				a.Reset()
			}
			if a.index != len(a.values) {
				a.current = a.values[a.index]
			}
		}
	}

	return v
}

func (a *And[T]) Continuous() bool {
	return a.continuous
}

func (a *And[T]) Reset() {
	for _, v := range a.values {
		v.Reset()
	}

	a.index = 0
	a.current = nil
}

func (a *And[T]) Finished() bool {
	return !a.continuous && a.index == len(a.values)
}

// Or chooses one of the generators each cycle
type Or[T any] struct {
	values     []Valuer[T]
	current    Valuer[T]
	index      int
	finished   bool
	continuous bool
}

func NewOr[T any](values []Valuer[T], continuous bool) *Or[T] {
	return &Or[T]{
		values:     values,
		continuous: continuous,
	}
}

func (o *Or[T]) Value() T {
	var v T

	if o.current == nil {
		o.index = rand.Intn(len(o.values))
		o.current = o.values[o.index]
	}

	if o.current.Continuous() {
		v = o.current.Value()
		if o.continuous {
			o.index = rand.Intn(len(o.values))
			o.current = o.values[o.index]
		} else {
			o.finished = true
		}
	} else {
		v = o.current.Value()
		if o.current.Finished() {
			if o.continuous {
				o.current.Reset()
				o.index = rand.Intn(len(o.values))
				o.current = o.values[o.index]
			} else {
				o.finished = true
			}
		}
	}

	return v
}

func (o *Or[T]) Continuous() bool {
	return o.continuous
}

func (o *Or[T]) Reset() {
	for _, v := range o.values {
		v.Reset()
	}

	o.current = nil
	o.finished = false
}

func (o *Or[T]) Finished() bool {
	return !o.continuous && o.finished
}

// Flatten flattens a Valuer that generates slices of a certain type
type Flatten[T any] struct {
	sliceGenerator Valuer[[]T]
	currentSlice   []T
	currentIndex   int
}

func NewFlatten[T any](sliceGen Valuer[[]T]) *Flatten[T] {
	return &Flatten[T]{sliceGenerator: sliceGen}
}

func (f *Flatten[T]) Value() T {
	var null T

	if f.Finished() {
		return null
	}

	if f.currentSlice == nil || f.currentIndex >= len(f.currentSlice) {
		f.currentSlice = f.sliceGenerator.Value()
		f.currentIndex = 0
	}

	v := f.currentSlice[f.currentIndex]

	f.currentIndex++

	return v
}

func (f *Flatten[T]) Continuous() bool {
	return f.sliceGenerator.Continuous()
}

func (f *Flatten[T]) Reset() {
	f.sliceGenerator.Reset()
	f.currentSlice = nil
}

func (f *Flatten[T]) Finished() bool {
	return f.sliceGenerator.Finished() && (f.currentSlice == nil || f.currentIndex >= len(f.currentSlice))
}

// Transform transforms the output of a value with a transformer
type Transform[T any] struct {
	value       Valuer[T]
	transformer Transformer[T]
}

func NewTransform[T any](v Valuer[T], t Transformer[T]) *Transform[T] {
	return &Transform[T]{
		value:       v,
		transformer: t,
	}
}

func (t *Transform[T]) Value() T {
	return t.transformer.Transform(t.value.Value())
}

func (t *Transform[T]) Continuous() bool {
	return t.value.Continuous()
}

func (t *Transform[T]) Reset() {
	t.value.Reset()
}

func (t *Transform[T]) Finished() bool {
	return t.value.Finished()
}

type TFunc[T any] func(T) T

func (f TFunc[T]) Transform(v T) T {
	return f(v)
}

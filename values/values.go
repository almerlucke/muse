package values

import (
	"math/rand"
)

type Transformer[T any] interface {
	Transform(T) T
}

type Generator[T any] interface {
	Next() T
	Continuous() bool
	Reset()
	Finished() bool
}

func Generate[T any](generator Generator[T], n int) []T {
	sequence := []T{}

	for i := 0; i < n; i++ {
		sequence = append(sequence, generator.Next())
		if generator.Finished() {
			generator.Reset()
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

func (c *Const[T]) Next() T {
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

func NewSequence[T any](values []T, continuous bool) *Sequence[T] {
	return &Sequence[T]{values: values, continuous: continuous}
}

func (s *Sequence[T]) ChangeValues(values []T) {
	s.values = values
	if s.index >= len(values) {
		s.index = 0
	}
}

func (s *Sequence[T]) Next() T {
	var v T

	if s.index < len(s.values) {
		v = s.values[s.index]
		s.index++
		if s.index == len(s.values) && s.continuous {
			s.index = 0
		}
	} else if s.continuous {
		s.index = 0
		return s.Next()
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

type Function[T any] struct {
	f func() T
}

func NewFunction[T any](f func() T) *Function[T] {
	return &Function[T]{f: f}
}

func (f *Function[T]) Next() T {
	return f.f()
}

func (f *Function[T]) Continuous() bool {
	return true
}

func (f *Function[T]) Reset() {}

func (f *Function[T]) Finished() bool {
	return false
}

type Ramp struct {
	current    float64
	increment  float64
	continuous bool
}

func NewRamp(increment float64, continuous bool) *Ramp {
	return &Ramp{
		current:    0,
		increment:  increment,
		continuous: continuous,
	}
}

func (r *Ramp) Next() float64 {
	if r.current >= 1.0 {
		return 0
	}
	v := r.current
	r.current += r.increment
	if r.current >= 1.0 {
		if r.continuous {
			r.current -= 1.0
		}
	}
	return v
}

func (r *Ramp) Continuous() bool {
	return r.continuous
}

func (r *Ramp) Reset() {
	r.current = 0
}

func (r *Ramp) Finished() bool {
	return !r.continuous && r.current >= 1.0
}

// Repeat can make a continous generator non-continous by only repeating Next() n times,
// or can extend a non-continuos generator by repeating it n times
type Repeat[T any] struct {
	generator Generator[T]
	min       int
	max       int
	n         int
}

func NewRepeat[T any](generator Generator[T], min int, max int) *Repeat[T] {
	return &Repeat[T]{
		generator: generator,
		min:       min,
		max:       max,
		n:         min + rand.Intn((max-min)+1),
	}
}

func (r *Repeat[T]) Next() T {
	var v T

	if r.n > 0 {
		v = r.generator.Next()
		if r.generator.Continuous() {
			r.n--
		} else if r.generator.Finished() {
			r.n--
			if r.n > 0 {
				r.generator.Reset()
			}
		}
	}

	return v
}

func (r *Repeat[T]) Continuous() bool {
	return false
}

func (r *Repeat[T]) Reset() {
	r.generator.Reset()
	r.n = r.min + rand.Intn((r.max-r.min)+1)
}

func (r *Repeat[T]) Finished() bool {
	return r.n == 0
}

// And is a meta-sequence, a sequence of Generators
type And[T any] struct {
	generators []Generator[T]
	current    Generator[T]
	index      int
	continous  bool
}

func NewAnd[T any](generators []Generator[T], continous bool) *And[T] {
	return &And[T]{
		generators: generators,
		continous:  continous,
	}
}

func (a *And[T]) Next() T {
	var v T

	if a.index == len(a.generators) {
		return v
	}

	if a.current == nil {
		a.current = a.generators[a.index]
	}

	if a.current.Continuous() {
		v = a.current.Next()
		a.index++
		if a.index == len(a.generators) && a.continous {
			a.Reset()
		}
		if a.index != len(a.generators) {
			a.current = a.generators[a.index]
		}
	} else if !a.current.Finished() {
		v = a.current.Next()
		if a.current.Finished() {
			a.index++
			if a.index == len(a.generators) && a.continous {
				a.Reset()
			}
			if a.index != len(a.generators) {
				a.current = a.generators[a.index]
			}
		}
	}

	return v
}

func (a *And[T]) Continuous() bool {
	return a.continous
}

func (a *And[T]) Reset() {
	for _, v := range a.generators {
		v.Reset()
	}

	a.index = 0
	a.current = nil
}

func (a *And[T]) Finished() bool {
	return !a.continous && a.index == len(a.generators)
}

// Or chooses one of the generators each cycle
type Or[T any] struct {
	generators []Generator[T]
	current    Generator[T]
	finished   bool
	continuous bool
}

func NewOr[T any](generators []Generator[T], continuous bool) *Or[T] {
	return &Or[T]{
		generators: generators,
		continuous: continuous,
	}
}

func (o *Or[T]) Next() T {
	var v T

	if o.current == nil {
		o.current = o.generators[rand.Intn(len(o.generators))]
	}

	if o.current.Continuous() {
		v = o.current.Next()
		if o.continuous {
			o.current = o.generators[rand.Intn(len(o.generators))]
		} else {
			o.finished = true
		}
	} else {
		v = o.current.Next()
		if o.current.Finished() {
			if o.continuous {
				o.current.Reset()
				o.current = o.generators[rand.Intn(len(o.generators))]
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
	for _, v := range o.generators {
		v.Reset()
	}

	o.current = nil
	o.finished = false
}

func (o *Or[T]) Finished() bool {
	return !o.continuous && o.finished
}

// Transform transforms the output of a generator with a transformer
type Transform[T any] struct {
	generator   Generator[T]
	transformer Transformer[T]
}

func (t *Transform[T]) Next() T {
	return t.transformer.Transform(t.generator.Next())
}

func (t *Transform[T]) Continuous() bool {
	return t.generator.Continuous()
}

func (t *Transform[T]) Reset() {
	t.generator.Reset()
}

func (t *Transform[T]) Finished() bool {
	return t.generator.Finished()
}

type Map map[string]any
type MapPrototype map[string]any

func (p MapPrototype) Map() Map {
	m := Map{}

	for k, v := range p {
		if sub, ok := v.(MapPrototype); ok {
			m[k] = sub.Map()
		} else {
			m[k] = v.(Generator[any]).Next()
		}
	}

	return m
}

func (p MapPrototype) MapRaw() map[string]any {
	m := map[string]any{}

	for k, v := range p {
		if sub, ok := v.(MapPrototype); ok {
			m[k] = sub.MapRaw()
		} else {
			m[k] = v.(Generator[any]).Next()
		}
	}

	return m
}

func (m Map) MR(key string) map[string]any {
	if sub, ok := m[key].(map[string]any); ok {
		return sub
	}

	return nil
}

func (m Map) M(key string) Map {
	if sub, ok := m[key].(Map); ok {
		return sub
	}

	return nil
}

func (m Map) S(key string) string {
	if value, ok := m[key].(string); ok {
		return value
	}

	return ""
}

func (m Map) F(key string) float64 {
	if value, ok := m[key].(float64); ok {
		return value
	}

	return 0
}

func (m Map) I(key string) int64 {
	if value, ok := m[key].(int64); ok {
		return value
	}

	return 0
}

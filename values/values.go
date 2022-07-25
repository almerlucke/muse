package values

import (
	"log"
	"math/rand"
)

type Generator[T any] interface {
	Next() T
	Continuous() bool
	Reset()
	Finished() bool
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
	values    []T
	index     int
	continous bool
}

func NewSequence[T any](values []T, continuous bool) *Sequence[T] {
	return &Sequence[T]{values: values, continous: continuous}
}

func (s *Sequence[T]) Next() T {
	var v T

	if s.index < len(s.values) {
		v = s.values[s.index]
		s.index++
		if s.index == len(s.values) && s.continous {
			s.index = 0
		}
	} else {
		v = s.values[s.index-1]
	}

	return v
}

func (s *Sequence[T]) Continuous() bool {
	return s.continous
}

func (s *Sequence[T]) Reset() {
	s.index = 0
}

func (s *Sequence[T]) Finished() bool {
	return !s.continous && s.index == len(s.values)
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

// Repeat can make a continous generator non-continous by only repeating Next() n times,
// or can extend a non-continuos generator by repeating it n times
type Repeat[T any] struct {
	generator Generator[T]
	min       int
	max       int
	n         int
}

func NewRepeat[T any](generator Generator[T], min int, max int) *Repeat[T] {
	n := min + rand.Intn((max-min)+1)
	log.Printf("n %d", n)
	return &Repeat[T]{
		generator: generator,
		min:       min,
		max:       max,
		n:         n,
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

type Pattern[T any] struct {
	generators []Generator[T]
	current    Generator[T]
	index      int
	continous  bool
}

func NewPattern[T any](generators []Generator[T], continous bool) *Pattern[T] {
	return &Pattern[T]{
		generators: generators,
		current:    nil,
		continous:  continous,
	}
}

func (p *Pattern[T]) Next() T {
	var v T

	if p.index == len(p.generators) {
		return v
	}

	if p.current == nil {
		p.current = p.generators[p.index]
	}

	if p.current.Continuous() {
		v = p.current.Next()
		p.index++
		if p.index == len(p.generators) && p.continous {
			p.Reset()
		}
		if p.index != len(p.generators) {
			p.current = p.generators[p.index]
		}
	} else if !p.current.Finished() {
		v = p.current.Next()
		if p.current.Finished() {
			p.index++
			if p.index == len(p.generators) && p.continous {
				p.Reset()
			}
			if p.index != len(p.generators) {
				p.current = p.generators[p.index]
			}
		}
	}

	return v
}

func (p *Pattern[T]) Continuous() bool {
	return p.continous
}

func (p *Pattern[T]) Reset() {
	for _, v := range p.generators {
		v.Reset()
	}

	p.index = 0
	p.current = nil
}

func (p *Pattern[T]) Finished() bool {
	return !p.continous && p.index == len(p.generators)
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

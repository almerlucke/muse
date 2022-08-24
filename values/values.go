package values

import (
	"math/rand"

	"github.com/almerlucke/muse"
)

type Transformer[T any] interface {
	Transform(T) T
}

type Generator[T any] interface {
	muse.Stater
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

func (c *Const[T]) GetState() map[string]any {
	if _, ok := any(c.value).(muse.Stater); ok {
		return map[string]any{"value": any(c.value).(muse.Stater).GetState()}
	}

	return map[string]any{"value": c.value}
}

func (c *Const[T]) SetState(state map[string]any) {
	if _, ok := any(c.value).(muse.Stater); ok {
		any(c.value).(muse.Stater).SetState(state["value"].(map[string]any))
	} else {
		c.value = state["value"].(T)
	}
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

func (s *Sequence[T]) GetState() map[string]any {
	return map[string]any{"values": s.values, "index": s.index, "continuous": s.continuous}
}

func (s *Sequence[T]) SetState(state map[string]any) {
	s.values = state["values"].([]T)
	s.index = state["index"].(int)
	s.continuous = state["continuous"].(bool)
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

func (f *Function[T]) GetState() map[string]any {
	return map[string]any{}
}

func (f *Function[T]) SetState(state map[string]any) {
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

func (r *Ramp) GetState() map[string]any {
	return map[string]any{"current": r.current, "increment": r.increment, "continuous": r.continuous}
}

func (r *Ramp) SetState(state map[string]any) {
	r.current = state["current"].(float64)
	r.increment = state["increment"].(float64)
	r.continuous = state["continuous"].(bool)
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

func (r *Repeat[T]) GetState() map[string]any {
	return map[string]any{"generator": r.generator.GetState(), "min": r.min, "max": r.max, "n": r.n}
}

func (r *Repeat[T]) SetState(state map[string]any) {
	r.generator.SetState(state["generator"].(map[string]any))
	r.min = state["min"].(int)
	r.max = state["max"].(int)
	r.n = state["n"].(int)
}

// And is a meta-sequence, a sequence of Generators
type And[T any] struct {
	generators []Generator[T]
	current    Generator[T]
	index      int
	continuous bool
}

func NewAnd[T any](generators []Generator[T], continuous bool) *And[T] {
	return &And[T]{
		generators: generators,
		continuous: continuous,
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
		if a.index == len(a.generators) && a.continuous {
			a.Reset()
		}
		if a.index != len(a.generators) {
			a.current = a.generators[a.index]
		}
	} else if !a.current.Finished() {
		v = a.current.Next()
		if a.current.Finished() {
			a.index++
			if a.index == len(a.generators) && a.continuous {
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
	return a.continuous
}

func (a *And[T]) Reset() {
	for _, v := range a.generators {
		v.Reset()
	}

	a.index = 0
	a.current = nil
}

func (a *And[T]) Finished() bool {
	return !a.continuous && a.index == len(a.generators)
}

func (a *And[T]) GetState() map[string]any {
	states := []map[string]any{}
	for index, generator := range a.generators {
		states[index] = generator.GetState()
	}

	return map[string]any{"generators": states, "isCurrentSet": a.current != nil, "index": a.index, "continuous": a.continuous}
}

func (a *And[T]) SetState(state map[string]any) {
	for index, state := range state["generators"].([]map[string]any) {
		a.generators[index].SetState(state)
	}

	a.continuous = state["continuous"].(bool)
	a.index = state["index"].(int)
	a.current = nil

	if state["isCurrentSet"].(bool) {
		a.current = a.generators[a.index]
	}
}

// Or chooses one of the generators each cycle
type Or[T any] struct {
	generators []Generator[T]
	current    Generator[T]
	index      int
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
		o.index = rand.Intn(len(o.generators))
		o.current = o.generators[o.index]
	}

	if o.current.Continuous() {
		v = o.current.Next()
		if o.continuous {
			o.index = rand.Intn(len(o.generators))
			o.current = o.generators[o.index]
		} else {
			o.finished = true
		}
	} else {
		v = o.current.Next()
		if o.current.Finished() {
			if o.continuous {
				o.current.Reset()
				o.index = rand.Intn(len(o.generators))
				o.current = o.generators[o.index]
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

func (o *Or[T]) GetState() map[string]any {
	states := []map[string]any{}
	for index, generator := range o.generators {
		states[index] = generator.GetState()
	}

	return map[string]any{"generators": states, "isCurrentSet": o.current != nil, "index": o.index, "continuous": o.continuous, "finished": o.finished}
}

func (o *Or[T]) SetState(state map[string]any) {
	for index, state := range state["generators"].([]map[string]any) {
		o.generators[index].SetState(state)
	}

	o.index = state["index"].(int)
	o.continuous = state["continuous"].(bool)
	o.finished = state["finished"].(bool)

	if state["isCurrentSet"].(bool) {
		o.current = o.generators[o.index]
	}
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

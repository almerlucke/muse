package values

import (
	"math/rand"

	"github.com/almerlucke/muse"
)

type Transformer[T any] interface {
	Transform(T) T
}

type Valuer[T any] interface {
	muse.Stater
	Value() T
	Continuous() bool
	Reset()
	Finished() bool
}

func Generate[T any](value Valuer[T], n int) []T {
	sequence := []T{}

	for i := 0; i < n; i++ {
		sequence = append(sequence, value.Value())
		if value.Finished() {
			value.Reset()
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

func (r *Ramp) Value() float64 {
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

func (r *Repeat[T]) GetState() map[string]any {
	return map[string]any{"value": r.value.GetState(), "min": r.min, "max": r.max, "n": r.n}
}

func (r *Repeat[T]) SetState(state map[string]any) {
	r.value.SetState(state["value"].(map[string]any))
	r.min = state["min"].(int)
	r.max = state["max"].(int)
	r.n = state["n"].(int)
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

func (a *And[T]) GetState() map[string]any {
	states := []map[string]any{}
	for index, value := range a.values {
		states[index] = value.GetState()
	}

	return map[string]any{"values": states, "isCurrentSet": a.current != nil, "index": a.index, "continuous": a.continuous}
}

func (a *And[T]) SetState(state map[string]any) {
	for index, state := range state["values"].([]map[string]any) {
		a.values[index].SetState(state)
	}

	a.continuous = state["continuous"].(bool)
	a.index = state["index"].(int)
	a.current = nil

	if state["isCurrentSet"].(bool) {
		a.current = a.values[a.index]
	}
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

func (o *Or[T]) GetState() map[string]any {
	states := []map[string]any{}
	for index, value := range o.values {
		states[index] = value.GetState()
	}

	return map[string]any{"values": states, "isCurrentSet": o.current != nil, "index": o.index, "continuous": o.continuous, "finished": o.finished}
}

func (o *Or[T]) SetState(state map[string]any) {
	for index, state := range state["values"].([]map[string]any) {
		o.values[index].SetState(state)
	}

	o.index = state["index"].(int)
	o.continuous = state["continuous"].(bool)
	o.finished = state["finished"].(bool)

	if state["isCurrentSet"].(bool) {
		o.current = o.values[o.index]
	}
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

func (t *Transform[T]) GetState() map[string]any {
	return t.value.GetState()
}

func (t *Transform[T]) SetState(state map[string]any) {
	t.value.SetState(state)
}

type TFunc[T any] func(T) T

func (f TFunc[T]) Transform(v T) T {
	return f(v)
}

// Prototype is a prototype of a map. When Map() is called, a deep copy of the prototype is made with all Valuer values
// in the prototype replaced with the Value() from that Valuer. In the deep copy all placeholder values are replaced with the matching replacements
type Prototype map[string]any

type Placeholder struct {
	Name string
}

type Replacement struct {
	Name  string
	Value any
}

func NewPlaceholder(name string) *Placeholder {
	return &Placeholder{Name: name}
}

func NewReplacement(name string, value any) *Replacement {
	return &Replacement{Name: name, Value: value}
}

func (p Prototype) Map(replacements []*Replacement) map[string]any {
	m := map[string]any{}

	for k, v := range p {
		switch vt := v.(type) {
		case Prototype:
			m[k] = vt.Map(replacements)
		case Valuer[any]:
			m[k] = vt.Value()
		case *Placeholder:
			for _, replacement := range replacements {
				if vt.Name == replacement.Name {
					m[k] = replacement.Value
					break
				}
			}
		case Placeholder:
			for _, replacement := range replacements {
				if vt.Name == replacement.Name {
					m[k] = replacement.Value
					break
				}
			}
		default:
			m[k] = v
		}
	}

	return m
}

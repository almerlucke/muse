package markov

import (
	"math/rand"

	"github.com/almerlucke/muse/utils/queue"
)

type State[T any] struct {
	Value        T
	Transitioner Transitioner[T]
}

func NewState[T any](value T, transitioner Transitioner[T]) *State[T] {
	return &State[T]{
		Value:        value,
		Transitioner: transitioner,
	}
}

type Starter[T any] interface {
	Start(*Markov[T]) *State[T]
}

type StateStarter[T any] struct {
	State *State[T]
}

func NewStateStarter[T any](state *State[T]) *StateStarter[T] {
	return &StateStarter[T]{State: state}
}

func (s *StateStarter[T]) Start(m *Markov[T]) *State[T] {
	return s.State
}

type Transitioner[T any] interface {
	Transition(*Markov[T]) *State[T]
}

type ProbabilityTransition[T any] struct {
	State       *State[T]
	Probability float64
}

func NewProbabilityTransition[T any](state *State[T], probability float64) *ProbabilityTransition[T] {
	return &ProbabilityTransition[T]{State: state, Probability: probability}
}

type ProbabilityTransitioner[T any] struct {
	Transitions []*ProbabilityTransition[T]
	Total       float64
}

func NewProbabilityTransitioner[T any](transitions []*ProbabilityTransition[T]) *ProbabilityTransitioner[T] {
	pt := &ProbabilityTransitioner[T]{Transitions: transitions}

	total := 0.0
	for _, transition := range transitions {
		total += transition.Probability
	}

	pt.Total = total

	return pt
}

func NewProbabilityTransitionerVariadic[T any](states ...any) *ProbabilityTransitioner[T] {
	trans := []*ProbabilityTransition[T]{}

	for i := 0; i < len(states); i += 2 {
		trans = append(trans, NewProbabilityTransition(states[i].(*State[T]), states[i+1].(float64)))
	}

	return NewProbabilityTransitioner(trans)
}

func (pt *ProbabilityTransitioner[T]) Transition(m *Markov[T]) *State[T] {
	r := rand.Float64() * pt.Total
	acc := 0.0

	for _, transition := range pt.Transitions {
		acc += transition.Probability
		if r < acc {
			return transition.State
		}
	}

	return nil
}

type Markov[T any] struct {
	History      *queue.Lifo[*State[T]]
	Starter      Starter[T]
	CurrentState *State[T]
}

func NewMarkov[T any](starter Starter[T], historySize int) *Markov[T] {
	beginState := starter.Start(nil)

	return &Markov[T]{
		History:      queue.NewLifo[*State[T]](historySize),
		Starter:      starter,
		CurrentState: beginState,
	}
}

func (m *Markov[T]) Value() T {
	var v T

	if m.CurrentState == nil {
		return v
	}

	v = m.CurrentState.Value

	m.History.Push(m.CurrentState)

	m.CurrentState = m.CurrentState.Transitioner.Transition(m)

	return v
}

func (m *Markov[T]) Continuous() bool {
	return false
}

func (m *Markov[T]) Reset() {
	m.CurrentState = m.Starter.Start(m)
	m.History = queue.NewLifo[*State[T]](m.History.Size())
}

func (m *Markov[T]) Finished() bool {
	return m.CurrentState == nil
}

func (m *Markov[T]) GetState() map[string]any {
	return map[string]any{}
}

func (m *Markov[T]) SetState(map[string]any) {

}

package arpeggio

import "math/rand"

type Mode int

type MirrorMode int

const (
	Up Mode = iota
	Converge
	Alternate
	Random
)

const (
	None MirrorMode = iota
	Exclusive
	Inclusive
)

type Arpeggiator[T any] struct {
	Sequence   []T
	Mode       Mode
	MirrorMode MirrorMode
	Reverse    bool
	pattern    []int
	index      int
	continuous bool
}

func NewArpeggiator[T any](sequence []T, mode Mode, mirrorMode MirrorMode, reverse bool) *Arpeggiator[T] {
	a := &Arpeggiator[T]{
		Sequence:   sequence,
		Mode:       mode,
		MirrorMode: mirrorMode,
		Reverse:    reverse,
		continuous: true,
	}

	a.generatePattern()

	return a
}

func NewArpeggiatorNC[T any](sequence []T, mode Mode, mirrorMode MirrorMode, reverse bool) *Arpeggiator[T] {
	a := &Arpeggiator[T]{
		Sequence:   sequence,
		Mode:       mode,
		MirrorMode: mirrorMode,
		Reverse:    reverse,
		continuous: false,
	}

	a.generatePattern()

	return a
}

func (a *Arpeggiator[T]) Value() T {
	var v T

	if a.index < len(a.pattern) {
		v = a.Sequence[a.pattern[a.index]]
		a.index++
		if a.index == len(a.pattern) && a.continuous {
			a.index = 0
		}
	} else if a.continuous {
		a.index = 0
		return a.Value()
	} else {
		v = a.Sequence[a.pattern[a.index-1]]
	}

	return v
}

func (a *Arpeggiator[T]) Continuous() bool {
	return a.continuous
}

func (a *Arpeggiator[T]) Reset() {
	a.index = 0
	a.generatePattern()
}

func (a *Arpeggiator[T]) Finished() bool {
	return !a.continuous && a.index == len(a.pattern)
}

func (a *Arpeggiator[T]) GetState() map[string]any {
	return map[string]any{
		"pattern":    a.pattern,
		"sequence":   a.Sequence,
		"index":      a.index,
		"continuous": a.continuous,
		"mode":       int(a.Mode),
		"reverse":    a.Reverse,
		"mirrorMode": int(a.MirrorMode),
	}
}

func (a *Arpeggiator[T]) SetState(state map[string]any) {
	a.pattern = state["pattern"].([]int)
	a.Sequence = state["sequence"].([]T)
	a.index = state["index"].(int)
	a.continuous = state["continuous"].(bool)
	a.Mode = Mode(state["mode"].(int))
	a.MirrorMode = MirrorMode(state["mirrorMode"].(int))
}

func (a *Arpeggiator[T]) generatePattern() {
	n := len(a.Sequence)
	pattern := []int{}

	switch a.Mode {
	case Up:
		for i := 0; i < n; i++ {
			pattern = append(pattern, i)
		}
	case Converge:
		for i := 0; i < n/2; i++ {
			pattern = append(pattern, i, n-1-i)
		}
		if n%2 == 1 {
			pattern = append(pattern, n/2)
		}
	case Alternate:
		for i := 1; i < n; i++ {
			pattern = append(pattern, 0, i)
		}
	case Random:
		for i := 0; i < n; i++ {
			pattern = append(pattern, i)
		}
		rand.Shuffle(len(pattern), func(i, j int) { pattern[i], pattern[j] = pattern[j], pattern[i] })
	}

	n = len(pattern)

	if a.MirrorMode == Exclusive {
		for i := n - 2; i >= 1; i-- {
			pattern = append(pattern, pattern[i])
		}
	} else if a.MirrorMode == Inclusive {
		for i := n - 1; i >= 0; i-- {
			pattern = append(pattern, pattern[i])
		}
	}

	n = len(pattern)

	if a.Reverse {
		for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
			pattern[i], pattern[j] = pattern[j], pattern[i]
		}
	}

	a.pattern = pattern
}

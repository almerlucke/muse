package queue

type Lifo[T any] struct {
	q    []T
	size int
}

func NewLifo[T any](size int) *Lifo[T] {
	return &Lifo[T]{
		q:    []T{},
		size: size,
	}
}

func (lifo *Lifo[T]) Size() int {
	return lifo.size
}

func (lifo *Lifo[T]) Queue() []T {
	return lifo.q
}

func (lifo *Lifo[T]) Pop() T {
	var elem T

	if len(lifo.q) == 0 {
		return elem
	}

	elem = lifo.q[0]

	lifo.q = lifo.q[1:]

	return elem
}

func (lifo *Lifo[T]) Push(v T) {
	if len(lifo.q) < lifo.size {
		lifo.q = append([]T{v}, lifo.q...)
		return
	}

	copy(lifo.q[1:], lifo.q)

	lifo.q[0] = v
}

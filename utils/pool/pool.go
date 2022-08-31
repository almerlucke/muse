package pool

type Element[T any] struct {
	Value T
	Prev  *Element[T]
	Next  *Element[T]
}

func (e *Element[T]) Unlink() {
	e.Prev.Next = e.Next
	e.Next.Prev = e.Prev
}

type Pool[T any] struct {
	sentinel *Element[T]
}

func NewPool[T any]() *Pool[T] {
	p := &Pool[T]{}
	sentinel := &Element[T]{}
	sentinel.Next = sentinel
	sentinel.Prev = sentinel
	p.sentinel = sentinel
	return p
}

func (p *Pool[T]) End() *Element[T] {
	return p.sentinel
}

func (p *Pool[T]) First() *Element[T] {
	return p.sentinel.Next
}

func (p *Pool[T]) Pop() *Element[T] {
	first := p.sentinel.Next

	if first == p.sentinel {
		return nil
	}

	first.Unlink()

	return first
}

func (p *Pool[T]) Push(e *Element[T]) {
	e.Next = p.sentinel.Next
	e.Prev = p.sentinel
	p.sentinel.Next.Prev = e
	p.sentinel.Next = e
}

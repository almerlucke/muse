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

type Iterator[T any] struct {
	cur      *Element[T]
	sentinel *Element[T]
}

func (it *Iterator[T]) InitWithPool(pool *Pool[T]) {
	it.cur = pool.sentinel.Next
	it.sentinel = pool.sentinel
}

func (it *Iterator[T]) Finished() bool {
	return it.cur != it.sentinel
}

func (it *Iterator[T]) Reset() {
	it.cur = it.sentinel.Next
}

func (it *Iterator[T]) Value() (v T, ok bool) {
	e := it.cur

	if e != it.sentinel {
		v = it.cur.Value
		ok = true
	}

	return
}

func (it *Iterator[T]) Remove() {
	var e = it.cur

	if e == it.sentinel {
		return
	}

	it.cur = it.cur.Next

	e.Unlink()
}

func (it *Iterator[T]) Next() (v T, ok bool) {
	var e = it.cur

	if e != it.sentinel {
		v = e.Value
		ok = true
		it.cur = it.cur.Next
	}

	return
}

type Pool[T any] struct {
	sentinel *Element[T]
}

func New[T any]() *Pool[T] {
	p := &Pool[T]{}
	sentinel := &Element[T]{}
	sentinel.Next = sentinel
	sentinel.Prev = sentinel
	p.sentinel = sentinel
	return p
}

func (p *Pool[T]) Iterator() *Iterator[T] {
	return &Iterator[T]{
		cur:      p.sentinel.Next,
		sentinel: p.sentinel,
	}
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

func (p *Pool[T]) PushValue(v T) {
	p.Push(&Element[T]{Value: v})
}

func (p *Pool[T]) PopValue() T {
	var v T

	e := p.Pop()
	if e != nil {
		v = e.Value
	}

	return v
}

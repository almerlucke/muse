package list

type Element[T any] struct {
	Value T
	Prev  *Element[T]
	Next  *Element[T]
}

func NewElement[T any](t T) *Element[T] {
	return &Element[T]{
		Value: t,
		Prev:  nil,
		Next:  nil,
	}
}

func (e *Element[T]) Unlink() {
	e.Prev.Next = e.Next
	e.Next.Prev = e.Prev
}

type Iterator[T any] struct {
	cur  *Element[T]
	root *Element[T]
}

func (it *Iterator[T]) Finished() bool {
	return it.cur == it.root
}

func (it *Iterator[T]) Reset() {
	it.cur = it.root.Next
}

func (it *Iterator[T]) Value() (v T, ok bool) {
	e := it.cur

	if e != it.root {
		v = it.cur.Value
		ok = true
	}

	return
}

func (it *Iterator[T]) Element() *Element[T] {
	return it.cur
}

func (it *Iterator[T]) Remove() *Element[T] {
	var e = it.cur

	if e == it.root {
		return nil
	}

	it.cur = it.cur.Next

	e.Unlink()

	return e
}

func (it *Iterator[T]) Next() (v T, ok bool) {
	var e = it.cur

	if e != it.root {
		v = e.Value
		ok = true
		it.cur = it.cur.Next
	}

	return
}

func (it *Iterator[T]) Prev() (v T, ok bool) {
	var e = it.cur

	if e != it.root {
		v = e.Value
		ok = true
		it.cur = it.cur.Prev
	}

	return
}

func (it *Iterator[T]) Forward() {
	if it.cur != it.root {
		it.cur = it.cur.Next
	}
}

func (it *Iterator[T]) Backward() {
	if it.cur != it.root {
		it.cur = it.cur.Prev
	}
}

type List[T any] struct {
	root *Element[T]
}

func New[T any]() *List[T] {
	l := &List[T]{}
	root := &Element[T]{}
	root.Next = root
	root.Prev = root
	l.root = root
	return l
}

func (l *List[T]) Iterator(fromStart bool) *Iterator[T] {
	if fromStart {
		return &Iterator[T]{
			cur:  l.root.Next,
			root: l.root,
		}
	}

	// Return iterator from last
	return &Iterator[T]{
		cur:  l.root.Prev,
		root: l.root,
	}
}

func (l *List[T]) End() *Element[T] {
	return l.root
}

func (l *List[T]) Last() *Element[T] {
	return l.root.Prev
}

func (l *List[T]) First() *Element[T] {
	return l.root.Next
}

func (l *List[T]) PushElement(e *Element[T]) {
	e.Next = l.root.Next
	e.Prev = l.root
	l.root.Next.Prev = e
	l.root.Next = e
}

func (l *List[T]) PushBackElement(e *Element[T]) {
	e.Next = l.root
	e.Prev = l.root.Prev
	l.root.Prev.Next = e
	l.root.Prev = e
}

func (l *List[T]) PopElement() *Element[T] {
	first := l.root.Next

	if first == l.root {
		return nil
	}

	first.Unlink()

	return first
}

func (l *List[T]) PopBackElement() *Element[T] {
	last := l.root.Prev

	if last == l.root {
		return nil
	}

	last.Unlink()

	return last
}

func (l *List[T]) Push(v T) {
	l.PushElement(&Element[T]{Value: v})
}

func (l *List[T]) PushBack(v T) {
	l.PushBackElement(&Element[T]{Value: v})
}

func (l *List[T]) Pop() T {
	var v T

	e := l.PopElement()
	if e != nil {
		v = e.Value
	}

	return v
}

func (l *List[T]) PopBack() T {
	var v T

	e := l.PopBackElement()
	if e != nil {
		v = e.Value
	}

	return v
}

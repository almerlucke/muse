package list

type Element[T any] struct {
	Value T
	Prev  *Element[T]
	Next  *Element[T]
}

func (e *Element[T]) Unlink() *Element[T] {
	e.Prev.Next = e.Next
	e.Next.Prev = e.Prev
	return e
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

func (l *List[T]) Length() int {
	var (
		n  int
		it = l.Iterator(true)
	)

	for _, ok := it.Next(); ok; _, ok = it.Next() {
		n++
	}

	return n
}

func (l *List[T]) Clear() {
	l.root.Next = l.root
	l.root.Prev = l.root
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

func (l *List[T]) Until(f func(T, int) bool) {
	for e, i := l.root.Next, 0; e != l.root; e = e.Next {
		if !f(e.Value, i) {
			return
		}
		i++
	}
}

func (l *List[T]) ForEach(f func(T, int)) {
	l.Until(func(t T, index int) bool {
		f(t, index)
		return true
	})
}

func (l *List[T]) ForEachElement(f func(*Element[T], int)) {
	for e, i := l.root.Next, 0; e != l.root; {
		prev := e
		e = e.Next
		f(prev, i) // call f on prev so f can unlink element safely if needed
		i++
	}
}

func (l *List[T]) Map(f func(T) T) *List[T] {
	cpy := New[T]()

	l.ForEach(func(t T, _ int) {
		cpy.PushBack(f(t))
	})

	return cpy
}

func (l *List[T]) Reduce(initValue T, f func(T, T) T) T {
	var accum = initValue

	l.ForEach(func(t T, _ int) {
		accum = f(accum, t)
	})

	return accum
}

func (l *List[T]) EndElement() *Element[T] {
	return l.root
}

func (l *List[T]) FirstElement() *Element[T] {
	return l.root.Next
}

func (l *List[T]) LastElement() *Element[T] {
	return l.root.Prev
}

func (l *List[T]) First() T {
	return l.FirstElement().Value
}

func (l *List[T]) Last() T {
	return l.LastElement().Value
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

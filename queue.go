package nbchanqueue

// NewQueue returns a List using a queue implementation.
func NewQueue[E any]() *List[E, *ListQueue[E]] {
	return New(NewListQueue[E]())
}

// ListQueue is a queue implementation for ListType.
type ListQueue[T any] struct {
	q []T
}

func NewListQueue[T any]() *ListQueue[T] {
	return &ListQueue[T]{}
}

var _ ListType[int] = (*ListQueue[int])(nil)

func (l *ListQueue[T]) Empty() bool {
	return len(l.q) == 0
}

func (l *ListQueue[T]) Put(t T) {
	l.q = append(l.q, t)
}

func (l *ListQueue[T]) Peek() (T, bool) {
	if len(l.q) > 0 {
		return l.q[0], true
	}
	var t T
	return t, false
}

func (l *ListQueue[T]) Pop() {
	if len(l.q) > 0 {
		l.q = l.q[1:]
	}
}

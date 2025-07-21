package nbchanqueue

type ListType[T any] interface {
	Put(T)
	Empty() bool
	Peek() (T, bool)
	Pop()
}

package nbchanqueue

// ListType is an unbounded list implementation. It should never fail if not for memory starvation.
// It will never be accessed concurrently, so it doesn't need to be goroutine-safe.
type ListType[T any] interface {
	Put(T)           // add item to list
	Peek() (T, bool) // peek the next item that will be returned by the next Pop call if there is one.
	Pop()            // remove the item that was/would be returned by the previous Peek call.
}

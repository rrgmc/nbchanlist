package nbchanqueue

import "sync/atomic"

// Queue is a non-blocking unbounded lock-free channel-based queue for Golang.
type Queue[E any] struct {
	out    <-chan E
	in     chan<- E
	info   <-chan queueChanInfo
	closed atomic.Bool
}

func New[E any]() *Queue[E] {
	out, in, info := newQueueChan[E]()
	return &Queue[E]{
		out:  out,
		in:   in,
		info: info,
	}
}

// Get returns a channel to get items. The caller must check for it to be closed.
func (q *Queue[E]) Get() <-chan E {
	return q.out
}

// Put puts an element in the queue. It never fails or blocks.
func (q *Queue[E]) Put(e E) {
	_ = q.PutCheck(e)
}

// PutCheck puts an element in the queue. It never fails or blocks.
func (q *Queue[E]) PutCheck(e E) bool {
	if q.closed.Load() {
		return false
	}
	q.in <- e
	return true
}

func (q *Queue[E]) Size() int {
	info := <-q.info
	return info.itemAmount
}

func (q *Queue[E]) Closed() bool {
	return q.closed.Load()
}

func (q *Queue[E]) Close() {
	if q.closed.CompareAndSwap(false, true) {
		close(q.in)
	}
}

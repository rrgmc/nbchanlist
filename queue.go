package nbchanqueue

import "sync/atomic"

// Queue is a non-blocking unbounded lock-free channel-based queue for Golang.
type Queue[E any] struct {
	out  <-chan E
	in   atomic.Pointer[chan<- E]
	info <-chan queueChanInfo
}

func New[E any]() *Queue[E] {
	out, in, info := newQueueChan[E]()
	ret := &Queue[E]{
		out:  out,
		info: info,
	}
	ret.in.Store(&in)
	return ret
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
// Returns false if the queue is closed.
func (q *Queue[E]) PutCheck(e E) bool {
	if in := q.in.Load(); in != nil {
		*in <- e
		return true
	}
	return false
}

func (q *Queue[E]) Size() int {
	select {
	case info, ok := <-q.info:
		if !ok {
			return -1
		}
		return info.itemAmount
	}
}

func (q *Queue[E]) Closed() bool {
	return q.in.Load() == nil
}

func (q *Queue[E]) Close() {
	if old := q.in.Swap(nil); old != nil {
		close(*old)
	}
}

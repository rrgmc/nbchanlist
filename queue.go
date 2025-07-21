package nbchanqueue

import (
	"context"
	"errors"
	"sync/atomic"
)

var (
	ErrClosed = errors.New("queue is closed")
)

// Queue is a non-blocking unbounded lock-free channel-based queue for Golang.
type Queue[E any, Q ListType[E]] struct {
	out  <-chan E
	in   atomic.Pointer[chan<- E]
	info <-chan queueChanInfo
}

func New[E any]() *Queue[E, *ListQueue[E]] {
	return New2(&ListQueue[E]{})
}

func New2[E any, Q ListType[E]](list Q) *Queue[E, Q] {
	out, in, info := newQueueChan[E](list)
	ret := &Queue[E, Q]{
		out:  out,
		info: info,
	}
	ret.in.Store(&in)
	return ret
}

// Get returns a channel to get items. The caller must check for it to be closed.
func (q *Queue[E, Q]) Get() <-chan E {
	return q.out
}

// GetCtx returns one item, or an error if the context is done or the queue is closed.
func (q *Queue[E, Q]) GetCtx(ctx context.Context) (E, error) {
	select {
	case <-ctx.Done():
		var ret E
		return ret, context.Cause(ctx)
	case v, ok := <-q.Get():
		if !ok {
			var ret E
			return ret, ErrClosed
		}
		return v, nil
	}
}

// Put puts an element in the queue. It never fails or blocks.
func (q *Queue[E, Q]) Put(e E) {
	_ = q.PutCheck(e)
}

// PutCheck puts an element in the queue. It never fails or blocks.
// Returns ErrClosed if the queue is closed.
func (q *Queue[E, Q]) PutCheck(e E) error {
	if in := q.in.Load(); in != nil {
		*in <- e
		return nil
	}
	return ErrClosed
}

func (q *Queue[E, Q]) Size() int {
	select {
	case info, ok := <-q.info:
		if !ok {
			return -1
		}
		return info.itemAmount
	}
}

func (q *Queue[E, Q]) Closed() bool {
	return q.in.Load() == nil
}

func (q *Queue[E, Q]) Close() {
	if old := q.in.Swap(nil); old != nil {
		close(*old)
	}
}

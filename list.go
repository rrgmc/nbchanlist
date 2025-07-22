package nbchanlist

import (
	"context"
	"errors"
	"sync/atomic"
)

var (
	ErrClosed = errors.New("closed")
)

// List is a non-blocking unbounded lock-free channel-based list.
type List[E any, Q ListType[E]] struct {
	out  <-chan E
	in   atomic.Pointer[chan<- E]
	info <-chan listChanInfo
}

func New[E any, Q ListType[E]](list Q) *List[E, Q] {
	out, in, info := newListChan[E](list)
	ret := &List[E, Q]{
		out:  out,
		info: info,
	}
	ret.in.Store(&in)
	return ret
}

// Get returns a channel to get items. The caller must check for it to be closed.
// It can still be called even after Close.
func (q *List[E, Q]) Get() <-chan E {
	return q.out
}

// GetCtx returns one item, or an error if the context is done or the list is closed.
// It can still be called even after Close.
func (q *List[E, Q]) GetCtx(ctx context.Context) (E, error) {
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

// Put puts an element in the list. It never fails or blocks.
func (q *List[E, Q]) Put(e E) {
	_ = q.PutCheck(e)
}

// PutCheck puts an element in the list. It never fails or blocks.
// Returns ErrClosed if the list stopped accepting new items.
func (q *List[E, Q]) PutCheck(e E) error {
	if in := q.in.Load(); in != nil {
		*in <- e
		return nil
	}
	return ErrClosed
}

// Size returns the number of items in the list.
// If the list is shutdown, returns -1.
func (q *List[E, Q]) Size() int {
	select {
	case info, ok := <-q.info:
		if !ok {
			return -1
		}
		return info.itemAmount
	}
}

// Closed returns whether Close was called.
func (q *List[E, Q]) Closed() bool {
	return q.in.Load() == nil
}

// Close stops accepting new items. Get still works until the list is drained.
func (q *List[E, Q]) Close() {
	if in := q.in.Swap(nil); in != nil {
		close(*in)
	}
}

// Shutdown stops accepting new items and drains any existing ones, freeing all used resources.
func (q *List[E, Q]) Shutdown() {
	q.Close()
	// drain items to stop goroutine.
	for range q.out {
	}
}

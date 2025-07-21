package nbchanlist

import (
	"context"
	"errors"
	"sync/atomic"
)

var (
	ErrStopped = errors.New("stopped")
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
// It can still be called even after Stop.
func (q *List[E, Q]) Get() <-chan E {
	return q.out
}

// GetCtx returns one item, or an error if the context is done or the list is closed.
// It can still be called even after Stop.
func (q *List[E, Q]) GetCtx(ctx context.Context) (E, error) {
	select {
	case <-ctx.Done():
		var ret E
		return ret, context.Cause(ctx)
	case v, ok := <-q.Get():
		if !ok {
			var ret E
			return ret, ErrStopped
		}
		return v, nil
	}
}

// Put puts an element in the list. It never fails or blocks.
func (q *List[E, Q]) Put(e E) {
	_ = q.PutCheck(e)
}

// PutCheck puts an element in the list. It never fails or blocks.
// Returns ErrStopped if the list stopped accepting new items.
func (q *List[E, Q]) PutCheck(e E) error {
	if in := q.in.Load(); in != nil {
		*in <- e
		return nil
	}
	return ErrStopped
}

// Size returns the number of items in the list.
// If the list is closed, returns -1.
func (q *List[E, Q]) Size() int {
	select {
	case info, ok := <-q.info:
		if !ok {
			return -1
		}
		return info.itemAmount
	}
}

// Stopped returns whether Stop was called.
func (q *List[E, Q]) Stopped() bool {
	return q.in.Load() == nil
}

// Stop stops accepting new items. Get still works until the list is drained.
func (q *List[E, Q]) Stop() {
	if in := q.in.Swap(nil); in != nil {
		close(*in)
	}
}

// Close stops accepting new items and drains any existing ones, freeing all used resources.
func (q *List[E, Q]) Close() {
	q.Stop()
	// drain items to stop goroutine.
	for range q.out {
	}
}

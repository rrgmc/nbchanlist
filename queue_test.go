package nbchanqueue

import (
	"context"
	"errors"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestQueue(t *testing.T) {
	q := New[int]()
	q.Put(12)
	q.Put(13)
	assert.Equal(t, q.Size(), 2)
	assertItems(t, q, []int{12, 13})
	assert.Equal(t, q.Size(), 0)
}

func TestQueueCtx(t *testing.T) {
	ctx := context.Background()
	q := New[int]()
	q.Put(12)
	q.Put(13)
	assert.Equal(t, q.Size(), 2)
	assertItemsCtx(t, ctx, q, []int{12, 13})
	assert.Equal(t, q.Size(), 0)
}

func TestQueueClose(t *testing.T) {
	q := New[int]()
	q.Put(12)
	q.Put(13)
	q.Close()
	q.Put(16)
	assert.Assert(t, q.Closed())
	expected := []int{12, 13}
	items, err := readNWithTimeout(q, 3)
	assert.ErrorIs(t, err, ErrClosed)
	assert.DeepEqual(t, expected, items)
}

func TestQueueTimeout(t *testing.T) {
	q := New[int]()
	q.Put(12)
	q.Put(13)
	expected := []int{12, 13}
	items, err := readNWithTimeout(q, 3)
	assert.ErrorIs(t, err, errTestTimeout)
	assert.DeepEqual(t, expected, items)
}

func TestQueueCtxTimeout(t *testing.T) {
	ctx := context.Background()
	q := New[int]()
	q.Put(12)
	q.Put(13)
	expected := []int{12, 13}
	items, err := readNWithContext(ctx, q, 3)
	assert.ErrorIs(t, err, errTestTimeout)
	assert.DeepEqual(t, expected, items)
}

func assertItems[E any](t *testing.T, q *Queue[E], expected []E) {
	items, err := readNWithTimeout(q, len(expected))
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, items)
}

func assertItemsCtx[E any](t *testing.T, ctx context.Context, q *Queue[E], expected []E) {
	items, err := readNWithContext(ctx, q, len(expected))
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, items)
}

func readNWithContext[E any](ctx context.Context, q *Queue[E], n int) ([]E, error) {
	var ret []E
	for i := 0; i < n; i++ {
		xerr := func() error {
			subCtx, subCancel := context.WithTimeout(ctx, 300*time.Millisecond)
			defer subCancel()
			item, err := q.GetCtx(subCtx)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					return errTestTimeout
				}
				return err
			}
			ret = append(ret, item)
			return nil
		}()
		if xerr != nil {
			return ret, xerr
		}
	}
	return ret, nil
}

func readNWithTimeout[E any](q *Queue[E], n int) ([]E, error) {
	var ret []E
	for i := 0; i < n; i++ {
		select {
		case item, ok := <-q.Get():
			if !ok {
				return ret, ErrClosed
			}
			ret = append(ret, item)
		case <-time.After(300 * time.Millisecond):
			return ret, errTestTimeout
		}
	}
	return ret, nil
}

var (
	errTestTimeout = errors.New("timeout")
)

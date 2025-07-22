package nbchanlist

import (
	"cmp"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/v3/assert"
	cmp2 "gotest.tools/v3/assert/cmp"
)

func TestQueue(t *testing.T) {
	q := NewQueue[int]()
	q.Put(12)
	q.Put(13)
	assert.Equal(t, q.Size(), 2)
	assertItems(t, q, []int{12, 13})
	assert.Equal(t, q.Size(), 0)
}

func TestQueueCtx(t *testing.T) {
	ctx := context.Background()
	q := NewQueue[int]()
	q.Put(12)
	q.Put(13)
	assert.Equal(t, q.Size(), 2)
	assertItemsCtx(t, ctx, q, []int{12, 13})
	assert.Equal(t, q.Size(), 0)

	q.Shutdown()
	_, err := q.GetCtx(ctx)
	assert.ErrorIs(t, err, ErrClosed)
}

func TestQueueClose(t *testing.T) {
	q := NewQueue[int]()
	defer q.Shutdown()
	q.Put(12)
	q.Put(13)
	q.Close()
	q.Put(16)
	assert.Assert(t, q.Closed())
	expected := []int{12, 13}
	items, err := readNWithTimeout(q, 3) // can still Get after Close.
	assert.ErrorIs(t, err, ErrClosed)
	assert.DeepEqual(t, expected, items) // still gets the expected items.
}

func TestQueueTimeout(t *testing.T) {
	q := NewQueue[int]()
	defer q.Shutdown()
	q.Put(12)
	q.Put(13)
	expected := []int{12, 13}
	items, err := readNWithTimeout(q, 3)
	assert.ErrorIs(t, err, errTestTimeout)
	assert.DeepEqual(t, expected, items)
}

func TestQueueShutdown(t *testing.T) {
	q := NewQueue[int]()
	q.Put(12)
	q.Put(13)
	q.Shutdown()
	q.Put(16)
	assert.Assert(t, q.Closed())
	assert.Equal(t, -1, q.Size())
	var expected []int                   // Shutdown drains all pending items.
	items, err := readNWithTimeout(q, 2) // can still get after close
	assert.ErrorIs(t, err, ErrClosed)
	assert.DeepEqual(t, expected, items)
}

func TestQueueCtxTimeout(t *testing.T) {
	ctx := context.Background()
	q := NewQueue[int]()
	defer q.Shutdown()
	q.Put(12)
	q.Put(13)
	expected := []int{12, 13}
	items, err := readNWithContext(ctx, q, 3)
	assert.ErrorIs(t, err, errTestTimeout)
	assert.DeepEqual(t, expected, items)
}

func TestQueueConcurrency(t *testing.T) {
	var wgGet sync.WaitGroup
	var wgPut sync.WaitGroup

	q := NewQueue[int]()

	var getItems []int
	var putItems []int
	var putLock sync.Mutex

	wgGet.Add(1)
	go func() {
		defer wgGet.Done()
		for v := range q.Get() {
			getItems = append(getItems, v)
		}
	}()

	for i := 0; i < 10; i++ {
		wgPut.Add(1)
		go func() {
			defer wgPut.Done()
			for j := 0; j < 10; j++ {
				v := i*10 + j
				q.Put(v)
				putLock.Lock()
				putItems = append(putItems, v)
				putLock.Unlock()
			}
		}()
	}
	wgPut.Wait()
	q.Close()
	wgGet.Wait()
	q.Shutdown()

	assert.Assert(t, cmp2.Len(putItems, 10*10))
	assert.Assert(t, cmp2.Len(getItems, 10*10))
	assert.DeepEqual(t, putItems, getItems, cmpopts.SortSlices(cmp.Less[int]))
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

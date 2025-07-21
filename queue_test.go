package nbchanqueue

import (
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
}

func TestQueueClose(t *testing.T) {
	q := New[int]()
	q.Put(12)
	q.Put(13)
	q.Close()
	q.Put(16)
	expected := []int{12, 13}
	items, err := readNWithTimeout(q, 3)
	assert.ErrorIs(t, err, errTestClosed)
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

func assertItems[E any](t *testing.T, q *Queue[E], expected []E) {
	items, err := readNWithTimeout(q, len(expected))
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, items)
}

func readNWithTimeout[E any](q *Queue[E], n int) ([]E, error) {
	var ret []E
	for i := 0; i < n; i++ {
		select {
		case item, ok := <-q.Get():
			if !ok {
				return ret, errTestClosed
			}
			ret = append(ret, item)
		case <-time.After(1 * time.Second):
			return ret, errTestTimeout
		}
	}
	return ret, nil
}

var (
	errTestClosed  = errors.New("channel closed")
	errTestTimeout = errors.New("timeout")
)

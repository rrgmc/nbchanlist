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
	items, err := readNWithTimeout(q, 2)
	assert.NilError(t, err)
	assert.DeepEqual(t, []int{12, 13}, items)
}

func readNWithTimeout[E any](q *Queue[E], n int) ([]E, error) {
	var ret []E
	for i := 0; i < n; i++ {
		select {
		case item, ok := <-q.Get():
			if !ok {
				return nil, errors.New("channel closed")
			}
			ret = append(ret, item)
		case <-time.After(1 * time.Second):
			return nil, errors.New("timeout")
		}
	}
	return ret, nil
}

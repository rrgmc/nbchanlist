package nbchanlist_test

import (
	"fmt"
	"time"

	"github.com/rrgmc/nbchanlist"
)

func ExampleNewQueue() {
	q := nbchanlist.NewQueue[int]()
	q.Put(12) // never blocks
	q.Put(13)
	select {
	case v := <-q.Get():
		fmt.Println(v)
	case <-time.After(time.Second):
		fmt.Println("timeout")
	}
	q.Close() // stops goroutine and close channels
	select {
	case _, ok := <-q.Get():
		if !ok {
			fmt.Println("queue is closed")
		} else {
			fmt.Println("should never happen")
		}
	}

	// Output:
	// 12
	// queue is closed
}

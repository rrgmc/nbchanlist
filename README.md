# nbchanlist - Non-blocking unbounded lock-free channel-based list for Golang
[![GoDoc](https://godoc.org/github.com/rrgmc/nbchanlist?status.png)](https://godoc.org/github.com/rrgmc/nbchanlist)

nbchanlist is a non-blocking unbounded lock-free channel-based list for Golang.

For getting items, a channel is provided, so it can be used with `select`, and allows for context cancellation,
timeouts, and plays nice with other code using channels.

As it is unbounded, care must be taken to avoid memory exhaustion if adding faster than reading. 

It comes with a "queue" implementation, but other types of lists can be used by implementing the `ListType` interface.

```go
import (
    "fmt"
    "time"

    "github.com/rrgmc/nbchanlist"
)

func ExampleQueue() {
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
}
```

## Install

```shell
go get github.com/rrgmc/nbchanlist
```

# License

MIT

### Author

Rangel Reale (rangelreale@gmail.com)

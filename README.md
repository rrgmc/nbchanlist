# nbchanqueue - Non-blocking unbounded lock-free channel-based queue for Golang
[![GoDoc](https://godoc.org/github.com/rrgmc/nbchanqueue?status.png)](https://godoc.org/github.com/rrgmc/nbchanqueue)

nbchanqueue is a non-blocking unbounded lock-free channel-based queue for Golang.

For getting items, a channel is provided, so it can be used with `select`, and allows for context cancellation,
timeouts, and plays nice with other code using channels.

## Install

```shell
go get github.com/rrgmc/nbchanqueue
```

# License

MIT

### Author

Rangel Reale (rangelreale@gmail.com)

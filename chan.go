package nbchanqueue

// newQueueChan returns a non-blocking queue with element type E. The sending end must be closed by the caller to clean up resources.
func newQueueChan[E any]() (<-chan E, chan<- E, <-chan queueChanInfo) {
	out := make(chan E)
	in := make(chan E)
	info := make(chan queueChanInfo)
	qinfo := queueChanInfo{}
	go func() {
		defer close(out)
		defer close(info)
		var q []E
		for {
			var (
				ch chan E
				v  E
			)
			if len(q) > 0 {
				ch, v = out, q[0]
			} else if in == nil {
				return
			}
			select {
			case ch <- v:
				q = q[1:]
				qinfo.itemAmount--
			case info <- qinfo:
			case v, ok := <-in:
				if ok {
					q = append(q, v)
					qinfo.itemAmount++
				} else {
					in = nil
				}
			}
		}
	}()
	return out, in, info
}

type queueChanInfo struct {
	itemAmount int
}

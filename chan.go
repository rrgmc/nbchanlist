package nbchanlist

// newListChan returns a non-blocking list with element type E. The sending end must be closed by the
// caller to clean up resources.
func newListChan[E any, Q ListType[E]](list Q) (<-chan E, chan<- E, <-chan listChanInfo) {
	out := make(chan E)
	in := make(chan E)
	info := make(chan listChanInfo)
	qinfo := listChanInfo{}
	go func() {
		defer close(out)
		defer close(info)
		var ok bool
		for {
			var (
				ch chan E
				v  E
			)

			v, ok = list.Peek()
			if ok {
				ch = out
			} else if in == nil {
				return
			}

			select {
			case ch <- v:
				list.Pop()
				qinfo.itemAmount--
			case info <- qinfo:
			case v, ok := <-in:
				if ok {
					list.Put(v)
					qinfo.itemAmount++
				} else {
					in = nil
				}
			}
		}
	}()
	return out, in, info
}

type listChanInfo struct {
	itemAmount int
}

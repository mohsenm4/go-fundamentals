package once

import (
	"sync"
	"sync/atomic"
)

type BrokenOnce struct {
	state atomic.Uint32
}

func (o *BrokenOnce) Do(f func()) {
	if o.state.CompareAndSwap(0, 1) {
		f()
	}
}

type Once struct {
	done atomic.Uint32
	m    sync.Mutex
}

func (o *Once) Do(f func()) {
	if o.done.Load() == 1 {
		return
	}
	o.doSlow(f)
}

func (o *Once) doSlow(f func()) {
	o.m.Lock()
	defer o.m.Unlock()
	if o.done.Load() == 0 {
		defer o.done.Store(1)
		f()
	}
}

package boundedchannel

import "sync"

type BoundedChannelTest[T any] struct {
	mu sync.Mutex

	notEmpty *sync.Cond
	notFull  *sync.Cond

	buf    []T
	sendx  int
	recvx  int
	count  int
	closed bool
}

func New[T any](capacity int) *BoundedChannelTest[T] {
	if capacity <= 0 {
		panic("capacity must be greater than 0")
	}

	c := &BoundedChannelTest[T]{
		buf: make([]T, capacity),
	}

	c.notEmpty = sync.NewCond(&c.mu)
	c.notFull = sync.NewCond(&c.mu)

	return c
}

func (c *BoundedChannelTest[T]) Send(v T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		panic("send on closed channel")
	}

	for c.count == len(c.buf) {
		c.notFull.Wait()

		if c.closed {
			panic("send on closed channel")
		}
	}

	c.buf[c.sendx] = v

	c.sendx = (c.sendx + 1) % len(c.buf)

	c.count++

	c.notEmpty.Signal()
}

func (c *BoundedChannelTest[T]) Recv() (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var zero T

	for c.count == 0 {
		if c.closed {
			return zero, false
		}

		c.notEmpty.Wait()
	}

	v := c.buf[c.recvx]

	c.buf[c.recvx] = zero

	c.recvx = (c.recvx + 1) % len(c.buf)

	c.count--

	c.notFull.Signal()

	return v, true
}

func (c *BoundedChannelTest[T]) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	c.closed = true
	c.notEmpty.Broadcast()
	c.notFull.Broadcast()
}

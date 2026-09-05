package boundedchannel

import "sync"

type BoundedChannel struct {
	buf      []int
	capacity int
	head     int
	tail     int
	count    int
	mu       sync.Mutex
	cond     *sync.Cond
}

func NewBoundedChannel(capacity int) *BoundedChannel {
	bc := &BoundedChannel{
		buf:      make([]int, capacity),
		capacity: capacity,
	}
	bc.cond = sync.NewCond(&bc.mu)
	return bc
}

func (b *BoundedChannel) Send(value int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for b.count == b.capacity {
		b.cond.Wait()
	}

	b.buf[b.tail] = value
	b.tail = (b.tail + 1) % b.capacity
	b.count++
	b.cond.Signal()
}

func (b *BoundedChannel) Receive() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	for b.count == 0 {
		b.cond.Wait()
	}

	value := b.buf[b.head]
	b.head = (b.head + 1) % b.capacity
	b.count--
	b.cond.Signal()
	return value
}

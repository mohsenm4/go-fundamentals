package main

import "sync"

// recent keeps the last n raw blocks, so peers that are catching up get them
// from memory instead of from disk.
type recent struct {
	mu    sync.Mutex
	n     int
	order []uint64
	raw   map[uint64][]byte
}

func newRecent(n int) *recent {
	return &recent{n: n, raw: make(map[uint64][]byte)}
}

func (r *recent) put(height uint64, raw []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.raw[height] = raw
	r.order = append(r.order, height)
	if len(r.order) > r.n {
		old := r.order[0]
		r.order = r.order[1:]
		delete(r.raw, old)
	}
}

func (r *recent) get(height uint64) ([]byte, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.raw[height]
	return b, ok
}

func (r *recent) len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.raw)
}

package main

import (
	"errors"
	"sync"
)

type ctxt struct {
	done     chan struct{}
	err      error
	parent   *ctxt
	once     sync.Once
	children map[*ctxt]struct{}
	mu       sync.Mutex
	cancel   func()
}

var ErrCanceled = errors.New("context canceled")

func WithCancel(parent *ctxt) (*ctxt, func()) {
	ctx := &ctxt{
		parent:   parent,
		done:     make(chan struct{}),
		children: make(map[*ctxt]struct{}),
	}

	cancel := func() {
		ctx.once.Do(func() {
			ctx.mu.Lock()

			ctx.err = ErrCanceled
			close(ctx.done)

			var children []*ctxt
			for child := range ctx.children {
				children = append(children, child)
			}

			ctx.mu.Unlock()

			for _, child := range children {
				child.cancel()
			}

			if parent != nil {
				parent.mu.Lock()
				delete(parent.children, ctx)
				parent.mu.Unlock()
			}
		})
	}

	ctx.cancel = cancel

	if parent != nil {
		parent.mu.Lock()

		if parent.err != nil {
			parent.mu.Unlock()
			cancel()
		} else {
			parent.children[ctx] = struct{}{}
			parent.mu.Unlock()
		}
	}

	return ctx, cancel
}

func (c *ctxt) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.err
}

func (c *ctxt) Done() <-chan struct{} {
	return c.done
}

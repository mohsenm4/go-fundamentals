package main

import "sync"

type buf struct{ data [64]byte }

var pool sync.Pool

//go:noinline
func localOnly() byte {
	b := &buf{}
	b.data[0] = 1
	return b.data[0]
}

//go:noinline
func withPut() byte {
	b := &buf{}
	b.data[0] = 1
	x := b.data[0]
	pool.Put(b)
	return x
}

func main() {
	localOnly()
	withPut()
}

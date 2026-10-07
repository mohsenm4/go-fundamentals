package main

import (
	"io"
	"strings"
)

//go:noinline
func readIface(r io.Reader) int {
	buf := make([]byte, 64)
	n, _ := r.Read(buf)
	return n
}

//go:noinline
func readConcrete(r *strings.Reader) int {
	buf := make([]byte, 64)
	n, _ := r.Read(buf)
	return n
}

func main() {
	readIface(strings.NewReader("hello"))
	readConcrete(strings.NewReader("hello"))
}

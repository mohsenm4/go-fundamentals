package main

import "testing"

func TestAllocs(t *testing.T) {
	short := []byte("go")
	long := []byte("hello, escape analysis")
	t.Logf("lookup:  %v", testing.AllocsPerRun(1000, func() { lookup(short) }))
	t.Logf("convert: %v", testing.AllocsPerRun(1000, func() { convert(long) }))
}

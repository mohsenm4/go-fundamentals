package main

import (
	"io"
	"strings"
	"testing"
)

func TestAllocs(t *testing.T) {
	r := strings.NewReader("hello")
	var ri io.Reader = r
	t.Logf("readIface:    %v", testing.AllocsPerRun(1000, func() { readIface(ri) }))
	t.Logf("readConcrete: %v", testing.AllocsPerRun(1000, func() { readConcrete(r) }))
}

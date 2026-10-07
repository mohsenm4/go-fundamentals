package main

import "testing"

func TestAllocs(t *testing.T) {
	t.Logf("local:  %v", testing.AllocsPerRun(1000, func() { local(2, 3) }))
	t.Logf("passed: %v", testing.AllocsPerRun(1000, func() { passed(2, 3) }))
}

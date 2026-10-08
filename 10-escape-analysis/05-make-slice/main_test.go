package main

import "testing"

func BenchmarkSmall(b *testing.B) {
	for b.Loop() {
		variable(3)
	}
}

func BenchmarkBig(b *testing.B) {
	for b.Loop() {
		variable(10)
	}
}

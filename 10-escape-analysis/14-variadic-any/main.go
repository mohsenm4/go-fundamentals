package main

var kept []any

//go:noinline
func count(args ...any) int {
	n := 0
	for range args {
		n++
	}
	return n
}

//go:noinline
func keep(args ...any) {
	kept = args
}

//go:noinline
func callCount(n int) int {
	x := n * 1000
	return count(x)
}

//go:noinline
func callKeep(n int) {
	x := n * 1000
	keep(x)
}

func main() {
	callCount(7)
	callKeep(7)
}

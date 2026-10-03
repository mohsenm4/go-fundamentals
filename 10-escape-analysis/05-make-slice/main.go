package main

//go:noinline
func fixed() {
	s := make([]int, 10)
	s[0] = 1
}

//go:noinline
func variable(n int) {
	s := make([]int, n)
	s[0] = 1
}

func main() {
	fixed()
	variable(10)
}

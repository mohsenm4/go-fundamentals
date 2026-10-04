package main

//go:noinline
func grow() int {
	s := make([]int, 0, 4)
	for i := 0; i < 100; i++ {
		s = append(s, i)
	}
	return len(s)
}

func main() {
	grow()
}

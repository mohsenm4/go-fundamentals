package main

//go:noinline
func lookup(b []byte) int {
	m := map[string]int{"go": 1, "rust": 2}
	return m[string(b)]
}

//go:noinline
func convert(b []byte) string {
	s := string(b)
	return s
}

func main() {
	lookup([]byte("go"))
	convert([]byte("hello, escape analysis"))
}

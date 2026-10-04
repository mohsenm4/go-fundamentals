package main

//go:noinline
func fill() int {
	m := make(map[string]*int)
	x := 42
	m["a"] = &x
	return *m["a"]
}

func main() {
	fill()
}

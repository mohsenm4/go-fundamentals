package main

//go:noinline
func once() int {
	n := 0
	defer func() { n++ }()
	return n
}

//go:noinline
func inLoop() int {
	n := 0
	for i := 0; i < 3; i++ {
		defer func() { n++ }()
	}
	return n
}

func main() {
	once()
	inLoop()
}

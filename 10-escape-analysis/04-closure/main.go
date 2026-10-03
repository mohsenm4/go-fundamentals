package main

//go:noinline
func makeCounter() func() {
	count := 0

	return func() {
		count++
	}
}

func main() {
	counter := makeCounter()
	counter()
	counter()
}

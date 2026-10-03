package main

//go:noinline
func addOne(p *int) {
	*p++
}

func main() {
	x := 10
	addOne(&x)
}

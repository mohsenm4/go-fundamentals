package main

//go:noinline
func byValue() {
	ch := make(chan int, 1)
	x := 42
	ch <- x
	<-ch
}

//go:noinline
func byPointer() {
	ch := make(chan *int, 1)
	y := 42
	ch <- &y
	<-ch
}

func main() {
	byValue()
	byPointer()
}

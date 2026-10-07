package main

type shape interface {
	area() int
}

type rect struct{ w, h int }

//go:noinline
func (r rect) area() int {
	return r.w * r.h
}

//go:noinline
func local(w, h int) int {
	var s shape = rect{w: w, h: h}
	return s.area()
}

//go:noinline
func measure(s shape) int {
	return s.area()
}

//go:noinline
func passed(w, h int) int {
	r := rect{w: w, h: h}
	return measure(r)
}

func main() {
	local(2, 3)
	passed(2, 3)
}

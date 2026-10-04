package main

type point struct{ x, y int }

//go:noinline
func withNew() int {
	p := new(point)
	p.x = 1
	p.y = 2
	return p.x + p.y
}

//go:noinline
func withLiteral() int {
	q := &point{x: 1, y: 2}
	return q.x + q.y
}

func main() {
	withNew()
	withLiteral()
}

package main

type box struct{ p *int }

type plain struct{ v int }

//go:noinline
func makeBox() box {
	x := 1
	return box{p: &x}
}

//go:noinline
func makePlain() plain {
	x := 1
	return plain{v: x}
}

func main() {
	makeBox()
	makePlain()
}

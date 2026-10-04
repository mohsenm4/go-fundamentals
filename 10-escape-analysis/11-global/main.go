package main

var global *int

//go:noinline
func save() {
	x := 42
	global = &x
}

//go:noinline
func readOnly() int {
	y := 42
	p := &y
	return *p + *global
}

func main() {
	save()
	readOnly()
}

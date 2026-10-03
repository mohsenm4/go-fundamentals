package main

import "fmt"

//go:noinline
func createUser() *int {
	x := 42
	return &x
}

func main() {
	p := createUser()
	fmt.Println(*p)
}

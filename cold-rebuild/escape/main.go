package main

import "fmt"

// 1. reason: Returning a pointer to a local variable
func f1() *int {
	x := 10
	return &x
}

var global *int

// 2. reason: Storing a pointer in a global variable
func f2() {
	x := 20
	global = &x
}

var sink any

// 3. reason: Storing a pointer in an interface that outlives the function
func f3() {
	x := 30
	sink = &x
}

var saved []*int

// 4. reason: Passing a pointer to a function that retains it
func f4() {
	x := 40
	save(&x)
}

func save(p *int) {
	saved = append(saved, p)
}

// 5. reason: Allocating a large local array that may exceed the stack limit
func f5() {
	x := [1 << 20]int{}
	_ = x
}

func main() {
	fmt.Println(*f1())
	f2()
	f3()
	f4()
	f5()
}

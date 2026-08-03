package main

import "fmt"

func Constants() {
	const PiUntyped = 3.14159
	var a float32 = PiUntyped // کار می‌کنه!
	var b float64 = PiUntyped // کار می‌کنه!
	fmt.Println(a, b)

	const PiTyped float64 = 3.14159
	var c float32 = PiTyped // این خط رو uncomment کن، error چیه؟
	fmt.Println(c)
}
func main() {
	Constants()
}

package main

// i want to use every type int in go in a func
func useIntTypes() {
	var a int = 10
	var b int8 = 20
	var c int16 = 30
	var d int32 = 40
	var e int64 = 50

	println(a, b, c, d, e)
}

// i want to use every type uint in go in a func
func useUintTypes() {
	var a uint = 10
	var b uint8 = 20
	var c uint16 = 30
	var d uint32 = 40
	var e uint64 = 50

	println(a, b, c, d, e)
}

// i want to use every type float in go in a func
func useFloatTypes() {
	var a float32 = 10.5
	var b float64 = 20.5

	println(a, b)
}

// i want to use every type complex in go in a func
func useComplexTypes() {
	var a complex64 = 10 + 5i
	var b complex128 = 20 + 10i

	println(a, b)
}

// i want to use every type byte in go in a func
func useByteTypes() {
	var a byte = 10
	var b byte = 20

	println(a, b)
}

// i want to use every type rune in go in a func
func useRuneTypes() {
	var a rune = 'a'
	var b rune = 'b'

	println(a, b)
}

// i want to use every type uintptr in go in a func
func useUintptrTypes() {
	var a uintptr = 10
	var b uintptr = 20

	println(a, b)
}

// i want to use every type pointer in go in a func
func usePointerTypes() {
	var a *int = new(int)
	*a = 10

	var b *string = new(string)
	*b = "hello"

	println(*a, *b)
}

// i want to use every type interface in go in a func
func useInterfaceTypes() {
	var a interface{} = 10
	var b interface{} = "hello"

	println(a, b)
}

// i want to use every type string in go in a func
func useStringTypes() {
	var a string = "hello"
	var b string = "world"

	println(a, b)
}

// i want to use every type bool in go in a func
func useBoolTypes() {
	var a bool = true
	var b bool = false

	println(a, b)
}

// i want to show Conversion between different numeric types in go in a func
func useNumericConversions() {
	var a int = 10
	var b float64 = float64(a) // convert int to float64

	var c float32 = 20.5
	var d int = int(c) // convert float32 to int

	println(b, d)
}

func main() {
	useIntTypes()

	useUintTypes()

	useFloatTypes()

	useComplexTypes()

	useByteTypes()

	useRuneTypes()

	useUintptrTypes()

	usePointerTypes()

	useInterfaceTypes()

	useStringTypes()

	useBoolTypes()

	useNumericConversions()
}

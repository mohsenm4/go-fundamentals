package main

func main() {
	var a uint64 = 10000000
	var b float64 = 20.5
	var c complex128 = 1 + 2i

	println("Integer:", a)
	println("Float:", b)
	println("Complex:", c)

	var r rune = 'A'
	println("Rune:", r)
	var i32 int32 = r
	println("Int32 from Rune:", i32)

}

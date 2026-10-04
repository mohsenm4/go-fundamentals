package main

//go:noinline
func small() byte {
	buf := make([]byte, 1024) // 1 KB
	buf[0] = 1
	return buf[0]
}

//go:noinline
func big() byte {
	buf := make([]byte, 1<<20) // 1 MB
	buf[0] = 1
	return buf[0]
}

func main() {
	small()
	big()
}

package main

//go:noinline
func returned() []int {
	var arr [4]int
	for i := range arr {
		arr[i] = i + 1
	}
	return arr[:]
}

//go:noinline
func summed() int {
	var arr [4]int
	for i := range arr {
		arr[i] = i + 1
	}
	s := arr[:]
	total := 0
	for _, v := range s {
		total += v
	}
	return total
}

func main() {
	returned()
	summed()
}

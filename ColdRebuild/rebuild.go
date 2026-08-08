package main

// i want to use iota to create a set of constants that represent different rebuild strategies. Each constant will have a unique value, starting from 0 and incrementing by 1 for each subsequent constant.
const (
	StrategyA = iota //0
	StrategyB        // 1
	StrategyC        // 2
	_                // 3
	StrategyE        // 4
)

func main() {

	// i want to print constants StrategyA, StrategyB, StrategyC, and StrategyE
	println(StrategyA, StrategyB, StrategyC, StrategyE)

	var a *int = new(int) // new allocates memory for an int and returns a pointer to it
	*a = 42

	slice := make([]int, 5) // make creates a slice of length 5
	slice[0] = 1

	var i interface{} = nil
	var p *int = nil
	i = p

	if i == nil {
		println("i is nil")
	} else {
		println("i is not nil")
	}

}

package main

import (
	"flag"
	"fmt"
)

func main() {
	deadlock := flag.Bool("deadlock", false, "run intentional nil-channel deadlock")
	flag.Parse()

	if *deadlock {
		runDeadlock()
		return
	}

	runSelectNil()
}

func runSelectNil() {
	a := make(chan int)
	b := make(chan int)

	go func() {
		for _, v := range []int{1, 2, 3} {
			a <- v
		}
		close(a)
	}()

	go func() {
		for _, v := range []int{10, 20, 30} {
			b <- v
		}
		close(b)
	}()

	for a != nil || b != nil {
		select {
		case v, ok := <-a:
			if !ok {
				// If we don't set a closed channel to nil, this case stays
				// permanently ready, keeps returning the zero value, and can
				// burn the loop/fairness of the other cases.
				a = nil
				continue
			}
			fmt.Println("a:", v)

		case v, ok := <-b:
			if !ok {
				// Same reason as above: disable the closed channel.
				b = nil
				continue
			}
			fmt.Println("b:", v)
		}
	}

	fmt.Println("both channels are nil; done")
}

func runDeadlock() {
	var ch chan int

	// A nil channel has no ready cases. With no default case, selectgo
	// blocks. Since there are no other goroutines that can make ch ready,
	// the runtime detects a deadlock and prints:
	//
	// fatal error: all goroutines are asleep - deadlock!
	select {
	case <-ch:
		fmt.Println("received")
	}
}

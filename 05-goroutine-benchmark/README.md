
# Goroutine Benchmark

## Goal

Measure the time and memory cost of creating and keeping 1 million goroutines alive.

## Results

- Goroutines: 1,000,001
- Startup time: 1.014393666s
- HeapAlloc delta: 631,814,536 bytes
- Sys delta: 2,750,198,672 bytes
- StackInuse delta: 2,048,360,448 bytes

## Interpretation

Go can create and keep 1 million goroutines alive in about one second.
The results show that goroutines are lightweight compared to OS threads, but they are not free and still require significant memory.
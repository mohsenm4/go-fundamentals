# 11 — GC: gctrace, GOGC and GOMEMLIMIT

The program keeps a **live set** of ~10 MB (10,000 × 1 KB in a global) and then makes **1,000,000 short-lived 1 KB buffers** (~1 GB of garbage). Both `make` calls escape to the heap (`go build -gcflags="-m"`).

```
go build -o gcdemo . && GODEBUG=gctrace=1 ./gcdemo
```

(Not `GODEBUG=gctrace=1 go run .` — the `go` command and the compiler are Go programs too and print their own `gc` lines.)

## One gctrace line (GOGC=100)

```
gc 51 @0.097s 2%: 0.015+0.21+0.010 ms clock, 0.12+0.005/0.17/0+0.080 ms cpu, 22->23->11 MB, 23 MB goal, 0 MB stacks, 0 MB globals, 8 P
```

| Field | Meaning |
|-------|---------|
| `gc 51` | cycle number |
| `@0.097s` | time since the program started |
| `2%` | share of CPU spent on GC since the start |
| `0.015+0.21+0.010 ms clock` | wall time: sweep termination (**STW**) + concurrent mark + mark termination (**STW**) |
| `0.12+0.005/0.17/0+0.080 ms cpu` | CPU time: STW + mark (assist / background / idle) + STW — assist `0.005` = a little mark assist |
| `22->23->11 MB` | heap when GC started → heap when mark ended → **live** heap after mark |
| `23 MB goal` | ≈ live 11 MB × (1 + 100/100) |
| `0 MB stacks`, `0 MB globals` | scannable stacks and globals (counted in the goal) |
| `8 P` | number of Ps (GOMAXPROCS) |

## GOGC and GOMEMLIMIT

| Setting | Guess: GC count | Real: GC count | Total time | Peak heap |
|---------|-----------------|----------------|------------|-----------|
| `GOGC=50` | ≈ 200 | 169 | ~165 ms | 19 MB |
| `GOGC=100` | ≈ 100 | 85 | ~107 ms | 28 MB |
| `GOGC=200` | ≈ 50 | 42 | ~96 ms | 38 MB |
| `GOGC=off` | 0 | 0 | ~50–80 ms | grows to ~1 GB (never freed) |
| `GOGC=off GOMEMLIMIT=50MiB` | ≈ 25 | ~32 | ~90 ms | 42 MB |

**Why**: with a higher GOGC, Go lets the heap grow more and runs the GC later, so there are fewer GC cycles and less CPU overhead, but more memory is used.

**The last row**: GOMEMLIMIT counts **all** Go memory (heap + runtime + stacks, ~8 MB here), not only the heap. So the heap goal was 42 MB, not 50, each cycle freed ~31 MB instead of 40, and there were ~32 GCs instead of 25.

# 10 — Escape Analysis

Rule: write the guess **before** running `go build -gcflags="-m -m"`. Never change a guess after the check.

| # | Case | My guess | Compiler says | Why (1 sentence) | ✅/❌ |
|---|------|----------|---------------|------------------|-------|
| 1 | return pointer to local | heap | `moved to heap: x` (flow: `~r0 ← &x`) | `x` is returned by `return &x`, so it must live after `createUser` returns. | ✅ |
| 2 | pointer passed down only | stack | `p does not escape` (nothing about `x` → stays on stack) | `x` is only shared down to `addOne`, and nobody needs it after `addOne` returns. | ✅ |
| 3 | `fmt.Println(x)` | heap | `42 escapes to heap` · `... argument does not escape` | `Println` takes `...any`; the value is boxed into an interface and flows into `Fprintln`, which the compiler can't prove keeps it local. The `[]any` slice itself stays on the stack. | ✅ |
| 4 | closure returned, captures `count` | heap | `moved to heap: count` · `func literal escapes to heap` · `capturing by ref: count (assign=true)` | `count` must stay alive after `makeCounter` returns because the returned closure still uses it; the closure itself escapes too (returned via `~r0`). | ✅ |
| 5a | `make([]int, 10)` constant size, not returned | stack | `make([]int, 10) does not escape` | Size is known at compile time and the slice never leaves `fixed`. | ✅ |
| 5b | `make([]int, n)` variable size, not returned | heap | `make([]int, n) does not escape` | Surprise: since Go 1.25 the compiler reserves a small stack buffer (32 bytes) and decides at **run time** — fits → stack, too big → heap. "Does not escape" ≠ "never on heap". | ❌ |

## Open questions (for later)

- **Program 6–20 idea:** if the closure only *reads* `count` (no `count++`), is it still `capturing by ref`? Does `count` still move to heap?
- **Wednesday benchmark:** program 3 says `42 escapes to heap`. Does `fmt.Println(42)` really allocate at run time? Measure `allocs/op` with `-benchmem`.
- **Wednesday benchmark:** program 5b — compare `allocs/op` for `variable(3)` (24 bytes) vs `variable(10)` (80 bytes). Same "does not escape", different result at run time?

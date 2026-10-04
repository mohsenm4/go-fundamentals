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
| 6 | closure in `go func()` uses local `wg` and `msg` | heap (gut feeling, no clear reason) | `func literal escapes to heap` · `moved to heap: wg` · `capturing by value: msg` (no "moved to heap" for `msg`) · `"hello" escapes to heap` | The new goroutine runs on its **own** stack, so the closure can't live in `run`'s frame → heap; `wg` is captured **by ref** (`Done` needs `&wg`) so it follows the closure to the heap, but `msg` is only read, so it is **copied** into the closure and the variable itself stays on the stack. | 🟡 |
| 7a | `make([]byte, 1024)` constant, small, not returned | stack | `make([]byte, 1024) does not escape` | Size is known and small, and the slice never leaves `small`. | ✅ |
| 7b | `make([]byte, 1<<20)` constant, 1 MB, not returned | heap ("maybe, because it is very big") | `make([]byte, 1048576) escapes to heap` | Nothing leaves the function, but the size is over the limit for implicit stack allocations (64 KB for `make`/`new`/`&T{}`), so the compiler puts it on the heap. | ✅ |
| 8a | `make([]int, 0, 4)` start of a growing slice | stack | `make([]int, 0, 4) does not escape` | Small, constant size (32 bytes), and the slice never leaves `grow`. | ✅ |
| 8b | new array when `append` grows past cap 4 | heap | `append does not escape` | Trap: "does not escape" only means the slice doesn't leave the function. When `append` runs out of room, `runtime.growslice` makes the new array on the **heap**. Measured (`testing.AllocsPerRun`, Go 1.27): n=4 → 0 allocs, n=5 → 1, n=9 → 2, n=100 → 5 (one per growth: 8, 16, 32, 64, 128). | ✅ |
| 9a | `make(map[string]*int)` local map, not returned | stack | `make(map[string]*int) does not escape` | The map header never leaves `fill`, so the compiler can keep it local. | ✅ |
| 9b | `x` whose address is stored in that map | heap ("because it's a pointer" — wrong reason) | `moved to heap: x` | Taking `&x` alone is fine (program 2). The real reason: the compiler does **not** track what is inside a map, so any pointer stored into a map (or through any indirect store like `s[i] = &x`, `*p = &x`) is treated as "goes to the heap". | ✅ |
| 10a | send value `ch <- x` | stack | nothing about `x` (stays on stack) | Sending a value sends a **copy**; nobody shares `x`. | ✅ |
| 10b | send pointer `ch <- &y` | heap | `moved to heap: y` | The pointer goes to whoever receives — maybe another goroutine with its own stack — so `y` is **shared** → heap. | ✅ |
| 10c | the channel itself `make(chan T, 1)` | stack | **nothing printed** — but always heap | Trap: `-m` is silent, because a channel is never a stack candidate. `make(chan)` → `runtime.makechan` → `mallocgc`, always heap. Measured: `byValue` 1 alloc (hchan + buffer in one call, `int` has no pointers), `byPointer` 3 allocs (hchan, buffer separately because `*int` has pointers, and `y`). See `makechan` in `src/runtime/chan.go`. | ❌ |
| 11a | `global = &x` (store address in a package-level var) | heap — "the pointer is given to a global" | `moved to heap: x` | A global lives for the whole program, so anything it points to must outlive `save` → heap. | ✅ |
| 11b | `p := &y` used only locally, in a function that also **reads** `global` | stack | nothing about `y` (stays on stack) | Only the flow of `y`'s **own** address matters. `p` is only dereferenced, never stored; reading `*global` has no effect on `y`. | ✅ |
| 12a | `p := new(point)` used only inside the function | heap (no reason — "`new` sounds like heap") | `new(point) does not escape` | In Go, `new` does **not** choose heap (unlike Java/C++). Escape analysis decides; `p`'s address never leaves `withNew` → stack. Measured: 0 allocs. | ❌ |
| 12b | `q := &point{...}` used only inside the function | heap | `&point{...} does not escape` | Same: `&` alone means nothing. Only "where does the address go?" matters → nowhere → stack. Measured: 0 allocs. | ❌ |

## Note — why goroutines force the heap (program 6)

- Each goroutine has its **own stack**. `run` lives on stack A; the `go func()` runs on stack B.
- Stack B must not point into stack A, for two reasons:
  1. `run` may return before the goroutine finishes. Then the frame on stack A is dead → dangling pointer.
  2. Stacks **move**. When a stack grows, Go copies it to a bigger place and fixes pointers *inside that stack only*. A pointer from stack B into stack A would not be fixed.
- So anything shared between two goroutines must live on the **heap** (the one place both can safely reach).
- Capture rule: a variable that is **never changed** after capture and is small (≤ 128 bytes) is captured **by value** (copied). Otherwise **by ref** → it moves to the heap. (Program 4: `count++` → by ref. Program 6: `msg` only read → by value.)

## Open questions (for later)

- ✅ **Answered by program 6** (`msg` only read → `capturing by value`): if the closure only *reads* `count` (no `count++`), is it still `capturing by ref`? Does `count` still move to heap?
- **Wednesday benchmark:** program 3 says `42 escapes to heap`. Does `fmt.Println(42)` really allocate at run time? Measure `allocs/op` with `-benchmem`.
- **Wednesday benchmark:** program 5b — compare `allocs/op` for `variable(3)` (24 bytes) vs `variable(10)` (80 bytes). Same "does not escape", different result at run time?

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
| 13a | interface var used locally (`rect` value) | stack (call is devirtualized, type is known) | `devirtualizing s.area to rect` · `rect{...} does not escape` · measured: 0 allocs | `s` can only ever hold a `rect`, so the compiler turns `s.area()` into a direct call to `rect.area`; no unknown code touches the value, so it stays in `local`'s frame. | ✅ |
| 13b | interface passed to `measure(s shape)` | heap (`s.area()` is a call to an unknown method) | `leaking param: s` (in `measure`) · `r escapes to heap` (not `moved to heap: r`) · measured: 1 alloc | Inside `measure`, `s.area()` could call *any* `area` method, so the compiler assumes the data behind `s` leaks; `passed` boxes a **copy** of `r` on the heap, and the variable `r` itself stays on the stack. | ✅ |
| 14a | own `...any` func, only reads args | stack — both `x` and the `[]any` | `args does not escape` · `... argument does not escape` · `x does not escape` · measured: 0 allocs | `count` is my own code, so the compiler can see it only reads `args`; the `[]any` and the boxed `x` both stay in `callCount`'s frame. In program 3, `Fprintln` is too complex for the compiler to prove that, so `42` escaped. | ✅ |
| 14b | own `...any` func, stores args in a global | heap — both `x` and the `[]any` | `leaking param: args` · `... argument escapes to heap` · `x escapes to heap` · measured: 2 allocs | `kept = args` puts the slice in a global, so the backing array (1 alloc) and the boxed `x` it points to (1 alloc) must outlive `callKeep`. In program 3 only the value escaped; here the slice escapes too. | ✅ |
| 15a | `defer` closure with `n++` | stack — closure and `n` | `capturing by ref: n` · `func literal does not escape` (no `moved to heap: n`) · measured: 0 allocs | A single top-level `defer` is open-coded (runs inline at function exit), so the closure and the `n` it captures by ref both stay on the stack. | ✅ |
| 15b | `defer` closure with `n++` inside a `for` | heap — closure and `n` | `func literal escapes to heap` · `capturing by ref: n` · `moved to heap: n` · measured: 4 allocs | The compiler can't know at compile time how many defers a loop makes, so each one goes on the runtime's defer list and its closure escapes; the closures share `n` by ref, so `n` moves to the heap too (1 for `n` + 3 closures = 4). | ✅ |
| 16a | `m[string(b)]` lookup only | no copy at all | `string(b) does not escape` · `b does not escape` · measured: 0 allocs · asm: no `runtime.slicebytetostring` call | For `m[string(b)]` the compiler uses the bytes of `b` directly as the key for the lookup, so no string is built, not even on the stack. | ✅ |
| 16b | `s := string(b)`, returned | copy, on heap | `string(b) escapes to heap` · `b does not escape` · measured: 1 alloc (`runtime.slicebytetostring`) | A string is immutable, but `b` can change later, so the conversion **must** copy the bytes; the copy is returned, so it lives on the heap (and `b` itself does not escape). | ✅ |
| 17a | `return arr[:]` (local array) | heap | `moved to heap: arr` (flow: `~r0 ← &arr`) | `arr[:]` is a slice pointing **into** `arr`, so returning it is the same as returning `&arr` (program 1). | ✅ |
| 17b | `arr[:]` used only locally | stack | nothing about `arr` (stays on stack) | The slice points into `arr`, but it never leaves `summed`, so `arr` can stay in the frame. | ✅ |
| 18a | local `buf` passed to `io.Reader.Read` | heap | `leaking param: r` · `make([]byte, 64) escapes to heap` · measured: 1 alloc | The compiler can't see which `Read` runs behind `io.Reader`, and an unknown `Read` could keep `buf`, so it must assume `buf` escapes. Same reason as 13b. | ✅ |
| 18b | local `buf` passed to `*strings.Reader.Read` | stack | `r does not escape` · `make([]byte, 64) does not escape` · measured: 0 allocs | With the concrete type the compiler sees `(*strings.Reader).Read`, which only `copy`s into `buf` and never keeps it, so `buf` stays on the stack. | ✅ |
| 19a | return `box{p: &x}` by value | heap | `moved to heap: x` (flow: `~r0 ← &x` from `box{...}` struct literal element) | The struct is copied out by value, but the copy still holds `&x`, so `x` must outlive `makeBox`. A pointer inside a returned value is the same as returning the pointer. | ✅ |
| 19b | return `plain{v: x}` by value | stack | nothing about `x` (stays on stack) | `plain{v: x}` copies the **value** of `x`, so nobody needs `x` after `makePlain` returns. | ✅ |
| 20a | `&buf{}` used only locally | stack | `&buf{} does not escape` · measured: 0 allocs | Same as 12b: the 64-byte struct's address never leaves `localOnly`, so it stays on the stack. | ✅ |
| 20b | `&buf{}` then `pool.Put(b)` | heap | `&buf{} escapes to heap` (flow: `b (interface-converted)` → `(*sync.Pool).Put`) · measured: 1 alloc | `Put` keeps the object inside the pool to hand it to a later `Get`, maybe on another goroutine, so `b` must be on the heap. In this program, the pool doesn't save this alloc; it creates it. | ✅ |

## Note — why goroutines force the heap (program 6)

- Each goroutine has its **own stack**. `run` lives on stack A; the `go func()` runs on stack B.
- Stack B must not point into stack A, for two reasons:
  1. `run` may return before the goroutine finishes. Then the frame on stack A is dead → dangling pointer.
  2. Stacks **move**. When a stack grows, Go copies it to a bigger place and fixes pointers *inside that stack only*. A pointer from stack B into stack A would not be fixed.
- So anything shared between two goroutines must live on the **heap** (the one place both can safely reach).
- Capture rule: a variable that is **never changed** after capture and is small (≤ 128 bytes) is captured **by value** (copied). Otherwise **by ref** → it moves to the heap. (Program 4: `count++` → by ref. Program 6: `msg` only read → by value.)

## Note — an interface call is a wall (programs 13 + 18)

- Escape analysis works on code the compiler can **see**. Behind an interface it can't see which method will run, so anything passed to an interface method call is assumed to escape. That includes the receiver (13b: `s`) and the arguments (18a: `buf`).
- There are two ways around the wall:
  1. **Devirtualization** (13a): if the compiler can prove the concrete type, it calls the method directly.
  2. **A concrete type in the signature** (18b): `*strings.Reader` instead of `io.Reader`.
- Trade-off: `io.Reader` in a hot path costs one heap alloc per buffer. That's why `bufio` and `io.CopyBuffer` keep one buffer and reuse it.

## Note — boxing a constant into `any` (program 14 vs program 3)

- Program 14 uses `x := n * 1000` on purpose, not a constant.
- With a constant, `-m` still says `42 escapes to heap`, but the interface points at a read-only static copy, so the box costs no allocation. Measured: `keep(42)` = 1 alloc (only the `[]any`); `keep(x)` with a non-constant `x` = 2 allocs.

## Open questions (for later)

- ✅ **Answered by program 6** (`msg` only read → `capturing by value`): if the closure only *reads* `count` (no `count++`), is it still `capturing by ref`? Does `count` still move to heap?
- **Wednesday benchmark:** program 3 says `42 escapes to heap`. Does `fmt.Println(42)` really allocate at run time? Measure `allocs/op` with `-benchmem`. (Partial hint from program 14: boxing a constant uses static data and does **not** allocate. `Fprintln` itself is still unmeasured.)
- **Wednesday benchmark:** program 5b — compare `allocs/op` for `variable(3)` (24 bytes) vs `variable(10)` (80 bytes). Same "does not escape", different result at run time?

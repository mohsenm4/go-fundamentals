# Go Language Specification — Notes

Reference: https://go.dev/ref/spec

---

## 1. Introduction

- Go = **general-purpose**, systems-oriented, **strongly typed**, **garbage-collected**, native concurrency.
- Programs built from **packages** (dependency unit, not file).
- Syntax compact + easy to parse → fast tooling (`gofmt`, `gopls`, `go vet`).
- Concurrency primitives: **goroutine** (lightweight thread, runtime-managed) + **channel** (safe communication).

---

## 2. Notation (EBNF)

Grammar written in a variant of **Extended Backus-Naur Form**:

```
Syntax      = { Production } .
Production  = production_name "=" [ Expression ] "." .
Expression  = Term { "|" Term } .
```

Operators (increasing precedence): `|` alternation, `()` group, `[]` optional (0 or 1), `{}` repetition (0..n).

- **lowercase** name → lexical token (terminal).
- **CamelCase** name → non-terminal.
- `""` / `` ` `` → literal.
- `a … b` → char range. `…` (single char) ≠ `...` (Go token).
- `[Go 1.xx]` tag → feature requires that version.

---

## 3. Source Code Representation

- Source = **Unicode text, UTF-8 encoded**.
- **Not canonicalized**: `é` (U+00E9) ≠ `e` + combining accent (U+0065 + U+0301) — distinct code points.
- Case-sensitive.
- Implementation may forbid `NUL` (U+0000); may ignore leading BOM (U+FEFF).
- `newline` = U+000A (LF only).
- Unicode categories used: Lu, Ll, Lt, Lm, Lo = letter; Nd = digit. Based on **Unicode 8.0**.
- `_` (U+005F) counts as lowercase letter.
- `letter = unicode_letter | "_"`.

---

## 4. Lexical Elements

### 4.1 Comments
- Line: `// ...` → ends at newline.
- General: `/* ... */` → first `*/` closes; **do not nest**; cannot start inside string/rune/comment.
- General comment with no newline → acts like **space**; otherwise like **newline** (affects semicolon insertion).

### 4.2 Tokens (4 classes)
Identifiers, keywords, operators/punctuation, literals.
- Whitespace ignored except as separator.
- **Longest match rule**: `a+++b` = `a ++ + b`.

### 4.3 Semicolons (auto-inserted)
Semicolon inserted at end of line if last token is:
- identifier, literal (int/float/imag/rune/string)
- keyword: `break`, `continue`, `fallthrough`, `return`
- operator: `++`, `--`, `)`, `]`, `}`

Semicolon may be omitted before closing `)` / `}`.

Consequence: **opening `{` must be on same line as function/if/for header** — otherwise auto-`;` breaks the code.

### 4.4 Identifiers
- `identifier = letter { letter | unicode_digit }` — must start with letter (or `_`).
- **Uppercase first letter → exported** (public from package).
- Some names are **predeclared** (e.g., `int`, `nil`, `true`).

### 4.5 Keywords (25 total — reserved)
```
break     default      func         interface   select
case      defer        go           map         struct
chan      else         goto         package     switch
const     fallthrough  if           range       type
continue  for          import       return      var
```

### 4.6 Operators & Punctuation
Includes `+ - * / %`, bit ops `& | ^ << >> &^`, compound assign, comparison, logic `&& ||`, `<-` (channel), `:=`, `...`, `~` (Go 1.18, for generics constraints), `&^` (AND-NOT / bit clear).

### 4.7 Integer Literals
Prefixes: none = decimal; `0b`/`0B` = binary; `0`, `0o`, `0O` = octal; `0x`/`0X` = hex.
- `_` allowed between digits or after base prefix (Go 1.13). Must **separate** digits — cannot be leading/trailing/doubled.
- `_42` is an identifier, not a literal.

### 4.8 Floating-Point Literals
- Decimal: `72.40`, `.25`, `1e6`, `6.67428e-11` — integer or fraction may be omitted; point or exponent may be omitted.
- Hex (Go 1.13): `0x1p-2` = 0.25; **`p`/`P` exponent required** (because `e` is a hex digit). Scales by `2^exp`.

### 4.9 Imaginary Literals
- Int or float + suffix `i` → complex imaginary part: `2.5i`, `3+4i`.
- **Backward compat**: `0123i` = `123i` (not octal).

### 4.10 Rune Literals
- Single Unicode code point in `' '`. Type = `rune` (alias for `int32`).
- Escapes:
  - `\x` + 2 hex digits (byte value)
  - `\u` + 4 hex digits
  - `\U` + 8 hex digits
  - `\` + 3 octal digits (0–255)
  - Special: `\a \b \f \n \r \t \v \\ \' \"`
- Invalid: `'aa'`, `'\k'`, `'\xa'`, surrogate halves (`\uDFFF`), code points > `0x10FFFF`.

### 4.11 String Literals — 2 forms
- **Raw** `` `...` `` — no escapes, may span lines, `\r` stripped, no backticks allowed inside.
- **Interpreted** `"..."` — escapes processed; no newlines; `\"` legal, `\'` illegal.
- **Byte vs char distinction**:
  - `\xFF` / `\377` → single byte 0xFF.
  - `ÿ` / `ÿ` → **two bytes** (UTF-8 encoding: `\xc3\xbf`).
- `len("é")` may be 2 (bytes, not runes).

---

## 5. Constants

### 5.1 Kinds (6)
Boolean, rune, integer, floating-point, complex, string. Middle four = **numeric constants**.

### 5.2 Sources of Constants
Literals, constant identifiers, constant expressions, conversions with constant result, some built-ins: `min`, `max`, `unsafe.Sizeof`, `cap`, `len`, `real`, `imag`.

### 5.3 Typed vs Untyped
- Literals, `true`, `false`, `iota`, and expressions of only-untyped operands → **untyped**.
- Type assigned via explicit declaration/conversion, or implicitly when used in typed context.

### 5.4 Default Types (used when typed context is required)
| Kind | Default |
|------|---------|
| bool | `bool` |
| rune | `rune` |
| int | `int` |
| float | `float64` |
| complex | `complex128` |
| string | `string` |

Untyped constants are flexible:
```go
const Pi = 3.14
var f float32 = Pi   // OK
var d float64 = Pi   // OK
```
Typed constants are not:
```go
const Pi float64 = 3.14
var f float32 = Pi   // ERROR
```

### 5.5 Arbitrary Precision
- Numeric constants have **infinite precision at compile time**; **no overflow**.
- No IEEE-754 negative-zero, Inf, or NaN as constants.
- Error only when assigned to a type that can't hold the value.
- **Implementation minimums**: integer ≥ 256 bits; float mantissa ≥ 256 bits, exponent ≥ 16 bits.

### 5.6 Constant Declaration
```
ConstDecl = "const" ( ConstSpec | "(" { ConstSpec ";" } ")" ) .
ConstSpec = IdentifierList [ [ Type ] "=" ExpressionList ] .
```

Examples:
```go
const Pi float64 = 3.14159
const zero = 0.0                 // untyped
const a, b, c = 3, 4, "foo"
const (
    size int64 = 1024
    eof        = -1              // untyped
)
```

Within a `const (...)` block, if a ConstSpec omits the expression list, the previous non-empty list is repeated.

### 5.7 `iota`
- Predeclared identifier — represents **index of ConstSpec inside a `const` block**, starting at 0.
- Resets on each new `const (...)` block.
- Multiple `iota` in same ConstSpec share the same value.
- Increments even when unused (skipped lines still advance).
- Common patterns:
  ```go
  // enum
  const (
      Sunday = iota; Monday; Tuesday; ...
  )

  // bit flags
  const (
      Read    = 1 << iota  // 1
      Write                // 2
      Execute              // 4
  )
  ```

---

## 6. Variables

### 6.1 Concept
- Storage location holding a **value** of a given **type**.
- Allocated via declaration, function params/returns, `new()`, or address of composite literal.
- Elements of arrays/slices/structs are individually addressable — each acts like a variable.

### 6.2 Static vs Dynamic Type
- **Static type** = declared type. Never changes.
- **Dynamic type** = only relevant for **interface** variables; the actual concrete type of the stored value. May change at runtime. `nil` has no dynamic type.

```go
var x interface{}   // static: interface{}, nil
x = 42              // dynamic: int
x = "hi"            // dynamic: string
```

### 6.3 Zero Value
Uninitialized variable = **zero value** of its type:
- numeric → `0`
- bool → `false`
- string → `""`
- pointer, slice, map, chan, func, interface → `nil`
- struct → all fields zero
- array → all elements zero

### 6.4 Variable Declaration
```
VarDecl = "var" ( VarSpec | "(" { VarSpec ";" } ")" ) .
VarSpec = IdentifierList ( Type [ "=" ExpressionList ] | "=" ExpressionList ) .
```

Forms:
```go
var i int                       // zero value
var x int = 42                  // explicit type + value
var x = 42                      // type inferred
var a, b, c int                 // multi, same type
var name, age = "Ali", 30       // multi, mixed
var (
    i       int
    u, v, s = 2.0, 3.0, "bar"
)
```

## Zero values and nil semantics

- **Slice**: `var s []int` (nil) vs `s := []int{}` (empty) — both `len==0`, `cap==0`, both `range` zero times, both `append` identically (fresh backing array on first append). Only real differences: `s == nil` is `true` for nil / `false` for empty (empty has a valid non-nil pointer to a zero-size array), and JSON marshaling — nil → `null`, empty → `[]`.
- **Map**: reading from nil map is safe, returns zero value; `len==0`; `delete` is a safe no-op. Writing panics: "assignment to entry in nil map". Must `make()` before any write.
- **Channel**: send/receive on nil channel blocks forever (not a panic — actually useful in `select` to disable a case). `close(nilChan)` panics: "close of nil channel". `len`/`cap` both 0.
- **Interface**: holding a nil pointer of a concrete type makes the interface itself non-nil — `interface = (type, value)`, type part is set even when value is nil. Classic trap: `var p *T = nil; var i error = p; i != nil` → `true`.


### 6.5 Rules
- No initializer → zero value.
- No type → type taken from initializer; **untyped constants converted to default type**; untyped bool → `bool`.
- `nil` cannot infer type: `var n = nil` is ILLEGAL — needs explicit type.
- **Implementation restriction**: unused local variable = **compile error**. Package-level vars may be unused.

### 6.6 Short Variable Declaration `:=`
```
ShortVarDecl = IdentifierList ":=" ExpressionList .
```
Equivalent to `var ... = ...` with inferred types. **Only inside functions.**

```go
i, j := 0, 10
f := func() int { return 7 }
ch := make(chan int)
r, w, _ := os.Pipe()
```

### 6.7 Redeclaration Rule
`:=` may **redeclare** an existing variable if:
- It's in the **same block**.
- Same type.
- **At least one** non-blank name on the LHS is **new**.

Redeclaration does not create a new variable — it reassigns. Common idiom for error handling:
```go
data, err := readFile("a")
data, err = readFile("b")      // = (both exist)
data2, err := readFile("c")    // := (data2 new, err reused)
```

Invalid: `x, y, x := 1, 2, 3` (duplicate LHS name).

### 6.8 Special Contexts
`:=` allowed in initializer clauses of `if`, `for`, `switch`. Scope of the declared variable = that statement block:
```go
if err := doSomething(); err != nil { ... }
for i := 0; i < n; i++ { ... }
switch v := x.(type) { ... }
```

---

## Quick Comparison

| Tool    | When       | In function | At package | Mutable |
|---------|-----------|-------------|------------|---------|
| `const` | compile-time | ✅         | ✅         | ❌       |
| `var`   | runtime   | ✅          | ✅         | ✅       |
| `:=`    | runtime   | ✅          | ❌         | ✅       |

---

## builtin.go — Key takeaways
- **Aliases vs distinct types**: only `byte = uint8`, `rune = int32`, and `any = interface{}` use `=` — they are fully interchangeable. `int`, `uint`, `string` are declared as `type int int` etc., so they are **distinct types** — `int` is NOT an alias for `int32` even on 64-bit systems, and needs explicit conversion.
- **`make` vs `new`**: `make(T, ...)` returns a ready-to-use value of type T (only for slice/map/chan). `new(T)` returns `*T` pointing to a zero-value of any type. Rule of thumb: use `make` when the type needs internal init (header, buckets, buffer); use `new` when you just want a zero-valued T on the heap.
- **`comparable` is constraint-only**: `type comparable interface{ comparable }` — a special interface usable only as a **type parameter constraint** (`func Eq[T comparable](a, b T) bool`), never as a variable/field type. Same file also documents `error` as just `interface{ Error() string }` — nothing magical, only a convention.
- **File is documentation-only**: `builtin.go` declares nothing real. The compiler injects builtins into every package via **universe scope**. Every `type X X` line is a placeholder so `godoc` can render docs for language-level identifiers.
- **Not every builtin works on every type**: `make` is only for slice/map/chan. `close` is only for channels (and only the sender should call it). `len`/`cap` work for arrays, pointer-to-array, slices, maps, strings, channels.
- **`len` on strings counts bytes, not characters**: e.g., `len("سلام") == 8`, not 4. For characters, use `[]rune(s)` (Unicode code points). For raw bytes, use `[]byte(s)`. Remember: `byte = uint8` (raw bytes), `rune = int32` (Unicode).
- **Why `delete` returns nothing but `append` returns a slice**: maps are reference types — the variable is a pointer to a runtime hash table, so `delete` mutates in place. Slices are value types with a `{pointer, len, cap}` header — `append` may need a new header, so it must return one. Always write `s = append(s, x)`.

_Read: 2026-08-08_

---

## io.go — Reader, Writer, and composition

This file is one of the most interesting and useful packages in Go.

This part is about the `Reader` and `Writer` interfaces, and how they combine into bigger interfaces like `ReadWriter`. This composition pattern is used widely across Go.

- **Read**: you create a buffer (must be `[]byte`). The `Read` method fills your buffer with data from a source (file, network, memory, ...). Returns `n` (how many bytes were filled) and `err` (or `io.EOF` when the source ends).
- **Write**: you give a buffer with data to the method, and it writes those bytes to a destination (file, network, buffer, ...).

**Composition example** — `ReadWriter` has no methods of its own; it embeds `Reader` and `Writer`:
```go
type ReadWriter interface {
    Reader
    Writer
}
```
Any type with both `Read(p []byte)` and `Write(p []byte)` methods automatically satisfies `ReadWriter`. Small interfaces + composition = flexible design.

### Copy family and wrapper types

This part of the file shows how much you can build on top of the tiny `Reader` / `Writer` interfaces. `Copy` and `CopyN` are the practical, everyday helpers — both funnel into `copyBuffer`, which uses `WriterTo` / `ReaderFrom` when available for a big performance win (for `*os.File`, that path uses `sendfile` and skips user-space entirely). On top of the basic interfaces, `io` ships wrapper types: `SectionReader` exposes a slice of a `ReaderAt` as its own `Reader` + `Seeker` + `ReadAt`; `OffsetWriter` is its mirror image for `WriterAt`; and `TeeReader` forwards every read into a second `Writer` — great for logging or hashing without changing the surrounding code. Finally there are utility values like `Discard` (a black-hole `Writer` that reuses buffers via `sync.Pool`) and `NopCloser` (adds a no-op `Close` to any `Reader`, and carefully forwards `WriteTo` if the underlying reader has one, so wrapping does not kill the fast path). The theme is the same throughout: two tiny interfaces, endless composition.

_Read (lines 60–153): 2026-08-09_
_Read (lines 329–750): 2026-08-10_

### pipe.go and multi.go — bridging and combining streams

`pipe.go` and `multi.go` live next to `io.go` and finish the picture of what the `io` package offers. `io.Pipe` is the only tool in the stdlib that converts direction: it hands out a `PipeReader` and `PipeWriter` sharing one unbuffered, synchronous, in-memory channel — so code that writes and code that reads can meet in the middle without a buffer in between. Every `Write` blocks until a matching `Read` copies bytes out; this gives backpressure for free and enables true streaming between an API that demands a `Writer` (e.g. `json.Encoder`, `gzip.Writer`) and one that demands a `Reader` (e.g. `http.Post`). Internally it uses two channels (`wrCh` for the write buffer, `rdCh` for the byte count acknowledgement) plus a `sync.Once` around `close(done)` so either side can close idempotently, and a small `onceError` helper that stores only the first error while allowing many reads. The `multi.go` file is the opposite: no goroutines, no synchronization — just composition. `MultiReader` concatenates readers sequentially (falling through to the next on EOF, and flattening nested `multiReader`s to keep the call chain shallow), and `MultiWriter` fans a single write out to many writers in order, stopping at the first error. Together, `Pipe` bridges producers to consumers, `MultiReader` glues sources end-to-end, `MultiWriter` broadcasts to sinks — three tiny APIs that show how far you can go with the two-method `Reader`/`Writer` interfaces at the core of `io.go`.

_Read (pipe.go + multi.go): 2026-08-11_

---

## runtime/slice.go — how slices really work

A slice is a 24-byte header: `array unsafe.Pointer`, `len int`, `cap int`. The data lives elsewhere. That single fact drives everything else.

### slice header + aliasing

Copying a slice (assignment or function call) copies only the header, not the backing array. Two slices sharing the same array will see each other's writes. A callee receives its own header copy — it cannot change the caller's `len`/`cap`, but it can absolutely rewrite the shared array. `append` inherits this: if `len < cap` the write goes into the caller's array *while* the caller keeps seeing the old `len`. Classic Go footgun and a real source of data races. Rule: always `s = append(s, x)`, and use `s[:len(s):len(s)]` to cap sub-slices you hand out.

### makeslice — allocation with a safety net

`make([]T, len, cap)` lowers to `runtime.makeslice`, which returns just an `unsafe.Pointer` (compiler already knows `len`/`cap` — no point returning a full header). Before allocating it multiplies `element_size * cap` through `math.MulUintptr`, which reports overflow. A silent overflow would turn `make([]int64, MaxInt64)` into an 8-byte allocation with a slice header claiming quintillions of slots — memory corruption. Go turns that into a clean panic. The message even distinguishes "len" vs "cap" by re-running the check with just `len`.

Two subtleties around this function: (1) `//go:linkname makeslice` exists because external libs (`bytedance/sonic`, `cloudwego/dynamicgo`, `ugorji/codec`) reach into the private symbol; the runtime keeps a "hall of shame" comment as social pressure — Go can't make it public without freezing the API forever. (2) Panic paths live in separate tiny functions (`panicmakeslicelen`, `panicmakeslicecap`) so `makeslice` stays small enough for the compiler to inline it. Cold path out of the body → hot path stays fast.

### growslice + nextslicecap — capacity growth

When `append` runs out of room, `growslice` calls `nextslicecap` for the new capacity. The rule is a hybrid:
- `newLen > 2*oldCap` → just return `newLen`
- `oldCap < 256` → double
- otherwise → grow by ~1.25× per iteration until `newcap >= newLen`

The 1.25× formula `newcap += (newcap + 3*threshold) >> 2` is `newcap * 1.25 + 192` in disguise (`>> 2` = fast `/4`). The `+192` term smooths the transition around 256 so growth doesn't jerk from 2× to 1.25× at one point. Rationale: small slices favour speed (double, allocate rarely); large slices favour memory (doubling a 10M-element slice wastes a lot). Micro-optimization worth noticing: `doublecap := newcap + newcap` (add is a touch cheaper than mul in a hot function).

### roundupsize — respecting size classes

After `nextslicecap`, `growslice` calls `roundupsize`. One fact about Go's allocator: `mallocgc` doesn't hand out arbitrary sizes — only a fixed list of **size classes** (8, 16, 24, 32, 48, 64, 80, 96, 112, 128, 144, ...). Ask for 17 bytes, get 24. This "internal fragmentation" is the cost of avoiding external fragmentation — the same trick tcmalloc/jemalloc/mimalloc use. It makes allocation O(1) via per-class free-lists.

`roundupsize` answers: "if I ask for `size` bytes, what will `mallocgc` actually give me?" Two lookup-table hops for small allocations (`SizeToSizeClass8` / `SizeToSizeClass128` → `SizeClassToSize`), or page-align for large (>= 32KB). Pure arithmetic on tables generated by `mksizeclasses.go` — no allocation happens.

`growslice` uses this to bump `cap` up to whatever the allocator was going to give anyway. If the formula said cap=16 but the nearest size class holds 18 slots, Go sets cap=18 — two free slots. This is why `cap(s)` after a grow often isn't the round number the doubling rule predicts.

### Takeaways
- Slice = 24-byte header. Everything else follows.
- `append` returns a (possibly new) header — never mutates the caller's.
- Aliasing + `cap > len` is the sharpest edge; use 3-index slicing to cap sub-slices.
- Growth: 2× small, ~1.25× large, smooth transition. `roundupsize` bumps further to fit size classes.
- Pre-size with `make([]T, 0, n)` when the final size is known — skips every grow.

_Read (slice.go): 2026-08-12_

---

## errors — package basics

The whole `errors` package is 91 lines. It exposes one function (`New`), one struct (`errorString`), one method (`Error`), and one sentinel value (`ErrUnsupported`). Everything else — `Is`, `As`, `Unwrap`, `Join` — lives in `wrap.go` and builds on this foundation.

### the error interface

Declared in `builtin.go`, not `errors.go`:

```go
type error interface {
    Error() string
}
```

Any type with an `Error() string` method automatically satisfies `error` — Go's implicit interface satisfaction at its simplest.

### errorString and why pointer matters

```go
type errorString struct{ s string }
func (e *errorString) Error() string { return e.s }
func New(text string) error { return &errorString{text} }
```

Two things are deliberate here: the pointer receiver, and the `&` in `New`. Together they guarantee that every error returned by `New` has a **unique heap address**. This is what makes sentinel errors like `io.EOF` work:

```go
a := errors.New("EOF")
b := errors.New("EOF")
a == b   // false — different addresses
a == io.EOF  // false — different addresses
```

If `New` returned a value (`errorString{text}` without `&`), interface equality would compare struct contents instead of pointer identity, and two independently-created "EOF" errors would compare equal — breaking every sentinel check in the standard library.

**Takeaway**: pointer receiver here is not about mutation, it's about **identity**.

### the sentinel error pattern

```go
var ErrUnsupported = errors.New("unsupported operation")
```

One instance per package, referenced by everyone. The `Err` prefix is convention. Callers use `errors.Is(err, ErrUnsupported)` rather than `err == ErrUnsupported` so that wrapped errors still match.

Well-known examples across stdlib: `io.EOF`, `sql.ErrNoRows`, `context.Canceled`, `context.DeadlineExceeded`, `fs.ErrExist`, `fs.ErrNotExist`.

### Unwrap is a convention, not an interface

Nowhere in the `errors` package is there a declaration like `type Unwrapper interface { Unwrap() error }`. Instead, `wrap.go` uses a runtime type assertion:

```go
if u, ok := err.(interface{ Unwrap() error }); ok { ... }
```

Any error type that implements `Unwrap() error` (or `Unwrap() []error` for joined errors) participates automatically in `errors.Is` and `errors.As` chain walking. This is the smallest possible protocol: no import, no registration, no interface embed — just add the method and it works.

### Takeaways

- `error` is a one-method interface; that's why so many things "become" errors for free.
- Pointer identity is what makes sentinels reliable — never compare error content when you meant to compare identity.
- The `Err`-prefixed package-level variable is Go's standard way to expose a comparable sentinel.
- `Unwrap()` is a duck-typed hook, not a declared interface; implementing it opts your error type into the whole `errors` toolkit.
- Prefer `errors.Is(err, sentinel)` over `err == sentinel` — the former also matches wrapped errors.

_Read (errors.go): 2026-08-15_

---

## errors — Is/As/Unwrap and chain walking

Where `errors.go` defines the raw material (an error, a sentinel), `wrap.go` and `join.go` define the *protocol* for searching through error chains. The whole toolkit rests on three optional methods any error type may implement: `Unwrap() error`, `Unwrap() []error`, and `Is(error) bool`.

### errors are trees, not chains

Once `Join` exists, an error is no longer a linear chain — it's a tree. `errors.Is` and `errors.As` traverse this tree **pre-order, depth-first**: check the node itself, then recurse into each child. A single-wrapped error is just the degenerate case of a tree with one branch per node.

```go
a := errors.New("a")
b := errors.New("b")
outer := errors.Join(fmt.Errorf("wraps: %w", a), b)
// tree:  outer ─┬── (wraps a) ── a
//               └── b
errors.Is(outer, a)  // true — depth-first found it
errors.Is(outer, b)  // true — sibling branch
```

### Unwrap the function vs Unwrap the method

The `Unwrap` **function** goes exactly one step:

```go
func Unwrap(err error) error {
    u, ok := err.(interface{ Unwrap() error })
    if !ok { return nil }
    return u.Unwrap()
}
```

The `Unwrap` **method** is what your type implements. The function calls the method. It handles only the single-wrap case (`Unwrap() error`); it does **not** descend into `Unwrap() []error` — that's what `Is`/`As` are for.

### Is — three checks per node, in order

```go
for {
    if targetComparable && err == target      { return true }   // 1. identity
    if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) { return true } // 2. custom Is
    switch x := err.(type) {                                                          // 3. descend
    case interface{ Unwrap() error }:   err = x.Unwrap(); ...
    case interface{ Unwrap() []error }: for _, e := range x.Unwrap() { recurse... }
    default: return false
    }
}
```

Two subtleties:
- **`Comparable()` guard**: `target` gets a runtime check because comparing an incomparable type (e.g. struct with a slice field) would panic on `==`. `reflectlite.TypeOf(target).Comparable()` short-circuits step 1 for those.
- **Custom `Is` method**: this is how types declare semantic equivalence to another error without wrapping it. `syscall.Errno.Is` is the canonical example — it maps OS-level error numbers to `fs.ErrExist`, `fs.ErrNotExist`, etc.

### As and AsType — extract a typed value from the tree

`As` fills a pointer; `AsType` (Go 1.20+) is the generic-native replacement:

```go
// old style
var perr *fs.PathError
if errors.As(err, &perr) { fmt.Println(perr.Path) }

// preferred style
if perr, ok := errors.AsType[*fs.PathError](err); ok { fmt.Println(perr.Path) }
```

`AsType` avoids the `**T` gymnastics, gives real generic type inference, and reads like `type assertion but chain-aware`. Package doc explicitly recommends preferring it over `As`.

### Join — the multi-error tree node

```go
func Join(errs ...error) error {
    // filter nils
    // if zero remaining, return nil
    // otherwise wrap in *joinError{errs: [...]}
}

type joinError struct { errs []error }
func (e *joinError) Unwrap() []error { return e.errs }
```

Two contract details worth remembering:
- **`Join(nil, nil, nil)` returns `nil`**, not an empty joined error. Zero non-nil inputs → nil output.
- **The `Unwrap() []error` variant** is what makes `Is`/`As` recurse into every child. Without a slice-returning `Unwrap`, joined errors would be opaque.

The `Error()` implementation uses `unsafe.String(&b[0], len(b))` instead of `string(b)` — a deliberate zero-allocation conversion, safe because the `[]byte` isn't used after that line.

### The rule of thumb

- **Sentinel comparison** → `errors.Is(err, ErrX)`
- **Typed extraction** → `errors.AsType[*T](err)`
- **Wrapping** → `fmt.Errorf("context: %w", err)` — never build a wrapper struct if you don't need custom fields
- **Multiple errors** → `errors.Join(errs...)` — nil inputs are silently dropped
- **Custom equivalence** → implement `Is(error) bool` on your type
- **Custom type extraction** → implement `As(any) bool` on your type

### Takeaways

- The chain-walking algorithm is a `for` loop with three checks: identity, custom `Is`, then Unwrap-descend.
- `errors.Unwrap` is one step; `errors.Is`/`As` are the whole tree. Never write your own recursion — use these.
- `%w` in `fmt.Errorf` is the everyday way to wrap; a custom type with `Unwrap()` is only needed when you have extra fields.
- `Join`'s `Unwrap() []error` turns a linear chain into a tree — think trees from day one, chains are a special case.
- Prefer `AsType[T]` over `As(&target)` in Go 1.20+ code — it's the modern, generic-aware API.

_Read (wrap.go + join.go): 2026-08-16_

---

## runtime/string.go — internal layout

A Go string is 16 bytes on a 64-bit machine: `str unsafe.Pointer` + `len int`. Nothing else. That header sits on the stack (or inside a struct); the actual bytes live elsewhere — read-only segment for literals, heap for dynamic strings, sometimes on the caller's stack via a compiler-managed temporary buffer. `stringStruct` in the runtime is the actual layout, and `stringStructDWARF` is the same shape with `*byte` instead of `unsafe.Pointer` so debuggers can render the contents.

### immutability is a contract, not an enforcement

There's no syntax for `s[0] = 'H'` — the compiler rejects it. But the runtime relies on immutability everywhere: sentinel error comparison, map keys, string interning, safe cross-goroutine sharing without a mutex. That trust chain is what makes `slicebytetostring` and `stringtoslicebyte` both copy: a `[]byte` is mutable, so a shared backing array between the two would let a caller silently rewrite a "string" someone else is holding. One line in the runtime confirms this — `// unlike slicerunetostring, no race because strings are immutable.` The escape hatch (`unsafe.String(&b[0], len(b))`) exists precisely because the safe path is otherwise mandatory.

### slicebytetostring — a three-tier allocation strategy

Every `string(b)` conversion routes through this function, and the tier depends on size and escape analysis:
- **`n == 0`** → return `""`. Common enough that it's the first branch — parsing between markers hits this constantly.
- **`n == 1`** → the `staticuint64s` trick. That's a 256-entry `[uint64]` array in RODATA (defined in `ints.s`), where entry `i` holds the value `i`. On little-endian, the byte at `&staticuint64s[i]` is `i` itself; on big-endian, the same byte lives at offset `+7`, hence the `if goarch.BigEndian { p = add(p, 7) }`. Zero allocation, forever. The same table is reused for interface conversion of small integers in `iface.go` — one 2KB table, two use cases.
- **`n <= 64` and non-escaping** → the caller passes a `*tmpBuf` (a `[64]byte` on their stack), and the runtime copies into that. Zero heap allocation; the buffer dies with the caller's frame.
- **otherwise** → `mallocgc(uintptr(n), nil, false)`. `typ == nil` tells the GC "no pointers inside, don't scan"; `needzero == false` skips zeroing because `memmove` overwrites everything immediately.

Then `memmove` copies the bytes and `unsafe.String((*byte)(p), n)` wraps the destination in a string header.

### stringtoslicebyte — the mirror image

`[]byte(s)` uses the same two-tier idea (stack `tmpBuf` or heap `rawbyteslice`) but always copies. There's no single-byte trick going this direction, because the result is mutable — you can't point every `[]byte{x}` at a shared read-only slot, they'd alias each other.

### concatstrings — two-pass to avoid quadratic copying

`a + b + c` becomes `concatstring3(&tmpBuf, a, b, c)` at compile time (or `concatstrings(...)` for higher arity). The algorithm is two passes:
1. Walk the strings once. Sum the lengths; count non-empty strings; catch overflow with the `l+n < l` trick (in unsigned arithmetic, if the sum wraps below the previous value, it overflowed).
2. Handle two special cases before allocating:
   - `count == 0` → all empty, return `""`.
   - `count == 1` and the surviving string's data isn't on the current goroutine's stack (`stringDataOnStack`) → return that string directly. `"" + s + ""` costs zero.
3. Otherwise, `rawstringtmp(buf, l)` — one allocation sized exactly to the total, on the caller's stack if it fits in the 64-byte `tmpBuf`, else on the heap.
4. Walk again, `copy(b, x)` each source into the buffer.

The alternative — building left-to-right (`(a+b)+c`) — would allocate a temporary for `a+b`, discard it, then allocate for the final result, doubling the copy of `a` and `b`. Two passes trade one extra `len(x)` walk for a single allocation. The pattern generalises: `strings.Builder` uses the same principle in a loop by growing its internal `[]byte` with the slice doubling rule instead of allocating per append. Anything in a loop that concatenates should use `Builder` (ideally with `Grow(size)` upfront) — measured benchmarks show 4–7× speedups and 3–18× fewer allocations for 18 short words.

### Takeaways

- A string is a 16-byte header. Everything else — literals in RODATA, stack temporaries, heap-allocated bodies — is arrangement of the underlying bytes.
- Immutability is enforced by the compiler refusing index-assignment, and preserved by the runtime always copying at the `string ↔ []byte` boundary.
- Three allocation tiers for `string(b)`: zero for empty/single-byte (`staticuint64s`), stack for small non-escaping (`tmpBuf`), heap otherwise (`mallocgc`). Escape analysis picks between the last two.
- `concatstrings` is a two-pass measure-then-copy algorithm with a "single non-empty string" fast path. It guarantees exactly one allocation per `+` expression.
- Inside a loop, always reach for `strings.Builder` — the runtime's `+` optimization is per-expression, not per-loop, so cumulative concatenation degrades to O(n²).
- `unsafe.String(&b[0], len(b))` is the escape hatch when the safe copy cost is unacceptable. Correct usage means proving the `[]byte` won't be modified for the string's lifetime; getting it wrong corrupts everything downstream.

_Read (string.go): 2026-08-18_

---

## sort.Interface — pre-generic polymorphism

Before Go 1.18 there were no generics. If you wanted a generic algorithm that worked over any collection type, you had two bad choices: use `interface{}` with runtime type assertions (slow, unsafe), or write code generators. The `sort` package pioneered a third way that shaped Go's whole API design philosophy: **decouple the algorithm from the data by asking the caller to describe the data through a tiny interface**.

### The interface — three methods against integer indices

```go
type Interface interface {
    Len() int
    Less(i, j int) bool
    Swap(i, j int)
}
```

The interface never touches the underlying element type. It only knows indices. The caller supplies:

- how big the collection is,
- whether index `i` should come before index `j`,
- how to swap two indices.

`sort.Sort` then runs an algorithm that only calls these three methods. The algorithm has no idea whether it is sorting integers, strings, database rows, or graph nodes. This is the same trick `heap.Interface`, `flag.Value`, and `image.Image` use — describe behaviour through minimal method sets, let the standard library do the heavy lifting.

### The strict weak ordering contract

`Less` is not just "return true if smaller". It must define a **strict weak ordering** — the mathematical guarantee that comparisons are consistent:

- Transitivity: if `Less(i, j)` and `Less(j, k)` are both true, then `Less(i, k)` must be true.
- Equality is expressed by *both* directions being false: if `Less(i, j)` and `Less(j, i)` are both false, `i` and `j` are considered equal.
- Never true in both directions — that is a contradiction, and the algorithm will produce garbage, hang, or panic.

The most common bug is writing `<=` instead of `<`. For equal elements `<=` returns true in both directions, and the sort silently corrupts.

The floating-point NaN case is the canonical example: `NaN < 5.0` is false and `5.0 < NaN` is also false, so the raw `<` operator makes NaN "equal to everything" without being equal to anything. `sort.Float64Slice.Less` fixes this by forcing NaN to sort first:

```go
func (x Float64Slice) Less(i, j int) bool {
    return x[i] < x[j] || (isNaN(x[i]) && !isNaN(x[j]))
}
```

### `Sort` — four lines that guarantee O(n log n)

```go
func Sort(data Interface) {
    n := data.Len()
    if n <= 1 { return }
    limit := bits.Len(uint(n))
    pdqsort(data, 0, n, limit)
}
```

Two subtleties worth noticing:

- `n := data.Len()` is called **once** and cached. The docstring explicitly promises this ("one call to data.Len"). It is both a performance choice (avoiding repeated calls in the inner loop) and an API contract — the caller may implement `Len()` with an expensive computation and rely on it running exactly once.
- `limit := bits.Len(uint(n))` is `⌈log₂(n)⌉`. It is the maximum number of bad (unbalanced) pivots the pdqsort recursion is allowed to make before falling back to heapsort. This is what turns quicksort's O(n²) worst case into a guaranteed O(n log n).

### pdqsort, not introsort

The docs and older articles frequently say "introsort" (quicksort + heapsort fallback). Since Go 1.19 the algorithm is **pdqsort** (pattern-defeating quicksort). The structure in `zsortinterface.go` is a three-way hybrid:

```go
if length <= maxInsertion {       // maxInsertion = 12
    insertionSort(data, a, b)      // small ranges → insertion sort
    return
}
if limit == 0 {
    heapSort(data, a, b)           // too many bad pivots → heapsort
    return
}
// otherwise: quicksort partition with smart pivot selection
```

Insertion sort wins on tiny ranges because its constant factor is tiny. Quicksort wins on average-case data. Heapsort is the worst-case safety net. The `limit` counter is what triggers the fallback — every time pdqsort picks a badly unbalanced pivot, `limit` drops. Adversarial or already-pathological inputs run out of budget after `log₂(n)` bad picks and get switched to heapsort.

### `Reverse` — embedding as an override mechanism

```go
type reverse struct { Interface }
func (r reverse) Less(i, j int) bool { return r.Interface.Less(j, i) }
func Reverse(data Interface) Interface { return &reverse{data} }
```

Four lines produce a reversed sort without a new algorithm. The trick is Go's struct embedding: `reverse` inherits `Len` and `Swap` from the embedded `Interface`, and only overrides `Less`, swapping the argument order. This is the canonical "decorator" pattern in Go — wrap the value, override the one method you care about, get the rest for free. Same shape as `bufio.Reader` wrapping `io.Reader`, or middleware wrapping `http.Handler`.

### `sort.Slice` — closure ergonomics, reflect cost

Writing three methods just to sort a `[]Person` is tedious. `sort.Slice` lets you supply only the `Less` closure:

```go
sort.Slice(people, func(i, j int) bool {
    return people[i].Age < people[j].Age
})
```

The implementation (in `slice.go`) uses `reflectlite`:

```go
func Slice(x any, less func(i, j int) bool) {
    rv := reflectlite.ValueOf(x)
    swap := reflectlite.Swapper(x)
    length := rv.Len()
    limit := bits.Len(uint(length))
    pdqsort_func(lessSwap{less, swap}, 0, length, limit)
}
```

`reflectlite.Swapper(x)` builds a swap function at runtime by inspecting the slice element type. This works for any slice, but every swap pays a reflect penalty. Benchmarks typically show `sort.Slice` running 2–3× slower than a hand-rolled `sort.Interface` implementation. `pdqsort_func` (vs `pdqsort`) is a code-generated variant that takes function values instead of an interface — see the `//go:generate` directive at the top of `sort.go`.

**When to choose**:
- `sort.Sort` with a custom `Interface` — hot paths, large data, when the sort is called many times.
- `sort.Slice` — one-off code, ergonomics matter more than a small perf delta.
- `slices.SortFunc` (Go 1.21+) — the modern answer. Generic, no reflect, no allocation, closure-based ergonomics. `sort.Ints`, `sort.Float64s`, and `sort.Strings` are now thin wrappers over `slices.Sort`.

### `Stable` — preserving the order of equal elements

```go
func Stable(data Interface) { stable(data, data.Len()) }
```

`Stable` guarantees that elements considered equal by `Less` keep their original relative order. Regular `Sort` provides no such guarantee. The cost is O(n · log² n) swaps versus Sort's O(n · log n) — the algorithm is a symmetric merge with block swaps rather than pdqsort.

This matters because it enables **multi-key sorting by composition**. To sort by primary key A, then secondary key B, run stable sorts in the reverse order of importance:

```go
// Step 1: sort by the secondary key first.
sort.SliceStable(books, func(i, j int) bool {
    return books[i].Year > books[j].Year   // newest first
})

// Step 2: sort by the primary key. Equal-author books retain
// the year ordering from step 1 because the sort is stable.
sort.SliceStable(books, func(i, j int) bool {
    return books[i].Author < books[j].Author
})
```

If you used non-stable `Slice` in step 2, the year ordering from step 1 would be scrambled inside each author group. The pattern generalises to N keys — sort N times, from least important to most important, always stable.

### Takeaways

- `sort.Interface` decouples algorithms from data by exposing only Len/Less/Swap over indices — no knowledge of the element type. This is Go's answer to generics before generics existed, and remains the clearest example of interface design in the stdlib.
- `Less` must define a strict weak ordering. The failure mode is silent: using `<=` where you meant `<` produces contradictory comparisons and corrupt sorts. NaN is the canonical case where `<` alone is not enough.
- `Sort` promises exactly one `Len()` call — both a performance choice and an API contract callers can rely on.
- The algorithm is pdqsort (not introsort since Go 1.19): a three-way hybrid of insertion sort for tiny ranges, quicksort for the general case, heapsort as the O(n log n) fallback when `limit` (initialised to `log₂(n)`) runs out.
- `Reverse` demonstrates the decorator pattern via struct embedding — override one method, inherit the rest. Idiomatic Go, worth internalising.
- `sort.Slice` trades performance for ergonomics using reflect; `sort.Sort` with a custom Interface is 2–3× faster. Modern code should reach for `slices.SortFunc` instead — same ergonomics, no reflect.
- `Stable` costs O(n · log² n) swaps but enables multi-key sorting by composition. Rule: sort N times, from least important key to most important, always stable.

_Read (sort.go + slice.go + zsortinterface.go): 2026-08-20_

---

## internal/runtime/maps — Swiss Tables architecture (H1/H2, control word, Group)

Go 1.24 (February 2025) rewrote the entire map implementation. The old `runtime/map.go` — a chained hashmap where every bucket held eight `(key, value)` pairs and collisions spilled into linked overflow buckets — is now a thin shim. The real code lives in `internal/runtime/maps/`, and the design is a Go port of Google's Swiss Tables (open in Abseil since 2017). The migration wasn't cosmetic. Three properties of modern CPUs made chaining a bad fit: cache lines are 64 bytes and pointer-chasing wastes them, SIMD can compare 16 bytes in a single instruction and a per-bucket loop can't, and grow-time rehashing of a million-entry map produced latency spikes that were hard to bound. Swiss Tables address all three with one design: contiguous 8-slot groups, a compact metadata word compared in parallel, and per-table splitting instead of whole-map rehash.

### The hierarchy — Map → directory → Table → Group → Slot

A `Map` no longer owns a single flat array of buckets. It owns a **directory** — an array of pointers to `Table` structs. Each `Table` owns a small array of `Group`s. Each `Group` owns exactly 8 slots plus an 8-byte control word describing those slots. The five-level shape is deliberate: the outer layers (`Map`, directory, `Table`) exist so growth can happen incrementally — one `Table` at a time — instead of rehashing the whole map at once. The inner layers (`Group`, slot) are cache-line-friendly and designed for parallel matching.

The number eight is not arbitrary. AMD64 SIMD instructions like `pmovmskb` operate on 128-bit registers as sixteen 8-bit lanes; the 8-byte control word fits half a register and produces an 8-bit match bitmask in one instruction. The choice cascades: 8 slots per group → 8-byte control word → one SIMD op per lookup step. If you understand nothing else about Swiss Tables, understand that the architecture is shaped around a single hardware primitive.

### H1/H2 split — 57 bits for probing, 7 bits for the control word

Every key gets hashed to a 64-bit value, then split asymmetrically:

- **H1** — the upper 57 bits, used to pick the starting `Group` (`H1 mod numGroups`).
- **H2** — the lower 7 bits, used as a fingerprint stored inside the control word.

The split is asymmetric because the control word has room for only 7 bits per slot — the top bit is reserved for state (empty vs occupied vs tombstone). H2 alone is not enough to identify a key: with only 7 bits, roughly 1 in 128 unrelated keys will match by accident. That's why matching H2 is only a filter — every hit must be confirmed with a full key comparison. What H2 buys is speed: reject 127 out of 128 non-matches in one SIMD instruction before ever touching the actual keys.

### The control word — 8 bytes describing 8 slots

The control word is the beating heart of Swiss Tables. One byte per slot, encoding both state and fingerprint:

```
0x80              → empty
0xFE              → tombstone (deleted, but probe chain must continue)
0hhhhhhh (H2)     → occupied, low 7 bits carry the H2 fingerprint
```

The high bit is the state marker. Empty slots use `0x80` (top bit set, rest zero) so a byte-wide `& 0x80` reveals emptiness. Tombstones use `0xFE` — top bit set, rest one — distinguishing them from empty when a delete needs to preserve the probe chain. Occupied slots always have the top bit clear, so `H2 & 0x7F` never collides with `0x80` or `0xFE`.

Because all 8 bytes live in a single machine word, the CPU loads all slot metadata in one memory access. A lookup that finds nothing in the group still pays only one cache miss for the metadata — the actual keys are only touched when H2 says there might be a match.

### Takeaways

- Swiss Tables replaced `hmap` in Go 1.24 because chaining wasted cache lines, couldn't use SIMD, and produced unpredictable grow latency. The redesign fixes all three with one shape: contiguous 8-slot groups + compact control word + per-table splitting.
- The hierarchy Map → directory → Table → Group → Slot exists for two reasons: outer layers (Table + directory) enable incremental grow via extendible hashing; inner layers (Group + control word) enable SIMD parallel matching.
- H1/H2 splits the 64-bit hash asymmetrically: 57 bits pick the Group, 7 bits become an in-control-word fingerprint. H2 filters candidates in one instruction but never confirms — every H2 match still requires a real key comparison because 1 in 128 unrelated keys collide.
- The control word encodes state and fingerprint in one byte per slot: `0x80` empty, `0xFE` tombstone, `0hhhhhhh` occupied. Living in one machine word is what makes the whole design possible.

_Read (internal/runtime/maps/{map,table,group}.go): 2026-08-26_

---

## Swiss Table lookup: probe sequence + parallel match

A Swiss Table lookup is three steps: find the starting Group from H1, filter candidate slots with `matchH2` on the control word, then verify each candidate with a real key comparison. If the Group has no match but also no empty slot, probe to the next Group using a quadratic sequence. The design is a careful balance — H2 filters cheaply, key compare confirms correctly, and the probe rule guarantees every Group is visited exactly once before giving up.

### Step 1 — H1 picks the starting Group

```go
startGroup := h1 % numGroups
```

Because `numGroups` is always a power of two, the modulo compiles to a single bitwise `AND (numGroups - 1)`. This is a hard constraint: everything downstream — the quadratic probe formula, the extendible hashing directory sizing, the small-map optimization — assumes power-of-two sizing.

### Step 2 — `matchH2`: parallel compare with bit tricks

Once you're in a Group, you have 8 slots and an 8-byte control word describing them. The naive approach — loop 8 times, compare each byte to H2 — is exactly what chaining did and exactly what Swiss Tables avoid. Instead, `matchH2` runs three bitwise operations over the whole control word simultaneously:

```
// pseudocode; real code uses SIMD on amd64
func matchH2(cw uint64, h2 byte) uint8 {
    // 1. XOR each byte with a repeated H2. Matching bytes become 0x00.
    x := cw ^ (uint64(h2) * 0x0101010101010101)
    // 2. Bit-tricks to produce 0x80 for each byte that was 0x00.
    y := (x - 0x0101010101010101) & ^x & 0x8080808080808080
    // 3. Compress those 0x80 markers into an 8-bit mask.
    return compressToBitmap(y)
}
```

On AMD64 the whole thing is one `pmovmskb` after a SIMD compare — 16 lanes in ~1 nanosecond. The output is a bitmask where bit `i` is set if slot `i` has a fingerprint matching H2. No loops, no branches, no per-slot dispatch.

The false-positive rate is `2^-7 = 1/128`. So `matchH2` might return spurious matches, but at that rate the wasted work is negligible compared to the branch-free win.

### Step 3 — Verify with key compare, then probe

For each bit set in the match bitmask, load the actual key from that slot and compare to the lookup key. First real match → return the value. If all H2 candidates fail (false positives), continue.

Then check the control word for any empty slot (`0x80`). If found, the key **cannot exist** — the probe chain would have stopped here on insertion, so if it isn't in this Group and there's still empty room, it was never inserted anywhere down the chain. Return "not found".

If the Group is full (no empty, no match), the key might be further down. Probe to the next Group using the quadratic sequence and repeat.

### Quadratic probing — `p(i) = start + (i² + i)/2 mod n`

Linear probing (`p(i) = start + i`) has a classic pathology called **primary clustering**: once collisions form a run of occupied slots, every subsequent insert into that region extends the run, and lookup times degrade quadratically with load. Quadratic probing breaks the pattern by taking increasingly-spaced jumps: `+1, +3, +6, +10, +15, ...`.

The specific formula `(i² + i)/2 mod n` isn't cosmetic. When `n` is a power of two, this sequence is a **bijection on Z/nZ** — every value from 0 to n-1 appears exactly once as `i` goes from 0 to n-1. That guarantees the probe will visit every Group before giving up, so a key present anywhere in the table will be found, and a full table won't cause an infinite loop.

### Takeaways

- Lookup is three steps: H1 → starting Group, `matchH2` → candidate slots, key compare → confirm. The design is built around H2 filtering being cheap and wrong-sometimes; key compare being expensive and always-right.
- `matchH2` uses three bitwise ops (or one SIMD instruction on amd64) to compare H2 against all 8 slots at once and return an 8-bit match mask. No loop, no branch — this is the operation that makes the whole architecture worth it.
- An empty slot (`0x80`) in the current Group ends the search. This is why delete cannot simply zero a slot — that would end searches early and break the probe chain. Tombstones (`0xFE`) exist for exactly this reason.
- Quadratic probing with `p(i) = start + (i² + i)/2 mod n` avoids the primary clustering of linear probing, and because `n` is always a power of two, the sequence is a bijection — every Group visited exactly once before terminating.

_Read (internal/runtime/maps/table.go + group.go): 2026-08-26_

---

## hashmap with open addressing — my implementation vs stdlib Swiss Tables

Writing [04-hashmap/hashmap.go](04-hashmap/hashmap.go) after reading `internal/runtime/maps` made the design differences between a textbook open-addressing map and a production one concrete. The core ideas — open addressing, probing, tombstones, load-factor-triggered resize — carry over. What's different is that stdlib takes every one of them and does it in a way that's specifically tuned for how CPU cache lines, SIMD, and predictable-latency growth work.

### What my implementation shares with stdlib

Both use open addressing rather than chaining — no linked overflow lists, keys live directly in the primary array. Both use tombstones for delete: a plain removal would break the probe chain by turning an occupied slot into empty, causing subsequent lookups to stop early and miss keys that legitimately hashed further down. My `entry.state` field uses `0/1/2` for empty/occupied/tombstone; stdlib uses `0x80/0xhh/0xFE` in a control byte — same three states, different encoding for different reasons (mine is readable, theirs is SIMD-friendly).

Both trigger resize based on load. Mine uses `(size + tombstones) * 4 >= len(slots) * 3`, effectively a 75% threshold that includes tombstones — this matters because a table dense with tombstones probes as badly as one dense with real entries, even if `size` looks low. Stdlib uses a similar rule per-Table with `maxTableCapacity = 1024`.

### What stdlib does that mine doesn't

- **Groups of 8 with a control word.** My design is one slot per index. Stdlib bundles 8 slots into a Group with an 8-byte metadata word packed alongside. This is the entire point: hardware can load and process the metadata for all 8 slots in one instruction.
- **`matchH2` parallel filter.** My lookup walks slots one at a time, comparing each `entry.key` to the search key. Stdlib does an SIMD compare of H2 against all 8 control bytes at once, then does a full key compare only for the ~1/128 that matched. The typical lookup touches one Group's metadata (cheap) and zero or one full keys (also cheap).
- **Quadratic probing, not linear.** I picked linear probing because it's two lines and easy to reason about (`index = (index + 1) % len(h.slots)`). It also clusters badly at high load: any run of collisions absorbs subsequent inserts and grows. Stdlib probes with `p(i) = start + (i²+i)/2 mod n`, which breaks clusters without giving up the "visits every group exactly once" guarantee.
- **Extendible hashing on grow.** My `resize()` copies every occupied entry from the old array into a new array of double size, in one shot. For a map of 1 million entries, that's 1 million rehashes and copies at once — a latency spike. Stdlib bounds each Table at 1024 entries; when one fills up, it splits *only that Table* into two, updates the directory, and moves at most ~1024 entries. The bound holds regardless of overall map size, so grow latency is independent of scale.
- **Small-map optimization.** For maps with ≤ 8 entries, stdlib skips the Table layer entirely — `dirPtr` points directly at a single Group. No probe sequence, no tombstones, no directory. My design pays the full open-addressing cost even for a two-entry map.

### Trade-offs I accepted

Linear probing gave me clarity — I can trace a lookup by hand. Quadratic probing needs more thought about whether the formula covers every slot (only if `n` is a power of two). Full rehash on grow is O(n) but has zero conceptual overhead; extendible hashing needs a directory abstraction, local vs global depth tracking, and Table-level splitting logic. And running everything as individual slots without a control word means my `Get` is a comparison loop instead of a bitwise trick — slower, but the loop fits in my head without diagrams.

For a hobby-scale hashmap this is fine. For a language runtime where `map[string]int` is one of the hot paths across every program ever written in the language, none of it is fine.

### What I'd change if this were production

The two most impactful changes wouldn't be SIMD — that's platform-specific and Go abstracts it. They'd be:

1. **Quadratic probing.** Two-line change to the probe formula, no other design impact, removes worst-case clustering behavior. Costs nothing.
2. **Cap resize granularity.** Split into a fixed-size table structure so a single grow moves at most N entries. This alone eliminates the pathological "the map hit its limit and now the request-handling latency spiked by 200ms" incident that killed old chaining-based `hmap` in some workloads.

SIMD parallel matching is a further order-of-magnitude win, but only if the rest of the design (Group layout, control word, fingerprint) is already in place. It's not something you retrofit — the data structure has to be shaped for it from the start.

### Takeaways

- Open addressing, tombstones, and load-factor resize are the shared foundation between any textbook implementation and stdlib Swiss Tables. The interesting engineering is all in *how* those primitives are laid out for hardware.
- The single biggest win of Swiss Tables over a naive open-addressing map is `matchH2` — one SIMD instruction that filters 128 candidates down to 1 without any per-slot branching. Everything else in the design exists to make that instruction possible.
- Extendible hashing is what makes stdlib's grow cost independent of map size. My full-rehash approach is O(n) per grow; stdlib's is O(1024) per grow, regardless of whether the map has 10k or 10M entries.
- If I were retrofitting a production map without going full SIMD, the two changes with the best cost/benefit ratio would be quadratic probing (removes clustering, two-line change) and bounded-size Tables with per-Table splits (removes latency spikes on grow).

_Read (internal/runtime/maps/{map,table,group}.go): 2026-08-26_ · _Wrote (04-hashmap/hashmap.go): 2026-08-26_

---

## Go Memory Model — happens-before and sync semantics

Reference: [go.dev/ref/mem](https://go.dev/ref/mem)

### The problem: why do we need a memory model?

Without synchronization, three actors can reorder your writes so another goroutine sees them out of order:

| Actor    | What it does                                                                       |
| -------- | ---------------------------------------------------------------------------------- |
| Compiler | Reorders independent statements when it thinks the result is faster                |
| CPU      | Uses store buffers — writes may reach main memory later than issued                |
| Cache    | Each core has its own cache — a write on core A is not instantly visible on core B |

Classic broken example:

```go
var a = 0
var done = false

// goroutine 1
a = 42
done = true

// goroutine 2
if done {
    fmt.Println(a) // may print 0, not 42
}
```

There is no guarantee that goroutine 2 sees `a=42` even when it sees `done=true`. The writes can be reordered, delayed in a store buffer, or invisible in another core's cache.

### happens-before

The memory model defines a relation called **happens-before** between events (reads, writes, sync ops) across all goroutines.

> If event A **happens-before** event B, then any goroutine that observes the effect of B is guaranteed to also observe the effect of A.

Key properties:

- It is a **partial order**, not a total order. Some events have no ordering relation → those are candidates for a data race.
- A read of variable `v` is allowed to observe a write to `v` only if that write is not ordered _after_ the read by happens-before, and no later write is ordered before the read.
- If two goroutines access the same variable without a happens-before between the accesses, and at least one is a write, that is a **data race**. Behavior is undefined.

### What creates happens-before in Go

Plain assignment does not. You need one of these sync primitives:

1. **`sync.Mutex` / `sync.RWMutex`**
   - `Unlock()` on mutex `m` happens-before the next `Lock()` on the same `m` returns.
   - For `RWMutex`: `Unlock()` happens-before any subsequent `RLock()` return; `RUnlock()` happens-before any subsequent `Lock()` return.

2. **Channels**
   - A send on channel `c` happens-before the corresponding receive from `c` completes.
   - The `k`th receive on a channel with capacity `C` happens-before the `(k+C)`th send completes (backpressure ordering).
   - `close(c)` happens-before a receive that returns because the channel is closed.

3. **`sync/atomic`**
   - An atomic operation A happens-before atomic operation B on the same address if A is sequenced before B in program order and A is a "release" and B is an "acquire" (all Go atomics are sequentially consistent by default).

4. **`sync.Once`**
   - The completion of the function `f` inside `once.Do(f)` happens-before the return of any other `once.Do(f)` call.

5. **Goroutine start**
   - The `go f()` statement happens-before the start of `f`'s execution.

6. **Goroutine exit**
   - The exit of a goroutine happens-before **no** event. If you need to wait for it, use `sync.WaitGroup` or a channel — the exit itself provides no synchronization.

### Why a mutex is not "just a lock"

A mutex does two things, and the second is the one people forget:

1. **Mutual exclusion** — only one goroutine holds it at a time.
2. **Memory barrier** — `Unlock` tells the CPU to flush pending writes to main memory; the next `Lock` tells the CPU to invalidate its cache and reload from memory.

Without the barrier, each core would happily live in its own cache and never see the others' writes. Locks give you _both_ exclusion and visibility. This is why replacing a mutex with a plain `bool` flag is broken even if only one goroutine writes it.

### Practical rules

- If two goroutines share a variable and at least one writes, protect it with a mutex, atomic, or channel. No exceptions.
- Never assume word-sized writes are atomic. On most architectures they are, but the memory model does not guarantee it — a torn read is legal.
- The race detector (`go test -race`, `go run -race`) finds violations at runtime. It has false negatives (only catches races that actually happen during the run), never false positives. Run it in CI.
- Double-checked locking is broken in Go without `sync/atomic` or `sync.Once`. The "fast path" read outside the lock has no happens-before with the write inside the lock.

_Read (go.dev/ref/mem — Introduction, Advice, Informal Overview, Synchronization): 2026-08-30_

## runtime/proc.go — G-M-P scheduler basics

Why G-M-P: Replaces heavy, expensive OS threads with lightweight, cheap goroutines.

Roles in one word: G = Task/Code, M = Worker (OS Thread), P = Resource/Execution Context.

runnext slot: A 1-element fast track; runs the woken goroutine immediately to exploit CPU cache locality.

Local Runqueue: Lockless 256-element array; owner pops from head, pushes to tail; stealers CAS the head.

findrunnable ordering: Local → Global → Netpoller → Steal → Park. (The `schedtick % 61 == 0` check prevents Global queue starvation).

_Read (src/runtime/proc.go — schedule, findrunnable, runqput, runqget) + Kavya Joshi "The Scheduler Saga": 2026-09-03_

## runtime/chan.go — hchan struct anatomy

buf (unsafe.Pointer): Enables generic storage (pre-generics) and contiguous allocation with hchan.

Circular Buffer: dataqsiz (capacity), qcount (item count), sendx (write index), recvx (read index).

waitq + sudog: Linked list of blocked Gs. sudog wraps a G so one G can wait on multiple channels (select).

runtime.mutex vs sync.Mutex: Avoids bootstrap circularity (sync.Mutex depends on gopark).

Buffered vs Unbuffered: Unbuffered has dataqsiz = 0 and bypasses buffer via sendDirect (stack-to-stack copy). Buffered uses dataqsiz > 0.

_Read (src/runtime/chan.go — hchan struct + file-top comments): 2026-09-03_

## runtime/proc.go — G-M-P scheduler basics

**The three players**: G (goroutine — the task + its stack), M (OS thread — the actual CPU worker), P (processor/context — holds the local run queue and the resources an M needs to run Go code, count = GOMAXPROCS). An M can't run Go code without holding a P.

**Local run queue (`pp.runq`)**: a fixed-size (256) circular array on each P, no lock — just `atomic` head/tail. Why lock-free is safe here: only the *owner* P ever pushes new work in; other P's are only allowed to read from the other end (steal). One-writer-many-occasional-readers is a well-known pattern (a work-stealing deque) that doesn't need a mutex.

**`runnext` — the fast-track slot**: a single extra slot outside the ring, set with a CAS loop (`runqput`, line ~7060). A newly-spawned goroutine usually lands here instead of the back of the queue — cache is still warm from the goroutine that just created it, so running it next is cheap. If `runnext` was already occupied, the old occupant gets kicked to the back of the regular queue (not lost).

**Overflow → global queue (`runqputslow`)**: when the local ring is full, instead of moving items one at a time (which would mean grabbing the global lock repeatedly), the owner P grabs *half* its local queue as one batch and dumps it on the global run queue in a single locked operation. Fewer lock acquisitions = less contention.

**`findRunnable()` search order** (verified against Go 1.26 source, `proc.go`):
1. **Local queue** (`runqget`) — cheapest, no lock.
2. **Global queue** — but only unconditionally checked every 61 scheduler ticks (`pp.schedtick % 61 == 0`), even if the local queue isn't empty. Why: without this, two goroutines that keep re-spawning each other could hog a P's local queue forever and starve everything sitting on the global queue. This is the scheduler's fairness escape hatch.
2b. Otherwise, global queue is tried right after local is empty.
3. **Netpoll** — check if any network I/O completed and unblocked a goroutine, before resorting to the expensive option.
4. **Work stealing** — an idle/spinning M starts grabbing work from *other* P's local queues (from the tail, i.e., the "cold" end, to avoid fighting the owner over `runnext`).
5. **Park** — nothing found anywhere; the M gives up its P and goes to sleep until woken.

**Lock-free vs locked — the general lesson**: whether a data structure needs a mutex isn't about "is it shared" (the local run queue *is* shared, via stealing) — it's about the *shape* of the sharing. Single-writer/occasional-reader → atomics are enough. Genuinely multi-writer/multi-reader with several fields that must change together (like a channel's buffer + wait queues + closed flag) → a real lock is simpler and safer than trying to be clever with atomics.

---

## runtime/chan.go — hchan struct anatomy + chansend walkthrough

**Struct fields** (Go 1.26; two fields — `timer`, `bubble` — are newer than what older references describe, used for timers and the `synctest` testing package; core logic is unchanged):
- `qcount`, `dataqsiz` — how many elements are queued right now, and the buffer's total capacity.
- `buf` — pointer to the circular buffer array (only allocated if the channel has a buffer size > 0).
- `sendx`, `recvx` — the write/read cursors into that circular buffer (each wraps back to 0 after reaching `dataqsiz`).
- `recvq`, `sendq` — FIFO linked lists (`waitq`) of `sudog`, i.e., goroutines currently parked waiting to receive / to send.
- `closed` — 0 or 1.
- `lock mutex` — a low-level `runtime.mutex`, **not** `sync.Mutex`. Why: `sync.Mutex` itself is implemented using `gopark`/`goready`, which are runtime primitives — channels sit *below* that layer, so using `sync.Mutex` here would be circular (it would depend on machinery that itself might depend on channels).

**Invariant from the file's own header comment**: "`qcount > 0` implies `recvq` is empty" (and vice versa for send). In plain words: you can never simultaneously have data sitting in the buffer *and* a goroutine parked waiting to receive — if a receiver is already waiting, the next value skips the buffer and goes straight to them.

**`chansend` — the four paths** (real source, `chan.go` ~line 176-310):
1. **`c == nil`** → `gopark` forever. Not a panic — a permanent block. (Useful trick: in a `select`, nil-ing out a channel variable is how you disable that case.)
2. **A receiver is already parked** (`recvq` non-empty) → direct handoff: the value is copied straight from the sender's stack to the receiver's stack (`send()` → one `memmove`), completely bypassing the buffer. Two reasons for this: (a) performance — one copy instead of two (sender→buffer, buffer→receiver); (b) it's required by the invariant above — buffer and waiting-receivers can't coexist.
3. **Buffer has room** (`qcount < dataqsiz`) → value goes into `buf[sendx]`, `sendx` advances (wrapping), `qcount++`. No blocking.
4. **No receiver, no room** → if non-blocking (`select` with `default`), return false immediately. Otherwise: build a `sudog`, enqueue it on `sendq`, call `gopark` — the goroutine leaves the scheduler entirely until some receiver wakes it up.

**Why direct handoff (path 2) matters for interviews**: it's the concrete answer to "how does an unbuffered channel actually work" — there's no hidden buffer of size 1; the handshake *is* the synchronization point (this is also why unbuffered channel send/receive is often described as "happens concurrently with" rather than "happens before" the matching operation, from the Memory Model note).

**`chanrecv` — the mirror image** (`chan.go` ~line 524): same four cases as `chansend`, reversed — nil channel blocks forever; a waiting sender gets served via `recv()`; data already in the buffer is taken directly; otherwise the receiver parks (`sudog` on `recvq`).

**`recv()` — the "full buffer + waiting sender" trick** (~line 702): when the buffer is completely full *and* a sender is already parked waiting, Go doesn't just hand the sender's value straight to the receiver — it does a **slot swap**: the oldest item (at `recvx`, the head of the circular buffer) goes to the receiver, and the parked sender's value is written into that exact same now-empty slot. Nothing shifts, order is preserved, and it's still just one extra copy, not a shift of the whole buffer.

**`sudog` — the "waiting ticket"**: a lightweight struct created only when a goroutine has to park on a channel (or similar primitive). Holds: `g` (which goroutine), `elem` (pointer to the value being sent/received), `c` (which channel), `success` (did it get real data, or did the channel just get closed). When the other side is ready, it looks up the `sudog` at the front of the wait queue and calls `goready(sg.g)` to wake exactly that goroutine — not a broadcast, a targeted wakeup.

## Bounded channel from scratch — mutex + slice

In Version 1 (append + buf[1:]), removing an element drops the front pointer, leaving orphaned memory at the beginning that triggers costly realocations and continuous memory copies during append.Version 2 (Ring Buffer with head/tail) fixes this by reusing fixed memory—incrementing pointer indices wrapping around via modulo (%), yielding true $\mathcal{O}(1)$ operations with zero memory allocations or slice reslicing overhead.

## RWMutex — readers don't block readers

`sync.RWMutex` has two locks: `RLock` for readers and `Lock` for writers.
Many readers can hold `RLock` at the same time; a writer needs everyone out.
Use it when reads are far more common than writes (config, caches).
Writer starvation problem: if readers keep coming, a writer could wait forever.
Go's fix: once a writer calls `Lock` and is waiting, new readers block until
that writer is done. So a pending writer stops the reader stream.
(To verify in `sync/rwmutex.go` on Tuesday: `readerCount` goes negative while
a writer is pending.)

## Buffered channel happens-before: k-th receive → (k+C)-th send

For a channel with capacity C, the k-th receive happens-before the (k+C)-th
send completes. Example, C = 1: send #1 fills the buffer; send #2 cannot
finish until receive #1 has happened. So a buffered channel of size C works
as a counting semaphore with C slots — at most C sends can be "in flight"
without a matching receive. With C = 0 (unbuffered) this collapses to
"receive happens-before send completes", i.e. a rendezvous.

## atomic vs non-atomic access to the same variable

Atomic operations only give guarantees if *every* access to that variable is
atomic. If one goroutine does `atomic.StoreInt64(&x, 1)` and another reads
`x` with a plain `x == 1`, that is still a data race — the race detector will
flag it. Mixing is never "half safe". Rule: pick one — all `atomic.*`, or
all under the same mutex.

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

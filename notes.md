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

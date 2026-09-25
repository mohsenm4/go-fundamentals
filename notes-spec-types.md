# Go Language Specification — Notes (Section 7: Types)

Reference: https://go.dev/ref/spec#Types

---

## 7.1 Types (intro)

A type = set of values + operations/methods on them.

```
Type     = TypeName [TypeArgs] | TypeLit | "(" Type ")"
TypeLit  = ArrayType | StructType | PointerType | FunctionType |
           InterfaceType | SliceType | MapType | ChannelType
```

- **Named types**: predeclared (`int`, `bool`, ...), defined (`type Foo ...`), type parameters.
- **Unnamed types**: type literals (`[]int`, `struct{...}`, `map[K]V`, ...).
- **Alias** (`type A = B`) is not a new type — just a second name.
- **Defined** (`type A B`) creates a new distinct type.

---

## 7.2 Boolean

- `bool`, values: `true`, `false`. Zero: `false`.

---

## 7.3 Numeric

**Sized ints:** `uint8..64`, `int8..64`. Aliases: `byte = uint8`, `rune = int32`.
**Arch-sized:** `int`, `uint` (32 or 64), `uintptr` (holds bits of a pointer).
**Floats:** `float32`, `float64` (IEEE 754). **Complex:** `complex64`, `complex128`.

- All numeric types are **distinct** — `int` ≠ `int32` even if same size. Explicit conversion required.
- Two's complement for ints. Zero value: `0` / `0+0i`.
- `uintptr` is used with `unsafe.Pointer` for low-level arithmetic. GC does **not** track it — holding only a `uintptr` won't keep memory alive.
- Numeric constants are arbitrary precision until assigned to a typed target.

---

## 7.4 String

- `string`, zero: `""`. **Immutable**.
- `len(s)` = number of **bytes**, not runes.
- `s[i]` → byte. `for _, r := range s` → rune.
- `&s[i]` **illegal** — would break immutability (strings may live in read-only memory or be shared).
- Mutate via `[]byte(s)`, modify, then `string(b)`.

---

## 7.5 Array

```
ArrayType = "[" ArrayLength "]" ElementType .
```

- **Length is part of the type**: `[5]int` ≠ `[10]int`.
- Length must be a `const int` expression, non-negative.
- **Value semantics**: assignment/pass copies the whole array.
- Zero: all elements zeroed.
- **Self-containment forbidden** via arrays/structs alone: `type T [10]T` illegal (size would be infinite). Pointer breaks the chain: `type T [10]*T` ok.

---

## 7.6 Slice

```
SliceType = "[" "]" ElementType .
```

- Header = (pointer, len, cap). Descriptor over an underlying array.
- Zero: `nil` (len=cap=0). `nil` slice and empty slice behave same for `len`, `range`, `append`.
- **Reference semantics**: assignment copies the header only — underlying array is shared. Mutation through one slice is visible in others sharing the array.
- Built via `make([]T, len, cap)` or a slice expression on an array/slice.

---

## 7.7 Struct

```
StructType    = "struct" "{" { FieldDecl ";" } "}" .
FieldDecl     = (IdentifierList Type | EmbeddedField) [ Tag ] .
EmbeddedField = [ "*" ] TypeName [ TypeArgs ] .
```

- Zero: each field zeroed.
- **Blank field** `_ T` for padding, no name conflict.
- **Embedded field**: `T` or `*T` with no name. Field name = the type's unqualified name. Fields and methods are **promoted** to the outer struct.
- Method-set promotion:
  - Embed `T`  → outer's method set gains `T`'s value-receiver methods. `*Outer` also gains `*T` methods.
  - Embed `*T` → both `Outer` and `*Outer` gain both `T` and `*T` methods.
- **Tags** (string_lit) are part of type identity; used by reflection only.
- Unexported names from different packages are always distinct.
- Empty struct `struct{}` occupies 0 bytes — useful as `map[K]struct{}` set.
- Self-containment: same rule as arrays. `type N struct{ next *N }` ok; `type N struct{ next N }` illegal.

---

## 7.8 Pointer

```
PointerType = "*" BaseType .
```

- Zero: `nil`. Deref via `*p`. Address via `&v`.
- Nil deref → runtime panic.
- Comparable (`==`, `!=`).
- `&x` requires `x` to be **addressable**:
  - ✅ variables, struct fields of addressable structs, array/slice elements
  - ❌ map values (rehash may move them), string bytes (immutable), function return values (temporary), arithmetic results
- Rule of thumb: if it can be on the LHS of `=`, it's usually addressable.

---

## 7.9 Function

```
FunctionType = "func" Signature .
Signature    = Parameters [ Result ] .
```

- Zero: `nil`.
- Param/result names: all named or all anonymous.
- Variadic: final param `...T`.
- A closure captures references to enclosed variables.

---

## 7.10 Map

```
MapType = "map" "[" KeyType "]" ElementType .
```

- Zero: `nil`. Read from nil map returns zero; **write to nil map panics**.
- Key type must be comparable (not func/map/slice). Interface key with non-comparable dynamic type → runtime panic.
- Built via `make(map[K]V)` or `make(map[K]V, sizeHint)`.
- `v, ok := m[k]` for presence check. `delete(m, k)`, `clear(m)`.

---

## 7.11 Interface (skim)

- Defines a **type set**.
- Zero: `nil`.
- `interface{}` = `any` = all non-interface types.
- Basic form: set of method signatures.

## 7.12 Channel (skim)

```
chan T          // bidirectional
chan<- T        // send-only
<-chan T        // receive-only
```

- `make(chan T)` unbuffered (sync). `make(chan T, n)` buffered.
- Zero: `nil` — never ready for send/recv.

## 7.13 Method sets (skim)

- Method set of `T`  = methods with receiver `T`.
- Method set of `*T` = methods with receiver `T` **or** `*T`.
- Interface satisfaction uses method sets. `*T` may satisfy an interface that `T` does not.

---

## 7.14 Properties of types and values

Introduction to the sub-rules that follow: representation, identity, assignability, representability.

---

## 7.15 Representation of values

**Self-contained** (variable holds the whole value):
- predeclared types, arrays, structs.

**Reference-holding** (variable holds a reference to shared data):
- pointer, function (with closure), slice, map, channel, interface.

Mutation through one reference is visible through others sharing the same underlying data.
Reference-holding types have `nil` as zero value.

---

## 7.16 Underlying types

Every type `T` has an underlying type:
- Predeclared or type literal → underlying is `T` itself.
- Otherwise → underlying of the type it refers to (walk the chain).
- For a type parameter → underlying of its constraint (an interface).

```go
type A1 = string  // alias
type A2 = A1
type B1 string    // defined
type B2 B1
type B3 []B1
type B4 B3
```

Underlying: `string` for `A1, A2, B1, B2`. `[]B1` for `B3, B4`.
Most type rules (assignability, conversion) work off underlying type.

---

## 7.17 Type identity

Are two types "the same"?

- **Named types are always distinct** from every other type. `type A []int` and `type B []int` are different.
- Two unnamed type literals are identical iff their structure matches component-wise.
- Alias (`type A = B`) does not create a new type: `A` and `B` are identical.

Per kind:
- **Array**: same element type + same length.
- **Slice**: same element type.
- **Struct**: same field sequence — names, types, tags, embeddedness all match. Unexported names from different packages never match.
- **Pointer**: same base type.
- **Function**: same param/result types, same variadic-ness. Names don't matter.
- **Interface**: same type set.
- **Map**: same key + element type.
- **Channel**: same element type + same direction.
- **Instantiated generic**: same defined type + identical type arguments.

Practical value: distinct named types (`UserID`, `ProductID`) give type safety — the compiler refuses to mix them without explicit conversion.

---

## 7.18 Assignability

Can `var t T = x` (where `x` has type `V`) compile without a conversion? Yes if any:

1. `V` and `T` are identical.
2. `V` and `T` have identical underlying types **and** at least one is not a named type (and neither is a type parameter).
3. Channel case: identical element types, `V` bidirectional, at least one is not named.
4. `T` is an interface (not a type parameter) and `x` implements `T`.
5. `x` is `nil` and `T` is pointer/func/slice/map/chan/interface.
6. `x` is an untyped constant representable by `T`.

Mental model for rule 2: two named types like `MyInt` and `int` won't cross-assign because both carry meaning. But a named type and an unnamed literal (`type S []int` vs `[]int`) will — the literal carries no separate meaning.

```go
type MyInt int
var x MyInt = 5
var y int = x       // ❌ both named
var y int = int(x)  // ✅ explicit

type S []int
var s S
var t []int = s     // ✅ S named, []int unnamed, same underlying
```

---

## 7.19 Representability

Only for **constants**. Is untyped constant `x` representable by type `T`?

- `x` is in `T`'s value set, or
- `T` is float and `x` rounds (IEEE 754 round-to-even) to `T`'s precision without overflow, or
- `T` is complex and both `real(x)` and `imag(x)` are representable by `T`'s component type.

```go
var b byte = 5      // ✅
var b byte = 256    // ❌ out of range
var b byte = -1     // ❌ signed
var i int = 42.0    // ✅ integral
var i int = 42.5    // ❌ not integer
var f float32 = 1e100  // ❌ overflow to Inf
```

Untyped constants stay flexible until assigned; the representability check happens at that moment. This is why the same literal (`65`) can be assigned to `int`, `byte`, `rune`, `float64` without any conversion.

---

## Cross-cutting summary

Three levels of type comparison:

| Question | Concept | Example that fails |
|---|---|---|
| Are these two types "the same"? | **Identity** | `type A []int`, `type B []int` — no |
| Can I assign `x` to a variable of `T`? | **Assignability** | `MyInt` → `int` without cast — no |
| Does this untyped constant fit into `T`? | **Representability** | `300` → `byte` — no |

Three foundations to keep in mind:

- **Underlying type** — the substrate most rules operate on.
- **Named vs unnamed** — drives identity and assignability.
- **Value vs reference semantics** — decides whether mutation is shared and whether `nil` is the zero.

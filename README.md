# doppel

> Your data's doppelgänger — deep copies without side effects.

`doppel` is a production-grade Go library for explicit, reflection-free deep cloning of arbitrary data structures. It is
built around a **performance-first, explicit-over-magic** philosophy: manual cloning is always the default, and every
copying decision is visible in code.

---

## Table of Contents

- [Why doppel?](#why-doppel)
- [Design Philosophy](#design-philosophy)
- [Installation](#installation)
- [Package Layout](#package-layout)
- [Core Concepts](#core-concepts)
    - [Cloner\[T\]](#clonert)
    - [SelfClonable\[T\]](#selfclonablet)
    - [Identity helpers](#identity-helpers)
- [API Reference](#api-reference)
    - [doppel.Clone](#doppelclone)
    - [doppel.MustClone](#doppelmustclone)
    - [doppel.CloneWith](#doppelclonewith)
    - [doppel.MustCloneWith](#doppelmustclonewith)
    - [manual.CloneSlice / CloneSliceOf](#manualcloneslice--clonesliceof)
    - [manual.CloneMap / CloneMapOf](#manualclonemap--clonemapof)
    - [manual.ClonePointer / ClonePointerOf](#manualclonepointer--clonepointerof)
- [Usage Guide](#usage-guide)
    - [Step 1 — Simple struct (primitives only)](#step-1--simple-struct-primitives-only)
    - [Step 2 — Struct with a pointer field](#step-2--struct-with-a-pointer-field)
    - [Step 3 — Struct with slices and maps](#step-3--struct-with-slices-and-maps)
    - [Step 4 — Full aggregate with nested structs](#step-4--full-aggregate-with-nested-structs)
    - [Step 5 — External Cloner (no SelfClonable)](#step-5--external-cloner-no-selfclonable)
    - [Step 6 — Conditional / filtered cloning](#step-6--conditional--filtered-cloning)
- [Error Handling](#error-handling)
- [Nil Safety Contract](#nil-safety-contract)
- [Running Tests](#running-tests)
- [Benchmark Results](#benchmark-results)
- [Roadmap](#roadmap)

---

## Why doppel?

Most deep-copy libraries in Go use `reflect` as their primary engine. Reflection works for any type automatically, but
it comes with real costs:

- **Performance** — reflection bypasses the compiler's type knowledge and pays allocation overhead on every field
  access.
- **Opacity** — a reflect-based clone silently skips unexported fields, mishandles certain interface values, and can
  surprise you with shared references you didn't expect.
- **No control** — you can't say "clone this map only if the value satisfies a condition" or "shallow-copy this field
  but deep-copy everything else".

`doppel` inverts the priority order:

| Priority | Strategy                                 | When used                                      |
|----------|------------------------------------------|------------------------------------------------|
| 1        | **Manual clone** (your `Clone()` method) | Always, by default                             |
| 2        | **External Cloner[T]** (via `CloneWith`) | When clone logic needs injected context        |
| 3        | **Reflection fallback**                  | Phase 4 — only when neither of the above exist |

In Phase 1, reflection is not present at all. Every copy decision is written explicitly by you, composed from small
generic helpers.

---

## Design Philosophy

1. **Manual cloning is the default.** No reflection is used anywhere in Phase 1, not even for type identification.
   Generic helpers (`CloneSlice`, `CloneMap`, `ClonePointer`) are resolved entirely at compile time.

2. **Explicit over magic.** You write the `Clone()` method. `doppel` gives you the helpers to make it concise and safe,
   but the logic is always yours to read and reason about.

3. **Composable, not monolithic.** Each helper does exactly one thing. You wire them together inside your type's own
   `Clone()` method. There is no global state or hidden orchestration.

4. **Open for extension, closed for modification.** The `Cloner[T]` interface is the single extension point. Registering
   a custom cloner (Phase 2), adding per-field logic (Phase 3), or opting into reflection (Phase 4) will all be
   additive — nothing in Phase 1 changes.

5. **Errors carry context.** Every helper wraps failures with a field-path string (`core.WrapError`) so that when a
   clone fails deep inside a nested struct, the error message tells you exactly which field triggered it.

---

## Installation

```bash
go get github.com/seyallius/doppel
```

Requires **Go 1.26.2** or later (for range-over-integer and generic type inference improvements).

---

## Package Layout

```
github.com/seyallius/doppel/
│
├── doppel.go          Public API entry points
│                      Clone, MustClone, CloneWith, MustCloneWith
│
├── core/
│   ├── cloner.go      Cloner[T] interface, FuncCloner[T] adapter,
│   │                  SelfClonable[T] interface
│   └── errors.go      CloneError, WrapError, ErrNilSource
│
└── manual/
    ├── primitives.go  Identity[T], IdentityValue[T]
    ├── slice.go       CloneSlice[T], CloneSliceOf[T]
    ├── map.go         CloneMap[K,V], CloneMapOf[K,V]
    └── pointer.go     ClonePointer[T], ClonePointerOf[T]
```

**Dependency graph** (acyclic, no circular imports):

```
doppel  ──imports──▶  core
doppel  ──imports──▶  manual
manual  ──imports──▶  core
core    ──imports──▶  (stdlib only)
```

---

## Core Concepts

### Cloner[T]

```go
type Cloner[T any] interface {
Clone(src T) (T, error)
}
```

`Cloner[T]` is the central extension interface. Any value that can produce an independent deep copy of a `T` satisfies
it. It is the contract that the registry (Phase 2) and field-level customisation (Phase 3) will build on. For now, you
can create one with `core.NewFuncCloner`:

```go
addressCloner := core.NewFuncCloner(func (src Address) (Address, error) {
return Address{Street: src.Street, City: src.City}, nil
})
```

### SelfClonable[T]

```go
type SelfClonable[T any] interface {
Clone() (T, error)
}
```

An optional interface your types can implement. When a type owns all the state it needs to copy, `SelfClonable[T]` keeps
the clone logic co-located with the type and lets you call `doppel.Clone(value)` directly.

**Choose `SelfClonable` when:** the type knows everything it needs to clone itself — which is true for most domain
structs.

**Choose an external `Cloner[T]` when:** the clone function needs injected dependencies (a database handle, a feature
flag, a logger), or when you want to apply different clone strategies at different call sites without touching the type
itself.

### Identity helpers

```go
manual.Identity[T](src T) (T, error) // for use with CloneSlice / CloneMap / ClonePointer
manual.IdentityValue[T](src T) T // for use with CloneSliceOf / CloneMapOf / ClonePointerOf
```

For primitive Go types (`bool`, all integer and float types, `string`, `complex64/128`), a direct assignment is already
a complete deep copy — they carry no pointers. `Identity` and `IdentityValue` are no-op pass-throughs that express this
intent explicitly rather than relying on implicit behaviour.

---

## API Reference

### doppel.Clone

```go
func Clone[T any](src core.SelfClonable[T]) (T, error)
```

Produces a deep copy of `src` by calling `src.Clone()`. The compiler enforces that `src` implements `SelfClonable[T]`.

```go
user := &User{ID: 1, Name: "Alice"}
cloned, err := doppel.Clone(user) // cloned is *User, independent of user
```

### doppel.MustClone

```go
func MustClone[T any](src core.SelfClonable[T]) T
```

Like `Clone`, but panics on error instead of returning it. Intended for tests and program initialisation where a clone
failure is always a programming error.

```go
cloned := doppel.MustClone(user)
```

### doppel.CloneWith

```go
func CloneWith[T any](src T, cloner core.Cloner[T]) (T, error)
```

Produces a deep copy of `src` using an external `Cloner[T]`. Use this when the source type does not implement
`SelfClonable`, or when you want to supply a different clone strategy at a specific call site.

```go
cloner := core.NewFuncCloner(cloneAddress)
cloned, err := doppel.CloneWith(addr, cloner)
```

### doppel.MustCloneWith

```go
func MustCloneWith[T any](src T, cloner core.Cloner[T]) T
```

Like `CloneWith`, but panics on error.

---

### manual.CloneSlice / CloneSliceOf

```go
// Fallible element cloner — use when cloneElem can return an error.
func CloneSlice[T any](src []T, cloneElem func (T) (T, error)) ([]T, error)

// Infallible element cloner — use for primitive element types.
func CloneSliceOf[T any](src []T, cloneElem func (T) T) []T
```

`CloneSlice` creates an independent copy of `src`. `cloneElem` is called for every element. On error, a contextual
message with the failing index is returned.

```go
// Slice of primitive type — use Identity
tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])

// Slice of struct type — pass the struct's clone function
items, err := manual.CloneSlice(o.Items, cloneItem)

// Infallible shorthand for primitive slices
tags := manual.CloneSliceOf(u.Tags, manual.IdentityValue[string])
```

**Nil contract:** a nil `src` returns `(nil, nil)`. An empty (non-nil) `src` returns a fresh empty slice, never nil.

---

### manual.CloneMap / CloneMapOf

```go
// Fallible value cloner.
func CloneMap[K comparable, V any](src map[K]V, cloneVal func (V) (V, error)) (map[K]V, error)

// Infallible value cloner.
func CloneMapOf[K comparable, V any](src map[K]V, cloneVal func (V) V) map[K]V
```

`CloneMap` creates an independent copy of `src`. Map keys are comparable value types in Go and do not require a clone
step. Only values are cloned via `cloneVal`.

```go
// Map with primitive values
scores, err := manual.CloneMap(u.Scores, manual.Identity[int])

// Map with struct values
records, err := manual.CloneMap(store, cloneRecord)

// Conditional clone — only include values passing a predicate
active, err := manual.CloneMap(allUsers, func (u User) (User, error) {
if !u.Active {
return User{}, nil // zero-out inactive users
}
return u.Clone() // or however User is cloned
})
```

**Nil contract:** a nil `src` returns `(nil, nil)`. An empty (non-nil) `src` returns a fresh empty map.

---

### manual.ClonePointer / ClonePointerOf

```go
// Fallible value cloner.
func ClonePointer[T any](src *T, cloneVal func (T) (T, error)) (*T, error)

// Infallible value cloner.
func ClonePointerOf[T any](src *T, cloneVal func (T) T) *T
```

`ClonePointer` allocates a new `*T` and fills it with the result of `cloneVal(*src)`. The original and the clone never
share a pointer address.

```go
// Pointer to a struct
addr, err := manual.ClonePointer(u.Address, cloneAddress)

// Pointer to a primitive
label, err := manual.ClonePointer(u.Label, manual.Identity[string])
```

**Nil contract:** a nil `src` returns `(nil, nil)` without calling `cloneVal`.

---

## Usage Guide

### Step 1 — Simple struct (primitives only)

A struct whose fields are all primitive types (`string`, `int`, `bool`, `float64`, …) needs no helpers. A plain struct
literal in the `Clone()` method is sufficient, because primitives are value types — there is nothing heap-allocated to
share.

```go
type Address struct {
Street string
City   string
State  string
Zip    string
}

func (a Address) Clone() (Address, error) {
return Address{
Street: a.Street,
City:   a.City,
State:  a.State,
Zip:    a.Zip,
}, nil
}
```

Or, as a stand-alone function (useful when you want to pass it to `ClonePointer`):

```go
func cloneAddress(src Address) (Address, error) {
return Address{Street: src.Street, City: src.City, State: src.State, Zip: src.Zip}, nil
}
```

---

### Step 2 — Struct with a pointer field

Use `manual.ClonePointer` to allocate a new pointer and clone the pointed-to value independently.

```go
type ContactInfo struct {
Email   string
Phone   string
Address *Address
}

func cloneContactInfo(src ContactInfo) (ContactInfo, error) {
clonedAddr, err := manual.ClonePointer(src.Address, cloneAddress)
if err != nil {
return ContactInfo{}, core.WrapError("ContactInfo.Address", err)
}
return ContactInfo{
Email:   src.Email,
Phone:   src.Phone,
Address: clonedAddr,
}, nil
}
```

---

### Step 3 — Struct with slices and maps

Use `manual.CloneSlice` and `manual.CloneMap`. For primitive element/value types, pass `manual.Identity[T]`.

```go
type Profile struct {
Tags   []string
Scores map[string]int
Badges []string
}

func (p Profile) Clone() (Profile, error) {
tags, err := manual.CloneSlice(p.Tags, manual.Identity[string])
if err != nil {
return Profile{}, core.WrapError("Profile.Tags", err)
}

scores, err := manual.CloneMap(p.Scores, manual.Identity[int])
if err != nil {
return Profile{}, core.WrapError("Profile.Scores", err)
}

badges := manual.CloneSliceOf(p.Badges, manual.IdentityValue[string]) // infallible shorthand

return Profile{Tags: tags, Scores: scores, Badges: badges}, nil
}
```

---

### Step 4 — Full aggregate with nested structs

Everything composes. Each layer calls the clone function of the layer below it.

```go
type User struct {
ID      int64
Name    string
Active  bool
Contact ContactInfo
Tags    []string
Scores  map[string]int
}

func (u *User) Clone() (*User, error) {
if u == nil {
return nil, nil
}

contact, err := cloneContactInfo(u.Contact)
if err != nil {
return nil, core.WrapError("User.Contact", err)
}

tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
if err != nil {
return nil, core.WrapError("User.Tags", err)
}

scores, err := manual.CloneMap(u.Scores, manual.Identity[int])
if err != nil {
return nil, core.WrapError("User.Scores", err)
}

return &User{
ID:      u.ID,
Name:    u.Name,
Active:  u.Active,
Contact: contact,
Tags:    tags,
Scores:  scores,
}, nil
}

// At the call site:
cloned, err := doppel.Clone(user)
```

---

### Step 5 — External Cloner (no SelfClonable)

When a type does not implement `SelfClonable` — for example, a third-party struct you cannot modify — use
`core.NewFuncCloner` and `doppel.CloneWith`.

```go
// ThirdPartyConfig is a type you don't own.
type ThirdPartyConfig struct {
Host    string
Port    int
Options map[string]string
}

configCloner := core.NewFuncCloner(func (src ThirdPartyConfig) (ThirdPartyConfig, error) {
opts, err := manual.CloneMap(src.Options, manual.Identity[string])
if err != nil {
return ThirdPartyConfig{}, core.WrapError("ThirdPartyConfig.Options", err)
}
return ThirdPartyConfig{Host: src.Host, Port: src.Port, Options: opts}, nil
})

cloned, err := doppel.CloneWith(cfg, configCloner)
```

---

### Step 6 — Conditional / filtered cloning

Because you supply the clone function, you have full control over what goes into the clone. This is the preview of the
field-level customisation that Phase 3 will formalise.

```go
// Clone a map, but only carry over entries whose value is above a threshold.
aboveThreshold, err := manual.CloneMap(rawScores, func(score int) (int, error) {
if score < passingGrade {
return 0, nil // zero-out failing scores
}
return score, nil
})

// Clone a slice, skipping nil pointers entirely.
validUsers, err := manual.CloneSlice(allUsers, func (u *User) (*User, error) {
if u == nil {
return nil, nil // preserved as nil in clone
}
return u.Clone()
})
```

---

## Error Handling

Every fallible helper returns `(T, error)`. Errors are wrapped with `core.WrapError` at each layer, building a context
path that pinpoints the failure:

```
doppel: error cloning Order.Customer: doppel: error cloning User.Contact: doppel: error cloning ContactInfo.Address: pointer: <root cause>
```

Errors implement `Unwrap()`, so `errors.Is` and `errors.As` work correctly through the chain:

```go
cloned, err := doppel.Clone(order)
if err != nil {
var cloneErr *core.CloneError
if errors.As(err, &cloneErr) {
log.Printf("failed at: %s", cloneErr.Context)
}
}
```

For program initialisation and tests where a clone failure is always a bug, use `MustClone` / `MustCloneWith` to
panic-on-error rather than propagating the error manually.

---

## Nil Safety Contract

All helpers treat nil consistently and without error:

| Input                             | Output                      |
|-----------------------------------|-----------------------------|
| nil `*T` passed to `ClonePointer` | `nil, nil`                  |
| nil slice passed to `CloneSlice`  | `nil, nil`                  |
| nil map passed to `CloneMap`      | `nil, nil`                  |
| empty (non-nil) slice             | fresh empty slice (not nil) |
| empty (non-nil) map               | fresh empty map (not nil)   |

This means a clone faithfully preserves the nil-vs-empty distinction. If your original slice is `nil`, the clone's slice
is also `nil` — not `[]T{}`.

---

## Running Tests

```bash
# All tests
go test ./...

# With race detector (recommended for CI)
go test -race ./...

# Verbose output per package
go test -v ./...
go test -v ./manual/...

# Benchmarks only
go test -bench=. -benchmem ./...

# Benchmarks for a specific package
go test -bench=. -benchmem ./manual/...
```

---

## Benchmark Results

Indicative results on an Apple M2 (your numbers will vary). The key takeaway is the comparison between manual deep copy
and a plain shallow struct copy — the gap is the cost of the allocations you're explicitly making, with no reflection
overhead on top.

```
BenchmarkManualClone_Address-8            ~  8 ns/op      16 B/op    1 allocs/op
BenchmarkShallowCopy_Address-8            ~  1 ns/op       0 B/op    0 allocs/op
BenchmarkManualClone_User-8               ~ 45 ns/op     320 B/op    5 allocs/op
BenchmarkShallowCopy_User-8               ~  2 ns/op       0 B/op    0 allocs/op
BenchmarkManualClone_Order-8              ~ 90 ns/op     560 B/op    9 allocs/op
BenchmarkCloneSlice_Strings_1000-8        ~ 12 µs/op    8192 B/op    1 allocs/op
BenchmarkShallowCopy_Strings_1000-8       ~  1 µs/op    8192 B/op    1 allocs/op
BenchmarkCloneMap_StringInt_500-8         ~ 20 µs/op    5632 B/op    8 allocs/op
BenchmarkShallowCopy_StringInt_500-8      ~  9 µs/op    4096 B/op    3 allocs/op
```

For slices and maps of primitives, the gap between manual deep copy and shallow copy is entirely the cost of allocating
independent backing storage — unavoidable for true independence. There is no reflection tax.

---

## Roadmap

| Phase | Focus                                                          | Status     |
|-------|----------------------------------------------------------------|------------|
| **1** | Manual deep copy foundation (this release)                     | ✅ Complete |
| **2** | Cloner registry — per-type override, thread-safe lookup        | 🔜 Next    |
| **3** | Field-level cloners — per-field override and conditional logic | 📋 Planned |
| **4** | Reflection fallback — automatic clone for unregistered types   | 📋 Planned |
| **5** | Cycle detection — safe cloning of pointer graphs               | 📋 Planned |
| **6** | Benchmarking suite — cross-strategy comparison                 | 📋 Planned |
| **7** | API polish — `CloneWithOptions`, JSON-tag filtering, docs      | 📋 Planned |

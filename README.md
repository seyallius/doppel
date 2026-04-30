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
    - [doppel.CloneWithRegistry](#doppelclonewithregistry)
    - [manual.CloneSlice / CloneSliceOf](#manualcloneslice--clonesliceof)
    - [manual.CloneMap / CloneMapOf](#manualclonemap--clonemapof)
    - [manual.ClonePointer / ClonePointerOf](#manualclonepointer--clonepointerof)
    - [engine.Engine (Reflection Fallback)](#engineengine-reflection-fallback)
- [Usage Guide](#usage-guide)
    - [Step 1 — Simple struct (primitives only)](#step-1--simple-struct-primitives-only)
    - [Step 2 — Struct with a pointer field](#step-2--struct-with-a-pointer-field)
    - [Step 3 — Struct with slices and maps](#step-3--struct-with-slices-and-maps)
    - [Step 4 — Full aggregate with nested structs](#step-4--full-aggregate-with-nested-structs)
    - [Step 5 — External Cloner (no SelfClonable)](#step-5--external-cloner-no-selfclonable)
    - [Step 6 — Conditional / filtered cloning](#step-6--conditional--filtered-cloning)
    - [Step 7 — Cloner registry (Phase 2)](#step-7--cloner-registry-phase-2)
    - [Step 8 — Reflection fallback with cycle policies (Phase 4+5)](#step-8--reflection-fallback-with-cycle-policies-phase-45)
- [Error Handling](#error-handling)
- [Nil Safety Contract](#nil-safety-contract)
- [Struct Tag Directives](#struct-tag-directives)
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
| 1        | **Manual clone** (your `Clone()` method) | Always, by default — fastest path              |
| 2        | **External Cloner[T]** (via `CloneWith`) | When clone logic needs injected context        |
| 3        | **Registry Cloner[T]** (via `CloneWithRegistry`) | When you want type-level override without modifying source |
| 4        | **Reflection fallback** (`engine.Engine`) | Phase 4 — only when none of the above exist    |

In Phase 1, reflection is not present at all. Every copy decision is written explicitly by you, composed of small
generic helpers. Reflection is now available as a **controlled fallback** in Phase 4 — never the default, always the
last resort.

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

6. **Graceful degradation.** When manual clone logic isn't available, the reflection engine provides a safe, correct
   fallback — with clear performance trade-offs documented.

---

## Installation

```bash
go get github.com/seyallius/doppel
```

Requires **Go 1.25** or later (for generic type inference and range-over-integer improvements).

---

## Package Layout

```
github.com/seyallius/doppel/
│
├── doppel.go          Public API entry points
│                      Clone, MustClone, CloneWith, MustCloneWith,
│                      CloneWithRegistry
│
├── core/
│   ├── cloner.go      Cloner[T] interface, FuncCloner[T] adapter,
│   │                  SelfClonable[T] interface
│   └── errors.go      CloneError, WrapError, ErrNilSource, ErrNoCloner
│
├── manual/
│   ├── primitives.go  Identity[T], IdentityValue[T]
│   ├── slice.go       CloneSlice[T], CloneSliceOf[T]
│   ├── map.go         CloneMap[K,V], CloneMapOf[K,V]
│   └── pointer.go     ClonePointer[T], ClonePointerOf[T]
│
├── registry/          ← Phase 2
│   └── registry.go    Registry, New, Register[T], Lookup[T],
│                      Deregister[T], Has[T], LookupAny (bridge)
│
└── engine/            ← Phase 4+5
    ├── engine.go      Engine, New, NewWithOptions, Clone, TypeLookup interface
    └── cycle.go       CyclePolicy, Options, CycleError — configurable cycle handling
```

**Dependency graph** (acyclic, no circular imports):

```
doppel    ──imports──▶  core
doppel    ──imports──▶  manual
doppel    ──imports──▶  registry
manual    ──imports──▶  core
registry  ──imports──▶  core
engine    ──imports──▶  core
engine    ──imports──▶  (stdlib reflect)
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
it. It is the contract that the registry (Phase 2) and field-level customization (Phase 3) will build on. For now, you
can create one with `core.NewFuncCloner`:

```go
addressCloner := core.NewFuncCloner(func(src Address) (Address, error) {
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
manual.Identity[T](src T) (T, error)      // for use with CloneSlice / CloneMap / ClonePointer
manual.IdentityValue[T](src T) T          // for use with CloneSliceOf / CloneMapOf / ClonePointerOf
```

For primitive Go types (`bool`, all integer and float types, `string`, `complex64/128`), a direct assignment is already
a complete deep copy — they carry no pointers. `Identity` and `IdentityValue` are no-op pass-throughs that express this
intent explicitly rather than relying on implicit behavior.

---

## API Reference

### doppel.Clone

```go
func Clone[T any](src core.SelfClonable[T]) (T, error)
```

Produces a deep copy of `src` by calling `src.Clone()`. The compiler enforces that `src` implements `SelfClonable[T]`.

```go
user := &User{ID: 1, Name: "Alice"}
cloned, err := doppel.Clone(user)  // cloned is *User, independent of user
```

### doppel.MustClone

```go
func MustClone[T any](src core.SelfClonable[T]) T
```

Like `Clone`, but panics on error instead of returning it. Intended for tests and program initialization where a clone
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

### doppel.CloneWithRegistry

```go
func CloneWithRegistry[T any](src T, reg *registry.Registry) (T, error)
```

Produces a deep copy of `src` by walking the following priority chain:

1. **Registered `Cloner[T]`** — if `reg` contains a cloner for type T, it is called. This is the fastest path.
2. **`core.SelfClonable[T]` fallback** — if T implements `SelfClonable[T]`, its `Clone()` method is called. All existing `SelfClonable` types work without registration.
3. **`core.ErrNoCloner`** — returned when neither strategy is available.

Reflection is used only inside the registry for type key derivation — never for field access or value copying.

```go
reg := registry.New()
registry.Register(reg, core.NewFuncCloner(cloneAddress))

cloned, err := doppel.CloneWithRegistry(addr, reg)
```

### manual.CloneSlice / CloneSliceOf

```go
// Fallible element cloner — use when cloneElem can return an error.
func CloneSlice[T any](src []T, cloneElem func(T) (T, error)) ([]T, error)

// Infallible element cloner — use for primitive element types.
func CloneSliceOf[T any](src []T, cloneElem func(T) T) []T
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

### manual.CloneMap / CloneMapOf

```go
// Fallible value cloner.
func CloneMap[K comparable, V any](src map[K]V, cloneVal func(V) (V, error)) (map[K]V, error)

// Infallible value cloner.
func CloneMapOf[K comparable, V any](src map[K]V, cloneVal func(V) V) map[K]V
```

`CloneMap` creates an independent copy of `src`. Map keys are comparable value types in Go and do not require a clone
step. Only values are cloned via `cloneVal`.

```go
// Map with primitive values
scores, err := manual.CloneMap(u.Scores, manual.Identity[int])

// Map with struct values
records, err := manual.CloneMap(store, cloneRecord)

// Conditional clone — only include values passing a predicate
active, err := manual.CloneMap(allUsers, func(u User) (User, error) {
    if !u.Active {
        return User{}, nil // zero-out inactive users
    }
    return u.Clone() // or however User is cloned
})
```

**Nil contract:** a nil `src` returns `(nil, nil)`. An empty (non-nil) `src` returns a fresh empty map.

### manual.ClonePointer / ClonePointerOf

```go
// Fallible value cloner.
func ClonePointer[T any](src *T, cloneVal func(T) (T, error)) (*T, error)

// Infallible value cloner.
func ClonePointerOf[T any](src *T, cloneVal func(T) T) *T
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

### engine.Engine (Reflection Fallback)

```go
// Engine performs reflection-based deep copying as the LAST resort.
// Priority chain: Registered Cloner[T] → SelfClonable[T] → engine (reflection)
type Engine struct { /* unexported */ }

// Options configures an Engine at construction time.
type Options struct {
    CyclePolicy CyclePolicy  // zero value = PreserveShared
}

// CyclePolicy controls how Engine handles cyclic and shared pointer references.
type CyclePolicy int

const (
    PreserveShared CyclePolicy = iota  // default: reproduce exact graph topology
    BreakCycles                         // break back-edges → nil, acyclic output
    ErrorOnCycle                        // return *CycleError on any back-edge
)

// CycleError is returned when CyclePolicy is ErrorOnCycle and a cycle is detected.
type CycleError struct {
    Addr     uintptr  // pointer address where cycle detected
    TypeName string   // reflect.Type.String() for debugging
}

// New creates an Engine with default options (PreserveShared cycle policy).
func New(lookup TypeLookup) *Engine

// NewWithOptions creates an Engine with explicitly configured options.
func NewWithOptions(lookup TypeLookup, opts Options) *Engine

// Clone deep-copies src and returns a reflect.Value of the same type.
func (e *Engine) Clone(src reflect.Value) (reflect.Value, error)
```

The reflection engine is **never the default** — it is consulted only when:

- No `Cloner[T]` is registered for the type, AND
- The value does not implement `core.SelfClonable[T]`

**Features:**

- ✅ Kind-specific cloning: `Ptr`, `Struct`, `Slice`, `Map`, `Array`, `Interface`, primitives
- ✅ Struct tag support: `doppel:"-"` (skip), `doppel:"shallow"` (share backing)
- ✅ Configurable cycle handling via `CyclePolicy` (Phase 5)
- ✅ Shared reference preservation under `PreserveShared`
- ✅ Nil-safety: preserves `nil` vs `empty` distinction for slices/maps/pointers
- ✅ Error context: wraps failures with field-path via `core.WrapError`
- ✅ Concurrency: all mutable state is per-call; `Engine` is safe for concurrent use

**Limitations:**

- ❌ Unexported fields are skipped (use `SelfClonable[T]` to include them)
- ❌ `chan`, `func`, `unsafe.Pointer` are shallow-copied (reference semantics)
- ❌ Interface values cloned via concrete type; dynamic dispatch preserved

```go
// Example: Using the reflection fallback with BreakCycles policy
type Node struct {
    Value int
    Next  *Node
}

// Build a self-loop: n → n
n := &Node{Value: 42}
n.Next = n

// Use BreakCycles to produce an acyclic clone safe for JSON encoding
eng := engine.NewWithOptions(nil, engine.Options{
    CyclePolicy: engine.BreakCycles,
})
clonedVal, err := eng.Clone(reflect.ValueOf(n))
// clonedVal.(*Node).Next == nil  ← cycle broken, safe to marshal
```

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

configCloner := core.NewFuncCloner(func(src ThirdPartyConfig) (ThirdPartyConfig, error) {
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
field-level customization that Phase 3 will formalize.

```go
// Clone a map, but only carry over entries whose value is above a threshold.
aboveThreshold, err := manual.CloneMap(rawScores, func(score int) (int, error) {
    if score < passingGrade {
        return 0, nil // zero-out failing scores
    }
    return score, nil
})

// Clone a slice, skipping nil pointers entirely.
validUsers, err := manual.CloneSlice(allUsers, func(u *User) (*User, error) {
    if u == nil {
        return nil, nil // preserved as nil in clone
    }
    return u.Clone()
})
```

---

### Step 7 — Cloner registry

Register custom clone logic for types at application startup. The registry is thread-safe and can be shared across
goroutines.

```go
reg := registry.New()

// Register a custom cloner for Address
registry.Register(reg, core.NewFuncCloner(func(src Address) (Address, error) {
    return Address{City: strings.ToUpper(src.City)}, nil // transform on clone
}))

// Use CloneWithRegistry — it will find the registered cloner automatically
cloned, err := doppel.CloneWithRegistry(addr, reg)
```

**Priority chain in `CloneWithRegistry`:**

1. Registered `Cloner[T]` → 2. `SelfClonable[T]` → 3. `core.ErrNoCloner`

---

### Step 8 — Reflection fallback with cycle policies

When you have a type with no manual clone and no registered cloner, use the reflection engine as a safe fallback.
Configure cycle handling via `engine.Options`:

```go
type GraphNode struct {
    ID    int
    Links []*GraphNode  // may contain cycles
}

src := &GraphNode{ID: 1}
src.Links = []*GraphNode{src}  // self-referential cycle

// Option A: PreserveShared (default) — reproduce exact topology
eng1 := engine.New(nil)  // or NewWithOptions(nil, Options{CyclePolicy: PreserveShared})
cloned1, _ := eng1.Clone(reflect.ValueOf(src))
// cloned1.(*GraphNode).Links[0] == cloned1  ← cycle preserved

// Option B: BreakCycles — produce acyclic output for serialization
eng2 := engine.NewWithOptions(nil, engine.Options{
    CyclePolicy: engine.BreakCycles,
})
cloned2, _ := eng2.Clone(reflect.ValueOf(src))
// cloned2.(*GraphNode).Links[0] == nil  ← cycle broken, safe for JSON

// Option C: ErrorOnCycle — strict validation during development
eng3 := engine.NewWithOptions(nil, engine.Options{
    CyclePolicy: engine.ErrorOnCycle,
})
_, err := eng3.Clone(reflect.ValueOf(src))
// err.(*engine.CycleError) → "cycle detected at 0x... (type *GraphNode)"
```

> 💡 **Pro Tip**: Prefer manual `Clone()` implementations for performance-critical paths. Use the reflection fallback
> for prototyping, legacy integration, or when you need "just works" behavior without boilerplate. Benchmarks show
> manual cloning is ~3-6× faster than reflection.

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

For program initialization and tests where a clone failure is always a bug, use `MustClone` / `MustCloneWith` to
panic-on-error rather than propagating the error manually.

The reflection engine may also return `*engine.CycleError` when `CyclePolicy` is `ErrorOnCycle`:

```go
_, err := eng.Clone(reflect.ValueOf(cyclicData))
if cycleErr, ok := err.(*engine.CycleError); ok {
    log.Printf("cycle at 0x%x (%s)", cycleErr.Addr, cycleErr.TypeName)
}
```

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

## Struct Tag Directives

The `engine.Engine` respects the following `doppel` struct tags for fine-grained control:

```go
type Example struct {
    SkipMe    string   `doppel:"-"`       // ← skipped; clone gets zero value
    ShareMe   []string `doppel:"shallow"` // ← shallow copy; shares backing array
    DeepClone string   // ← default: deep copy recursively
}
```

| Tag                | Behavior                                                                                                                |
|--------------------|-------------------------------------------------------------------------------------------------------------------------|
| `doppel:"-"`       | Field is skipped entirely; clone receives the zero value for that field                                                 |
| `doppel:"shallow"` | Field is assigned without recursing; clone shares the field's value (useful for immutable or reference-semantics types) |

> ⚠️ **Note**: Struct tags are only processed by the reflection engine (`engine/`). Manual clone methods and registry
> cloners ignore tags — you control the logic explicitly.

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

# Save benchmarks for comparison
just bench-save output=bench.txt

# Compare with benchstat
just benchstat file=bench.txt
```

---

## Benchmark Results

Indicative results on **11th Gen Intel(R) Core(TM) i5-11400H @ 2.70GHz** (your numbers will vary).

### Doppel (Manual) vs Reflection Comparison

| Benchmark        | Doppel (ns/op) | Reflection (ns/op) | Speedup   | Doppel (B/op) | Reflection (B/op) | Doppel (allocs/op) | Reflection (allocs/op) |
|------------------|----------------|--------------------|-----------|---------------|-------------------|--------------------|------------------------|
| `Score`          | 22.03 ± 8%     | 131.8 ± 3%         | **~6×**   | 24            | 96                | 1                  | 4                      |
| `User`           | 309.9 ± 2%     | 1.193µ ± 4%        | **~4×**   | 528           | 968               | 6                  | 18                     |
| `Order`          | 615.9 ± 4%     | 2.183µ ± 4%        | **~3.5×** | 1.109Ki       | 1.742Ki           | 11                 | 34                     |
| `UserLargeSlice` | 8.363µ ± 3%    | 32.76µ ± 0%        | **~4×**   | 32.44Ki       | 32.87Ki           | 6                  | 18                     |
| `UserLargeMap`   | 29.26µ ± 0%    | 99.41µ ± 4%        | **~3.4×** | 53.64Ki       | 101.4Ki           | 10                 | 2,016                  |

**Key takeaways:**
- 🚀 Manual cloning is **3-6× faster** than reflection-based cloning across all test cases
- 🧠 Manual cloning uses **~40-95% fewer allocations**, especially noticeable with large maps
- ⚡ The gap grows with complexity — nested structs and large collections benefit most from explicit cloning
- 🔁 Reflection fallback is still **correct and safe** — use it when convenience outweighs performance needs

### Engine-Specific Benchmarks

```
BenchmarkEngine_PlainStruct            2272365     520.8 ns/op     88 B/op     5 allocs/op
BenchmarkEngine_NestedStruct            736201    1586 ns/op    360 B/op    14 allocs/op
BenchmarkEngine_LargeSlice               12734   93976 ns/op  16240 B/op  1003 allocs/op
BenchmarkEngine_LargeMap                  8905  126708 ns/op  63640 B/op  2005 allocs/op
BenchmarkEngine_WithTypeLookup_Hit    14146173    83.50 ns/op     48 B/op     1 allocs/op  ← registry hit!
BenchmarkEngine_SelfClonable           1671700   715.6 ns/op    232 B/op     8 allocs/op
BenchmarkEngine_ShallowBaseline      975549486   1.229 ns/op      0 B/op     0 allocs/op
```

> 💡 **Pro Tip**: When using the reflection engine, register cloners for hot-path types via `TypeLookup`. The
> `BenchmarkEngine_WithTypeLookup_Hit` shows that a registry hit reduces reflection overhead to near-zero.

---

## Roadmap

**Summary**

| Phase | Focus                                                          | Status     |
|-------|----------------------------------------------------------------|------------|
| **1** | Manual deep copy foundation (this release)                     | ✅ Complete |
| **2** | Cloner registry — per-type override, thread-safe lookup        | ✅ Complete |
| **3** | Field-level cloners — per-field override and conditional logic | 🔜 Next    |
| **4** | Reflection fallback — automatic clone for unregistered types   | ✅ Complete |
| **5** | Cycle detection — configurable policies for cyclic graphs      | ✅ Complete |
| **6** | Benchmarking suite — cross-strategy comparison                 | 📋 Planned |
| **7** | API polish — `CloneWithOptions`, JSON-tag filtering, docs      | 📋 Planned |

**Detailed**

# 🚀 PHASE 1 — Manual Deep Copy Foundation (NO reflection)

| **Category**     | **Details**                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
|------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Goal**         | Build the core cloning system using explicit/manual cloning only                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| **Requirements** | 1. Create a Go module named `doppel`.<br>2. Define core interface: `type Cloner[T any] interface { Clone(src T) (T, error) }`<br>3. Provide built-in manual cloners for: Primitive types, Structs (explicit cloning functions), Slices, Maps, Pointer types.<br>4. DO NOT use reflection anywhere in this phase.<br>5. Design helper functions: `CloneSlice`, `CloneMap`, `ClonePointer`.<br>6. Allow struct-specific clone functions like: `func (u *User) Clone() *User`.<br>7. Ensure: Nil safety, No shared references, Proper allocation.<br>8. Organize code cleanly: `/doppel`, `/core`, `/manual`.<br>9. Add unit tests for: Nested structs, Maps, Slices, Pointer fields.<br>10. Follow clean code practices and proper naming conventions. |

---

# 🚀 PHASE 2 — Cloner Registry (Extensibility Layer)

| **Category**     | **Details**                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
|------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Goal**         | Allow users to plug in custom cloning logic                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| **Requirements** | 1. Implement registry system: `type Registry struct { typeCloners map[reflect.Type]any }`<br>2. Allow registration: `func Register[T any](r *Registry, cloner core.Cloner[T])`<br>3. Lookup logic: If custom cloner exists → use it, Otherwise fallback to manual cloning.<br>4. Still DO NOT implement reflection-based cloning yet. Reflection is only allowed for TYPE IDENTIFICATION.<br>5. Support: Per-type cloner override, Thread-safe registry.<br>6. Design API: `doppel.CloneWithRegistry(obj, registry)`<br>7. Add tests: Custom struct cloner override, Ensure override is respected. |

---

# 🚀 PHASE 3 — Field-Level Customization (🔥 this is a hot feature)

| **Category**     | **Details**                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
|------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Goal**         | Fine-grained control per field                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| **Requirements** | 1. Allow users to define cloners per struct field.<br>2. Design: `type FieldCloner func(value any) (any, error)`<br>3. Extend registry: `RegisterFieldCloner(structType reflect.Type, fieldName string, cloner FieldCloner)`<br>4. Behavior: If field-level cloner exists → use it, Else fallback to type cloner, Else fallback to manual clone.<br>5. Example use case: Clone map only if value satisfies condition.<br>6. Ensure: No reflection-based cloning yet, Reflection ONLY for field discovery.<br>7. Add tests: Conditional cloning, Partial field cloning. |

---

# 🚀 PHASE 4 — Reflection Fallback (Controlled, Not Default) ✅ COMPLETE

| **Category**       | **Details**                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
|--------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Goal**           | Introduce reflection as a fallback only when manual clone or registry cloner is unavailable                                                                                                                                                                                                                                                                                                                                                                                                                             |
| **Priority Chain** | `Registered Cloner[T]` → `core.SelfClonable[T]` → `engine.Engine` (reflection) — reflection is **never** the default, always the last resort                                                                                                                                                                                                                                                                                                                                                                            |
| **Requirements**   | ✅ Implement reflection-based deep copy engine in `engine/`<br>✅ Support: Structs, Maps, Slices, Pointers, Arrays, Interfaces (best-effort)<br>✅ Handle: Nested objects, Zero values, Unexported fields (skip safely)<br>✅ Respect `doppel:"-"` and `doppel:"shallow"` struct tags<br>✅ Cycle safety via `visited` map (shared reference preservation)<br>✅ Concurrency safety via per-call `cloneState`<br>✅ Error context via `core.WrapError`<br>✅ Bridge to registry via `TypeLookup` interface (acyclic dependency) |
| **API**            | `engine.New(lookup TypeLookup) *Engine`<br>`func (e *Engine) Clone(src reflect.Value) (reflect.Value, error)`<br>`registry.LookupAny(t reflect.Type)` — bridge method                                                                                                                                                                                                                                                                                                                                                   |
| **When to Use**    | - Prototyping with third-party types you can't modify<br>- Legacy code migration where manual `Clone()` isn't feasible yet<br>- Dynamic scenarios where type isn't known at compile time<br>⚠️ **Performance note**: Manual cloning is 3-6× faster; use reflection fallback judiciously                                                                                                                                                                                                                                 |
| **Limitations**    | - Unexported fields are skipped (use `SelfClonable[T]` to include them)<br>- `chan`, `func`, `unsafe.Pointer` are shallow-copied (reference semantics)<br>- Interface values cloned via concrete type; dynamic dispatch preserved                                                                                                                                                                                                                                                                                       |
| **Tests**          | ✅ Primitives, nested structs, slices, maps, arrays, interfaces<br>✅ Nil vs empty distinction<br>✅ Struct tag directives<br>✅ Cyclic graphs (self-loop, two-node cycle)<br>✅ Shared pointer preservation<br>✅ Error propagation with context<br>✅ Concurrency safety (50 goroutines)                                                                                                                                                                                                                                     |
| **Benchmarks**     | See `engine/engine_bench_test.go` and `benchstat.txt` — reflection path is ~3-6× slower than manual, but still correct and safe                                                                                                                                                                                                                                                                                                                                                                                         |

---

# 🚀 PHASE 5 — Cycle Detection (Configurable Policies) ✅ COMPLETE

| **Category**       | **Details**                                                                                                                                                                                                                        |
|--------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Goal**           | Provide configurable cycle-handling strategies for pointer graphs                                                                                                                                                                  |
| **Policies**       | ✅ `PreserveShared` (default): reproduce exact topology, deduplicate shared refs<br>✅ `BreakCycles`: break back-edges → nil, produce acyclic output<br>✅ `ErrorOnCycle`: return `*CycleError` on any cycle for strict validation    |
| **API**            | `engine.Options{CyclePolicy: engine.BreakCycles}`<br>`engine.NewWithOptions(lookup, opts)`<br>`engine.CycleError{Addr, TypeName}`                                                                                                  |
| **When to Use**    | - `PreserveShared`: general-purpose cloning, faithful reproduction<br>- `BreakCycles`: serialization (JSON/YAML), tree conversion, avoiding infinite loops<br>- `ErrorOnCycle`: development-time assertions, data model validation |
| **Implementation** | ✅ `cloneState.inStack` for DFS cycle tracking<br>✅ Policy-aware `clonePointer` and `cloneMap`<br>✅ Backward-compatible: `New()` defaults to `PreserveShared`                                                                       |
| **Tests**          | ✅ Self-loop preservation/breaking/error<br>✅ Two-node cycle handling<br>✅ Shared reference deduplication under `PreserveShared`                                                                                                    |

---

# 🚀 PHASE 6 — Performance & Benchmarking

| **Category**     | **Details**                                                                                                                                                                                                                                                                                                                           |
|------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Goal**         | Prove your design is legit and guide users on optimization                                                                                                                                                                                                                                                                            |
| **Requirements** | 1. Write benchmarks comparing: Manual cloning, Reflection fallback, Custom cloners.<br>2. Optimize: Allocation patterns, Map pre-sizing, Slice copying.<br>3. Avoid unnecessary interface{} conversions.<br>4. Output benchmark results clearly with `benchstat` integration.<br>5. Add `just` recipes for easy benchmark management. |

---

# 🚀 PHASE 7 — Developer Experience (Polish)

| **Category**     | **Details**                                                                                                                                                                                                                                                                                                                                     |
|------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Goal**         | Make it usable, not just powerful                                                                                                                                                                                                                                                                                                               |
| **Requirements** | 1. Provide clean API: `doppel.Clone(obj)`, `doppel.CloneWithOptions(obj, opts)`, `doppel.NewRegistry()`.<br>2. Add: Documentation, Usage examples, README, godoc comments.<br>3. Ensure: Minimal boilerplate, Clear error messages, IDE-friendly autocomplete.<br>4. Consider: CLI tool for generating `Clone()` stubs from struct definitions. |

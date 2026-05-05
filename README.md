<p align="center">
  <strong>doppel</strong>
</p>

<p align="center">
  Your data's doppelgänger — deep copies without side effects.
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/seyallius/doppel"><img src="https://pkg.go.dev/badge/github.com/seyallius/doppel.svg" alt="Go Reference"></a>
  <a href="https://github.com/seyallius/doppel/blob/main/go.mod"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go" alt="Go Version"></a>
  <a href="https://github.com/seyallius/doppel/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License"></a>
</p>

---

## What is doppel?

doppel is a Go library for safe, explicit deep cloning of complex data structures. It provides a layered architecture that prioritizes manual, zero-reflection cloning by default, with an optional reflection fallback for types you don't control — all fully composable and extensible through a type-safe registry.

Go assignment is a shallow copy. Structs with pointer fields, slices, and maps silently share memory between "originals" and "copies," leading to subtle bugs. doppel solves this by giving you full control over every field, every allocation, and every edge case.

## Installation

```bash
go get github.com/seyallius/doppel
```

Zero external dependencies — only the Go standard library.

## Quick start

### Implement `Clone()` on your type

```go
type User struct {
    ID     int64
    Name   string
    Tags   []string
    Scores map[string]int
}

func (u *User) Clone() (*User, error) {
    if u == nil { return nil, nil }

    tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
    if err != nil {
        return nil, core.WrapError("User.Tags", err)
    }

    scores, err := manual.CloneMap(u.Scores, manual.Identity[int])
    if err != nil {
        return nil, core.WrapError("User.Scores", err)
    }

    return &User{ID: u.ID, Name: u.Name, Tags: tags, Scores: scores}, nil
}
```

### Clone it

```go
cloned, err := doppel.Clone(original)

// Mutate the original — the clone is unaffected
original.Tags[0] = "mutated"
original.Scores["math"] = 0
```

### Or use automatic deep copy with selective override

```go
reg := registry.New()
registry.RegisterField[BigStruct, *Address](reg, "Address", core.NewFuncCloner(cloneAddr))

cloned, err := doppel.CloneDeep(bigStruct, reg)
// 198 fields cloned by reflection, 2 by custom field cloner
```

## Architecture

doppel is built in five layers, each adding capability on top of the previous:

```
Phase 1 — Manual Deep Copy Foundation
  core.Cloner[T]  ·  core.SelfClonable[T]  ·  core.FuncCloner[T]
  manual.CloneSlice / CloneMap / ClonePointer / Identity

Phase 2 — Cloner Registry
  registry.New()  ·  Register[T]  ·  Lookup[T]  ·  Deregister[T]

Phase 3 — Field-Level Customization
  registry.RegisterField[T, F]  ·  doppel.CloneDeep

Phase 4 — Reflection Engine
  engine.New()  ·  engine.Clone()  ·  SelfClonable detection

Phase 5 — Cycle & Sharing Policy
  PreserveShared  ·  BreakCycles  ·  ErrorOnCycle
```

### Clone dispatch chain

When you call `CloneDeep`, the library walks this priority chain:

```
Registered Cloner[T]  →  SelfClonable[T]  →  Field Cloner  →  Reflection Engine
     (fastest)              (type-owned)      (per-field)      (automatic)
```

## Choosing the right API

| You have...              | You want...                          | Use                                    |
|--------------------------|--------------------------------------|----------------------------------------|
| A type with `Clone()`    | Just clone it                        | `doppel.Clone(value)`                  |
| A type without `Clone()` | Clone with a custom function         | `doppel.CloneWith(value, cloner)`      |
| A type without `Clone()` | Clone via registry lookup            | `doppel.CloneWithRegistry(value, reg)` |
| Any type                 | Full deep copy (reflection fallback) | `doppel.CloneDeep(value, reg)`         |
| A large struct           | Override only a few fields           | `CloneDeep` + `registry.RegisterField` |

## Features

- **Zero-reflection manual cloning** — generics-based helpers with no runtime overhead
- **Type-safe registry** — register cloners for types you don't own
- **Field-level customization** — override individual struct fields without writing a full `Clone()` method
- **Reflection fallback** — automatic deep copy for any Go value when no manual clone exists
- **Struct tag directives** — `doppel:"-"`, `doppel:"shallow"`, `doppel:"readonly"`, `doppel:"clone"`, `doppel:"deep"`
- **Cycle handling** — three policies: `PreserveShared`, `BreakCycles`, `ErrorOnCycle`
- **Contextual errors** — field-path annotated errors with `errors.Is`/`errors.As` support
- **Thread-safe** — all public types are safe for concurrent use
- **Zero dependencies** — only the Go standard library

## Performance

| Approach                       | ns/op  | B/op   | allocs/op |
|--------------------------------|--------|--------|-----------|
| Manual `Clone()` (User)        | ~750   | ~480   | ~12       |
| `CloneDeep` with type cloner   | ~1,500 | ~800   | ~15       |
| `CloneDeep` with field cloners | ~2,800 | ~1,000 | ~25       |
| `CloneDeep` pure reflection    | ~3,500 | ~1,200 | ~30       |
| Shallow copy (baseline)        | ~1     | 0      | 0         |

Registry lookup adds ~85 ns. Reflection adds ~300-500 ns per struct. See the [Benchmarks](old_docs/benchmarks.md) page for detailed data.

## Documentation

Comprehensive step-by-step documentation is available in the [`docs/`](old_docs/) directory:

| #  | Page                                                | Topic                                                |
|----|-----------------------------------------------------|------------------------------------------------------|
| 1  | [Getting Started](old_docs/getting-started.md)      | Installation, first clone, API selection             |
| 2  | [SelfClonable Interface](old_docs/self-clonable.md) | The `Clone()` method pattern                         |
| 3  | [Manual Helpers](old_docs/manual-helpers.md)        | `CloneSlice`, `CloneMap`, `ClonePointer`, `Identity` |
| 4  | [Cloner Registry](old_docs/registry.md)             | Register cloners for external types                  |
| 5  | [Field-Level Cloners](old_docs/field-cloners.md)    | Per-field clone overrides                            |
| 6  | [Reflection Engine](old_docs/reflection-engine.md)  | Automatic deep copy fallback                         |
| 7  | [Struct Tags](old_docs/struct-tags.md)              | `doppel:"..."` directive reference                   |
| 8  | [Cycle & Sharing Policy](old_docs/cycle-policy.md)  | Handling cyclic and shared references                |
| 9  | [Error Handling](old_docs/error-handling.md)        | `CloneError`, `WrapError`, error inspection          |
| 10 | [Patterns & Best Practices](old_docs/patterns.md)   | Real-world usage patterns                            |
| 11 | [Benchmarks](old_docs/benchmarks.md)                | Performance data                                     |
| 12 | [API Reference](old_docs/api-reference.md)          | Complete function signatures                         |

The docs are designed to be read in order — each page builds on concepts from the previous one. Start with [Getting Started](old_docs/getting-started.md).

## Development

Requires [just](https://github.com/casey/just) for task running.

```bash
just test              # Run all tests
just test-coverage     # Generate coverage report
just bench             # Run all benchmarks
just all-checks        # Format, vet, lint, staticcheck
just generate-nav      # Update doc navigation links
```

## License

[MIT](./LICENSE)

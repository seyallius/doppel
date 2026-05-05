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

doppel is a Go library for safe, explicit deep cloning of complex data structures. It provides a minimal, zero-reflection API built around composable generic helpers that you wire together inside your type's `Clone()` method.

Go assignment is a shallow copy. Structs with pointer fields, slices, and maps silently share memory between "originals" and "copies," leading to subtle bugs. doppel solves this by giving you full control over every field, every allocation, and every edge case.

---

## Why doppel?

| Principle        | What it means                                                             |
|------------------|---------------------------------------------------------------------------|
| **Manual first** | No reflection, no magic, maximum speed. You write the clone logic.        |
| **Composable**   | `CloneSlice`, `CloneMap`, `ClonePointer` wire together in your `Clone()`. |
| **Explicit**     | Every clone path is visible and auditable — no hidden behavior.           |
| **Future-proof** | Struct tags prepare for automatic code generation (coming soon).          |

---

## Architecture

```
core          — Interfaces: Cloner[T], SelfClonable[T], FuncCloner[T], error types, tag contract
manual        — Generic helpers: CloneSlice, CloneMap, ClonePointer, Identity
doppel        — Top-level API: Clone[T], MustClone[T]
```

### Clone dispatch

```
doppel.Clone(user)  →  user.Clone()  →  manual.CloneSlice + CloneMap + ClonePointer
                        (zero overhead)      (your code, fully visible)
```

There is no priority chain, no registry, and no reflection. You own the clone path.

---

## Quick start

### 1. Define your type

```go
type User struct {
    ID     int64
    Name   string
    Active bool
    Tags   []string
    Scores map[string]int
}
```

### 2. Implement `Clone()`

```go
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

    return &User{ID: u.ID, Name: u.Name, Active: u.Active, Tags: tags, Scores: scores}, nil
}
```

### 3. Call `doppel.Clone`

```go
cloned, err := doppel.Clone(original)

// Mutate the original — the clone is unaffected
original.Tags[0] = "mutated"
// cloned.Tags[0] still has the original value
```

---

## Installation

```bash
go get github.com/seyallius/doppel
```

Requires **Go 1.25** or later. Zero external dependencies.

---

## Features

- **Zero-reflection manual cloning** — generics-based helpers, no runtime overhead
- **Composable helpers** — `CloneSlice`, `CloneMap`, `ClonePointer` wire together in `Clone()`
- **Contextual errors** — field-path annotated errors with `errors.Is`/`errors.As` support
- **Nil-safety** — all helpers preserve nil/empty distinction
- **Thread-safe** — all public types are safe for concurrent use
- **Struct tags** — `doppel:"-"`, `doppel:"shallow"`, `doppel:"clone"`, `doppel:"deep"` for future generator
- **Zero dependencies** — only the Go standard library

---

## Performance

| Approach                    | ns/op | B/op   | allocs/op |
|-----------------------------|-------|--------|-----------|
| Manual `Clone()` (User)     | ~750  | ~480   | ~12       |
| Manual `Clone()` (Order)    | ~1200 | ~800   | ~20       |
| Shallow copy (baseline)     | ~1    | 0      | 0         |

See [Benchmarks](docs/benchmarks.md) for detailed data.

---

## Documentation

| # | Page                                         | Topic                               |
|---|----------------------------------------------|-------------------------------------|
| 1 | [Getting Started](./docs/getting-started.md) | Install, first clone, API guide     |
| 2 | [SelfClonable](./docs/self-clonable.md)      | The `Clone()` method pattern        |
| 3 | [Manual Helpers](./docs/manual-helpers.md)   | CloneSlice, CloneMap, ClonePointer  |
| 4 | [Struct Tags](./docs/struct-tags.md)         | Tag directives for future generator |
| 5 | [Benchmarks](./docs/benchmarks.md)           | Performance data                    |
| 6 | [API Reference](./docs/api-reference.md)     | Complete function signatures        |

---

## Development

```bash
just test              # Run all tests
just test-coverage     # Generate coverage report
just bench             # Run all benchmarks
just all-checks        # Format, vet, lint, staticcheck
just generate-nav      # Update doc navigation links
```

## License

[MIT](./LICENSE)

# API Reference

Complete API reference for all doppel packages.

---

## Package `doppel` — Top-level API

### `Clone[T]`

```go
func Clone[T any](src core.SelfClonable[T]) (T, error)
```

Produces a deep copy of `src` by calling `src.Clone()`. `T` must satisfy `core.SelfClonable[T]`.

```go
cloned, err := doppel.Clone(user) // calls user.Clone()
```

### `MustClone[T]`

```go
func MustClone[T any](src core.SelfClonable[T]) T
```

Like `Clone` but panics on error.

### `CloneWith[T]`

```go
func CloneWith[T any](src T, cloner core.Cloner[T]) (T, error)
```

Produces a deep copy using an external `Cloner[T]`.

```go
cloner := core.NewFuncCloner(func(a Address) (Address, error) { return a, nil })
cloned, err := doppel.CloneWith(addr, cloner)
```

### `MustCloneWith[T]`

```go
func MustCloneWith[T any](src T, cloner core.Cloner[T]) T
```

Like `CloneWith` but panics on error.

### `CloneWithRegistry[T]`

```go
func CloneWithRegistry[T any](src T, reg *registry.Registry) (T, error)
```

Priority chain: `Registered Cloner[T]` → `SelfClonable[T]` → `core.ErrNoCloner`.

```go
cloned, err := doppel.CloneWithRegistry(value, reg)
```

### `CloneDeep[T]`

```go
func CloneDeep[T any](src T, reg *registry.Registry) (T, error)
```

Full priority chain: `Registered Cloner[T]` → `SelfClonable[T]` → `Field Cloner` → `Reflection Engine`.

```go
// With field cloners
cloned, err := doppel.CloneDeep(value, reg)

// Pure reflection (no registry)
cloned, err := doppel.CloneDeep(value, nil)
```

### `MustCloneDeep[T]`

```go
func MustCloneDeep[T any](src T, reg *registry.Registry) T
```

Like `CloneDeep` but panics on error.

---

## Package `core` — Interfaces & Error Types

### `Cloner[T]`

```go
type Cloner[T any] interface {
    Clone(src T) (T, error)
}
```

The central extension interface. Implementations must:
- Never return a value that shares mutable memory with `src`.
- Return a non-nil error only when cloning cannot complete safely.
- Be safe for concurrent calls.

### `SelfClonable[T]`

```go
type SelfClonable[T any] interface {
    Clone() (T, error)
}
```

Optional interface for types that manage their own deep-copy logic.

### `FuncCloner[T]`

```go
type FuncCloner[T any] struct { /* unexported */ }

func NewFuncCloner[T any](cloneFn func(T) (T, error)) *FuncCloner[T]
func (fc *FuncCloner[T]) Clone(src T) (T, error)
```

Adapts a plain function to the `Cloner[T]` interface.

### `CloneError`

```go
type CloneError struct {
    Context string
    Cause   error
}

func (e *CloneError) Error() string
func (e *CloneError) Unwrap() error
```

Contextual error with field-path information.

### `WrapError`

```go
func WrapError(context string, cause error) error
```

Creates a `CloneError` annotating `cause` with `context`.

### Sentinel errors

```go
var ErrNilSource = errors.New("doppel: clone source is nil")
var ErrNoCloner = errors.New("doppel/registry: no cloner registered and type does not implement SelfClonable")
```

---

## Package `manual` — Generic Helpers

### `CloneSlice[T]`

```go
func CloneSlice[T any](src []T, cloneElem func(T) (T, error)) ([]T, error)
```

Creates an independent deep copy of a slice. Nil returns `(nil, nil)`. On error, returns nil with index context.

### `CloneSliceOf[T]`

```go
func CloneSliceOf[T any](src []T, cloneElem func(T) T) []T
```

Infallible variant. Nil returns nil.

### `CloneMap[K, V]`

```go
func CloneMap[K comparable, V any](src map[K]V, cloneVal func(V) (V, error)) (map[K]V, error)
```

Creates an independent deep copy of a map. Nil returns `(nil, nil)`. On error, returns nil with key context.

### `CloneMapOf[K, V]`

```go
func CloneMapOf[K comparable, V any](src map[K]V, cloneVal func(V) V) map[K]V
```

Infallible variant. Nil returns nil.

### `ClonePointer[T]`

```go
func ClonePointer[T any](src *T, cloneVal func(T) (T, error)) (*T, error)
```

Creates an independent deep copy of the pointed-to value. Nil returns `(nil, nil)`. Allocates fresh memory via `new(T)`.

### `ClonePointerOf[T]`

```go
func ClonePointerOf[T any](src *T, cloneVal func(T) T) *T
```

Infallible variant. Nil returns nil.

### `Identity[T]`

```go
func Identity[T any](src T) (T, error)
```

No-op cloner for fallible helpers. Returns `(src, nil)`.

### `IdentityValue[T]`

```go
func IdentityValue[T any](src T) T
```

No-op cloner for infallible helpers. Returns `src`.

---

## Package `registry` — Cloner Store

### Constructor

```go
func New() *Registry
```

Creates an empty, thread-safe registry.

### Type-level cloners

```go
func Register[T any](r *Registry, cloner core.Cloner[T])
func Lookup[T any](r *Registry) (core.Cloner[T], bool)
func Deregister[T any](r *Registry)
func Has[T any](r *Registry) bool
func (r *Registry) Len() int
func (r *Registry) LookupAny(t reflect.Type) (func(reflect.Value) (reflect.Value, error), bool)
```

### Field-level cloners

```go
func RegisterField[T any, F any](r *Registry, fieldName string, cloner core.Cloner[F])
func LookupField[T any, F any](r *Registry, fieldName string) (core.Cloner[F], bool)
func DeregisterField[T any](r *Registry, fieldName string) bool
func HasField[T any](r *Registry, fieldName string) bool
func (r *Registry) FieldLen() int
func (r *Registry) LookupAnyField(structType reflect.Type, fieldName string) (func(reflect.Value) (reflect.Value, error), bool)
```

`RegisterField` panics if:
- `T` is not a struct type.
- `fieldName` does not exist on `T`.
- `fieldName` is an unexported field.

---

## Package `engine` — Reflection Engine

### `CyclePolicy`

```go
type CyclePolicy int

const (
    PreserveShared CyclePolicy = iota // Default: reproduce sharing and cycles
    BreakCycles                       // Back-edges become nil
    ErrorOnCycle                      // Returns *CycleError on back-edge
)
```

### `Options`

```go
type Options struct {
    CyclePolicy CyclePolicy
}
```

### `CycleError`

```go
type CycleError struct {
    Addr     uintptr
    TypeName string
}

func (e *CycleError) Error() string
```

### Constructor

```go
func New(lookup TypeLookup) *Engine
func NewWithOptions(lookup TypeLookup, opts Options) *Engine
```

`TypeLookup` and `FieldLookup` are satisfied automatically by `*registry.Registry`.

### Clone method

```go
func (e *Engine) Clone(src reflect.Value) (reflect.Value, error)
```

Performs a recursive deep copy. Returns a `reflect.Value` of the same type as `src`.

### Interfaces

```go
type TypeLookup interface {
    LookupAny(t reflect.Type) (func(reflect.Value) (reflect.Value, error), bool)
}

type FieldLookup interface {
    LookupAnyField(structType reflect.Type, fieldName string) (func(reflect.Value) (reflect.Value, error), bool)
}
```

Both are implemented by `*registry.Registry`. The engine auto-detects `FieldLookup` when the `TypeLookup` also implements it.

---

## Quick reference: API dispatch comparison

| API | Registry | SelfClonable | Field Cloners | Reflection |
|-----|----------|-------------|---------------|------------|
| `Clone` | — | Required | — | — |
| `CloneWith` | — | — | — | — |
| `CloneWithRegistry` | Checked | Checked | — | No |
| `CloneDeep` | Checked | Checked | Checked | Yes |
| `engine.Clone` | Checked | Checked | Checked | Yes |

---

## Quick reference: choosing between fallible and infallible

| Helper | Use when | Example |
|--------|---------|---------|
| `CloneSlice` | Element cloner can fail | `manual.CloneSlice(items, score.Clone)` |
| `CloneSliceOf` | Element cloner cannot fail | `manual.CloneSliceOf(tags, manual.IdentityValue[string])` |
| `CloneMap` | Value cloner can fail | `manual.CloneMap(data, cloneValue)` |
| `CloneMapOf` | Value cloner cannot fail | `manual.CloneMapOf(scores, manual.IdentityValue[int])` |
| `ClonePointer` | Value cloner can fail | `manual.ClonePointer(addr, cloneAddress)` |
| `ClonePointerOf` | Value cloner cannot fail | `manual.ClonePointerOf(label, manual.IdentityValue[string])` |

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • API Reference
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="benchmarks.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Benchmarks</span>
                </span>
            </a></div>
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: center; align-items: center;">
            <a href="INDEX.md" style="display: flex; align-items: center; justify-content: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #8b5cf6, #6d28d9); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(139, 92, 246, 0.3); text-align: center;">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">⌂</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Return to</span>
                    <span style="font-size: 1rem; font-weight: 600;">Index</span>
                </span>
            </a>
        </div>
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"></div>
    </div>
</div>
<!-- /Navigation -->


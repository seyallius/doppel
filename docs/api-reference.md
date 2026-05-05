# API Reference

Complete API reference for all doppel packages.

---

## Package `doppel` — Top-level API

### `Clone[T]`

```go
func Clone[T any](src core.SelfClonable[T]) (T, error)
```

Produces a deep copy of `src` by calling `src.Clone()`. `T` must satisfy `core.SelfClonable[T]`. This is a zero-overhead dispatch — a direct call to the type's own `Clone()` method.

```go
cloned, err := doppel.Clone(user)
```

### `MustClone[T]`

```go
func MustClone[T any](src core.SelfClonable[T]) T
```

Like `Clone` but panics on error. Use in tests and initialization code.

```go
cloned := doppel.MustClone(original)
```

---

## Package `core` — Interfaces & Error Types

### `Cloner[T]`

```go
type Cloner[T any] interface {
    Clone(src T) (T, error)
}
```

External clone logic interface for types you don't own.

### `SelfClonable[T]`

```go
type SelfClonable[T any] interface {
    Clone() (T, error)
}
```

Interface for types that manage their own deep-copy logic. This is the primary interface in doppel — implement it on your types to use `doppel.Clone`.

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

### `TagDirective`

```go
type TagDirective struct {
    Skip    bool
    Shallow bool
    Deep    bool
    Clone   bool
}
```

Parsed result of a doppel struct tag.

### `ParseTag`

```go
func ParseTag(tagValue string) TagDirective
```

Parses a doppel struct tag value into a `TagDirective`. Zero reflection.

### `TagKey`

```go
const TagKey = "doppel"
```

The struct tag key.

### `ErrNoCloner`

```go
var ErrNoCloner = errors.New("doppel: no cloner available")
```

Sentinel error for custom dispatch code.

---

## Package `manual` — Generic Helpers

### `CloneSlice[T]`

```go
func CloneSlice[T any](src []T, cloneElem func(T) (T, error)) ([]T, error)
```

Independent deep copy of a slice. Nil returns `(nil, nil)`. On error, returns nil with index context.

### `CloneMap[K, V]`

```go
func CloneMap[K comparable, V any](src map[K]V, cloneVal func(V) (V, error)) (map[K]V, error)
```

Independent deep copy of a map. Nil returns `(nil, nil)`. On error, returns nil with key context.

### `ClonePointer[T]`

```go
func ClonePointer[T any](src *T, cloneVal func(T) (T, error)) (*T, error)
```

Independent deep copy of the pointed-to value. Nil returns `(nil, nil)`. Allocates fresh memory.

### `Identity[T]`

```go
func Identity[T any](src T) (T, error)
```

No-op cloner for fallible helpers. Returns `(src, nil)`.

### `IdentityValue[T]`

```go
func IdentityValue[T any](src T) T
```

No-op cloner for custom infallible functions. Returns `src`.

---

## Quick reference

| Helper          | Error format                                |
|-----------------|---------------------------------------------|
| `CloneSlice`   | `doppel: CloneSlice index [2]: <your error>` |
| `CloneMap`     | `doppel: CloneMap key [myKey]: <your error>` |
| `ClonePointer` | `doppel: pointer: <your error>`              |

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        doppel Documentation &bull; API Reference
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;">
            <a href="benchmarks.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">&#8592;</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Benchmarks</span>
                </span>
            </a>
        </div>
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: center; align-items: center;">
            <a href="INDEX.md" style="display: flex; align-items: center; justify-content: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #8b5cf6, #6d28d9); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(139, 92, 246, 0.3); text-align: center;">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">&#8962;</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Return to</span>
                    <span style="font-size: 1rem; font-weight: 600;">Index</span>
                </span>
            </a>
        </div>
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"></div>
    </div>
</div>

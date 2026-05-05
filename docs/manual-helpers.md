# Manual Helpers

The `manual` package provides generic, zero-reflection helpers for deep-copying slices, maps, and pointers. These are the building blocks you compose inside your type's `Clone()` method. Every helper is a plain Go generic function — no reflection, no interface dispatch, just fast, explicit code.

---

## Overview

| Helper           | Signature                                                     | When to use                                               |
|------------------|---------------------------------------------------------------|-----------------------------------------------------------|
| `CloneSlice`     | `(src []T, cloneElem func(T) (T, error)) ([]T, error)`        | Element cloning can fail (nested structs with validation) |
| `CloneSliceOf`   | `(src []T, cloneElem func(T) T) []T`                          | Element cloning cannot fail (primitives)                  |
| `CloneMap`       | `(src map[K]V, cloneVal func(V) (V, error)) (map[K]V, error)` | Value cloning can fail                                    |
| `CloneMapOf`     | `(src map[K]V, cloneVal func(V) V) map[K]V`                   | Value cloning cannot fail                                 |
| `ClonePointer`   | `(src *T, cloneVal func(T) (T, error)) (*T, error)`           | Pointed-to value cloning can fail                         |
| `ClonePointerOf` | `(src *T, cloneVal func(T) T) *T`                             | Pointed-to value cloning cannot fail                      |
| `Identity`       | `(src T) (T, error)`                                          | No-op cloner for fallible helpers                         |
| `IdentityValue`  | `(src T) T`                                                   | No-op cloner for infallible helpers                       |

Each collection type has two variants: a **fallible** version (returns error) and an **infallible** version (no error return). Use the infallible version with `IdentityValue` for primitives to get the cleanest call sites.

---

## Identity — the primitive no-op

For primitive Go types (`bool`, `int`, `string`, `float`, etc.), assignment IS a complete deep copy. The `Identity` and `IdentityValue` helpers make this intent explicit:

```go
// Fallible — use with CloneSlice, CloneMap, ClonePointer
tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])

// Infallible — use with CloneSliceOf, CloneMapOf, ClonePointerOf
tags := manual.CloneSliceOf(u.Tags, manual.IdentityValue[string])
```

Both return the input unchanged. The difference is purely in the signature — `Identity` returns `(T, error)` and `IdentityValue` returns just `T`.

---

## CloneSlice — deep copy slices

### Fallible variant (CloneSlice)

Use `CloneSlice` when the element cloner can return an error, for example when cloning a slice of structs that have their own `Clone()` method:

```go
type Order struct {
    Items []Score
}

// Score.Clone() returns (Score, error)
items, err := manual.CloneSlice(order.Items, func(s Score) (Score, error) {
    return s.Clone()
})
if err != nil {
    return nil, core.WrapError("Order.Items", err)
}
```

**Behavior:**
- `nil` slice returns `(nil, nil)` — nil is preserved as nil.
- Empty (non-nil) slice returns a fresh empty slice, not nil.
- On error, returns `nil` and a contextual error that identifies the failing index: `doppel: CloneSlice index [2]: <error>`.

### Infallible variant (CloneSliceOf)

Use `CloneSliceOf` for primitive element types where cloning cannot fail:

```go
tags := manual.CloneSliceOf(u.Tags, manual.IdentityValue[string])
ids  := manual.CloneSliceOf(u.IDs, manual.IdentityValue[int])
```

**Behavior:** Same nil/empty semantics as `CloneSlice`, but no error return.

### Element transformation

You can transform elements during cloning. This is useful for normalization or enrichment:

```go
// Append "_copy" to every tag
tags, err := manual.CloneSlice(src.Tags, func(s string) (string, error) {
    return s + "_copy", nil
})
```

---

## CloneMap — deep copy maps

### Fallible variant (CloneMap)

```go
scores, err := manual.CloneMap(u.Scores, manual.Identity[int])
metadata, err := manual.CloneMap(o.Metadata, manual.Identity[string])
```

Map keys must be comparable (string, numeric, etc.). Comparable types are value types in Go, so keys are copied automatically during iteration. Only values are cloned — this is why `CloneMap` takes a `cloneVal` function, not a `cloneKey` function.

**Behavior:**
- `nil` map returns `(nil, nil)`.
- Empty (non-nil) map returns a fresh empty map.
- On error, returns `nil` and an error annotated with the failing key: `doppel: CloneMap key [myKey]: <error>`.

### Infallible variant (CloneMapOf)

```go
scores := manual.CloneMapOf(u.Scores, manual.IdentityValue[int])
```

### Value transformation

```go
// Double every value
cloned, err := manual.CloneMap(src, func(v int) (int, error) {
    return v * 2, nil
})
```

### Conditional cloning

You can use the clone function to selectively transform values — for example, zeroing out entries below a threshold:

```go
// Zero out values below 10
cloned, err := manual.CloneMap(src, func(v int) (int, error) {
    if v < 10 {
        return 0, nil
    }
    return v, nil
})
```

---

## ClonePointer — deep copy pointers

### Fallible variant (ClonePointer)

`ClonePointer` creates a **new allocation** for the clone. The original and clone never share the same pointer address:

```go
type User struct {
    Address *Address
}

// cloneAddress is a standalone function (Address doesn't implement Clone)
func cloneAddress(src Address) (Address, error) {
    return Address{Street: src.Street, City: src.City}, nil
}

// Inside User.Clone():
addr, err := manual.ClonePointer(u.Address, cloneAddress)
if err != nil {
    return nil, core.WrapError("User.Address", err)
}
```

**Behavior:**
- `nil` pointer returns `(nil, nil)` — no error, no allocation.
- Non-nil pointer creates a fresh `new(T)`, calls `cloneVal(*src)`, and stores the result.
- The original and clone have different addresses. Mutating the original's pointed-to value does not affect the clone.

### Infallible variant (ClonePointerOf)

```go
label := manual.ClonePointerOf(u.Label, manual.IdentityValue[string])
```

### Independence check

```go
original := &Address{City: "Springfield"}
cloned, _ := manual.ClonePointer(original, cloneAddress)

original.City = "Mutated"
// cloned.City is still "Springfield"
// cloned != original (different pointer addresses)
```

---

## Nil vs. empty semantics

All helpers preserve the distinction between nil and empty:

| Input           | CloneSlice output                   | CloneSliceOf output                 |
|-----------------|-------------------------------------|-------------------------------------|
| `nil`           | `nil`                               | `nil`                               |
| `[]string{}`    | `[]string{}` (non-nil)              | `[]string{}` (non-nil)              |
| `[]string{"a"}` | `[]string{"a"}` (new backing array) | `[]string{"a"}` (new backing array) |

This matters because Go code sometimes uses nil to mean "not set" and empty to mean "explicitly empty." doppel preserves this distinction.

---

## Error context

When a fallible helper encounters an error from your clone function, it wraps it with context to help you locate the failure:

| Helper         | Error format                                 |
|----------------|----------------------------------------------|
| `CloneSlice`   | `doppel: CloneSlice index [2]: <your error>` |
| `CloneMap`     | `doppel: CloneMap key [myKey]: <your error>` |
| `ClonePointer` | `doppel: pointer: <your error>`              |

Use `errors.Is` and `errors.As` to inspect the underlying cause:

```go
sentinel := errors.New("clone failed")
_, err := manual.CloneSlice(src, func(s string) (string, error) {
    return "", sentinel
})
// err.Error() contains "index [0]"
// errors.Is(err, sentinel) == true
```

---

## Performance characteristics

These helpers have zero reflection overhead. The only allocations are:

| Helper         | Allocations                      |
|----------------|----------------------------------|
| `CloneSlice`   | 1 slice header + 1 backing array |
| `CloneMap`     | 1 map header                     |
| `ClonePointer` | 1 `new(T)` allocation            |

The actual element cloning allocations depend on what your `cloneElem` / `cloneVal` function does. For primitive types with `Identity`, there are no per-element allocations at all — the slice or map is the only allocation.

---

## What's next?

- **[Cloner Registry](registry.md)** — Register external cloners for types you don't own, and use `CloneWithRegistry` and `CloneDeep`.
- **[Error Handling](error-handling.md)** — Deep dive into `CloneError`, `WrapError`, and error inspection patterns.

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Manual Helpers
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="self-clonable.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">SelfClonable</span>
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
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"><a href="registry.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Registry</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
            </a></div>
    </div>
</div>
<!-- /Navigation -->


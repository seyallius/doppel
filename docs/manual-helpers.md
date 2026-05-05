# Manual Helpers

The `manual` package provides generic, zero-reflection helpers for deep-copying slices, maps, and pointers. These are the building blocks you compose inside your type's `Clone()` method.

---

## Overview

| Helper          | Signature                                                     | When to use                |
|-----------------|---------------------------------------------------------------|-----------------------------|
| `CloneSlice`   | `(src []T, cloneElem func(T) (T, error)) ([]T, error)`        | Element cloning can fail   |
| `CloneMap`     | `(src map[K]V, cloneVal func(V) (V, error)) (map[K]V, error)` | Value cloning can fail     |
| `ClonePointer` | `(src *T, cloneVal func(T) (T, error)) (*T, error)`           | Pointed-to value can fail |
| `Identity`     | `(src T) (T, error)`                                          | No-op for fallible helpers |
| `IdentityValue`| `(src T) T`                                                   | No-op for custom functions |

---

## Identity — the primitive no-op

For primitive Go types (`bool`, `int`, `string`, `float`, etc.), assignment IS a complete deep copy. `Identity` makes this intent explicit:

```go
tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
scores, err := manual.CloneMap(u.Scores, manual.Identity[int])
```

---

## CloneSlice — deep copy slices

```go
items, err := manual.CloneSlice(order.Items, func(s Score) (Score, error) {
    return s.Clone()
})
```

**Behavior:**
- `nil` slice returns `(nil, nil)`.
- Empty (non-nil) slice returns a fresh empty slice, not nil.
- On error, returns nil with index context: `doppel: CloneSlice index [2]: <error>`.

---

## CloneMap — deep copy maps

Map keys must be comparable (string, numeric, etc.). Only values are cloned — keys are copied automatically.

```go
scores, err := manual.CloneMap(u.Scores, manual.Identity[int])
```

**Behavior:**
- `nil` map returns `(nil, nil)`.
- On error, returns nil with key context: `doppel: CloneMap key [myKey]: <error>`.

---

## ClonePointer — deep copy pointers

Creates a **new allocation** for the clone. The original and clone never share the same address:

```go
addr, err := manual.ClonePointer(u.Address, cloneAddress)
```

**Behavior:**
- `nil` pointer returns `(nil, nil)`.
- Non-nil pointer creates a fresh `new(T)`, calls `cloneVal(*src)`, and stores the result.

---

## Nil vs. empty semantics

| Input           | CloneSlice output          |
|-----------------|------------------------------|
| `nil`           | `nil`                        |
| `[]string{}`    | `[]string{}` (new backing) |
| `[]string{"a"}` | `[]string{"a"}` (new backing) |

---

## Error context

| Helper         | Error format                                |
|----------------|---------------------------------------------|
| `CloneSlice`   | `doppel: CloneSlice index [2]: <your error>` |
| `CloneMap`     | `doppel: CloneMap key [myKey]: <your error>` |
| `ClonePointer` | `doppel: pointer: <your error>`              |

---

## Performance characteristics

These helpers have zero reflection overhead. The only allocations are:

| Helper         | Allocations                      |
|----------------|----------------------------------|
| `CloneSlice`   | 1 slice header + 1 backing array |
| `CloneMap`     | 1 map header                     |
| `ClonePointer` | 1 `new(T)` allocation            |

For primitive types with `Identity`, there are no per-element allocations at all.

---

## What's next?

- **[Struct Tags](struct-tags.md)** — Tag directives for the future code generator.

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        doppel Documentation &bull; Manual Helpers
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;">
            <a href="self-clonable.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">&#8592;</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">SelfClonable</span>
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
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;">
            <a href="struct-tags.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Struct Tags</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">&#8594;</span>
            </a>
        </div>
    </div>
</div>

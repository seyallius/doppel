# Manual Helpers

The `manual` package provides composable, generic helper functions for
building `Clone()` implementations by hand. Each helper is zero-overhead — no
reflection, no hidden allocations.

> **Tip:** If you prefer not to write these by hand, use `doppelgen` to
> generate `Clone()` methods automatically from struct tags. See
> [Getting Started — Code Generator](getting-started.md#code-generator-doppelgen).

---

## `CloneSlice[T]`

```go
func CloneSlice[T any](src []T, cloneElem func(T) (T, error)) ([]T, error)
```

Creates an independent deep copy of a slice.

| `src`         | Result                                      |
|---------------|---------------------------------------------|
| `nil`         | `(nil, nil)` — nil is preserved.            |
| Empty non-nil | Fresh empty slice.                          |
| Non-empty     | New backing array with each element cloned. |

**Parameters:**

- `src` — the source slice.
- `cloneElem` — called once per element to produce its copy.

**Example (primitive elements):**

```go
tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
```

**Example (struct elements):**

```go
items, err := manual.CloneSlice(u.Items, func(v Item) (Item, error) {
    return v.Clone()
})
```

---

## `CloneMap[K, V]`

```go
func CloneMap[K comparable, V any](src map[K]V, cloneVal func(V) (V, error)) (map[K]V, error)
```

Creates an independent deep copy of a map. Keys (comparable value types) are
copied automatically during iteration; only values are cloned.

| `src`         | Result                           |
|---------------|----------------------------------|
| `nil`         | `(nil, nil)` — nil is preserved. |
| Empty non-nil | Fresh empty map.                 |
| Non-empty     | New map with cloned values.      |

**Example (primitive values):**

```go
scores, err := manual.CloneMap(u.Scores, manual.Identity[int])
```

**Example (struct values):**

```go
entries, err := manual.CloneMap(u.Entries, func(v Entry) (Entry, error) {
    return v.Clone()
})
```

---

## `ClonePointer[T]`

```go
func ClonePointer[T any](src *T, cloneVal func(T) (T, error)) (*T, error)
```

Creates an independent deep copy of the value behind a pointer. The result
is placed into freshly allocated memory — the original and clone never share
addresses.

| `src`   | Result                                         |
|---------|------------------------------------------------|
| `nil`   | `(nil, nil)` — nil is preserved without error. |
| Non-nil | New `*T` pointing to cloned value.             |

**Example (primitive pointer):**

```go
name, err := manual.ClonePointer(u.Name, manual.Identity[string])
```

**Example (struct pointer):**

```go
addr, err := manual.ClonePointer(u.Address, func(v Address) (Address, error) {
    return v.Clone()
})
```

---

## `Identity[T]` and `IdentityValue[T]`

```go
func Identity[T any](src T) (T, error)
func IdentityValue[T any](src T) T
```

No-op helpers that return `src` unchanged. For all primitive Go types, a
direct copy *is* a complete deep copy — these helpers make that intent
explicit.

- **`Identity[T]`** — returns `(T, error)` for use with fallible helpers
  (`CloneSlice`, `CloneMap`, `ClonePointer`).
- **`IdentityValue[T]`** — returns `T` for use when callers want an
  infallible `func(T) T` signature.

```go
tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
ids, err := manual.CloneSlice(u.IDs, manual.Identity[int])
```

---

## Putting it all together

Compose the helpers inside your type's `Clone()` method:

```go
func (u *User) Clone() (*User, error) {
    if u == nil {
        return nil, nil
    }

    tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
    if err != nil {
        return nil, core.WrapError("User.Tags", err)
    }

    scores, err := manual.CloneMap(u.Scores, manual.Identity[int])
    if err != nil {
        return nil, core.WrapError("User.Scores", err)
    }

    addr, err := manual.ClonePointer(u.Address, func(v Address) (Address, error) {
        return v.Clone()
    })
    if err != nil {
        return nil, core.WrapError("User.Address", err)
    }

    return &User{
        ID:     u.ID,
        Name:   u.Name,
        Tags:   tags,
        Scores: scores,
        Address: addr,
    }, nil
}
```

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
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"><a href="struct-tags.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Struct Tags</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
            </a></div>
    </div>
</div>
<!-- /Navigation -->


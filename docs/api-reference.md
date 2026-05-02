# 🔧 API Reference

> Complete documentation for all public `doppel` APIs. Copy-paste ready with examples. ✨

---

## Public Entry Points

### `doppel.Clone[T]`

```go
// Clone produces a deep copy of src by calling src.Clone().
// The compiler enforces that src implements core.SelfClonable[T].
// Returns (T, error) where error includes contextual field-path on failure.
func Clone[T any](src core.SelfClonable[T]) (T, error)
```

**Example**:
```go
user := &User{ID: 1, Name: "Alice"}
cloned, err := doppel.Clone(user)  // cloned is *User, independent of user
if err != nil {
    log.Fatalf("clone failed: %v", err)
}
```

---

### `doppel.MustClone[T]`

```go
// MustClone is like Clone, but panics on error instead of returning it.
// Intended for tests and program initialization where clone failure is always a bug.
func MustClone[T any](src core.SelfClonable[T]) T
```

**Example**:
```go
// In test setup or init()
cloned := doppel.MustClone(user)  // panics if clone fails
```

---

### `doppel.CloneWith[T]`

```go
// CloneWith produces a deep copy of src using an external Cloner[T].
// Use when src does not implement SelfClonable, or when you need
// a different clone strategy at a specific call site.
func CloneWith[T any](src T, cloner core.Cloner[T]) (T, error)
```

**Example**:
```go
cloner := core.NewFuncCloner(cloneAddress)
cloned, err := doppel.CloneWith(addr, cloner)
```

---

### `doppel.MustCloneWith[T]`

```go
// MustCloneWith is like CloneWith, but panics on error.
func MustCloneWith[T any](src T, cloner core.Cloner[T]) T
```

---

### `doppel.CloneWithRegistry[T]`

```go
// CloneWithRegistry produces a deep copy of src by walking a priority chain:
// 1. Registered Cloner[T] in reg (fastest)
// 2. core.SelfClonable[T] fallback (if T implements it)
// 3. core.ErrNoCloner if neither is available
// Reflection is used only for type key derivation — never for field access.
func CloneWithRegistry[T any](src T, reg *registry.Registry) (T, error)
```

**Example**:
```go
reg := registry.New()
registry.Register(reg, core.NewFuncCloner(cloneAddress))

cloned, err := doppel.CloneWithRegistry(addr, reg)
```

**Priority Chain**:
```
Registered Cloner[T] → SelfClonable[T] → core.ErrNoCloner
```

---

## Manual Helpers

### `manual.CloneSlice` / `CloneSliceOf`

```go
// CloneSlice creates an independent copy of src by calling cloneElem for each element.
// Returns contextual error with failing index on failure.
// Nil src returns (nil, nil). Empty non-nil src returns fresh empty slice.
func CloneSlice[T any](src []T, cloneElem func(T) (T, error)) ([]T, error)

// CloneSliceOf is the infallible shorthand for primitive element types.
func CloneSliceOf[T any](src []T, cloneElem func(T) T) []T
```

**Examples**:
```go
// Slice of primitives — use Identity
tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])

// Slice of structs — pass the struct's clone function
items, err := manual.CloneSlice(o.Items, cloneItem)

// Infallible shorthand for primitives
tags := manual.CloneSliceOf(u.Tags, manual.IdentityValue[string])
```

---

### `manual.CloneMap` / `CloneMapOf`

```go
// CloneMap creates an independent copy of src by cloning values via cloneVal.
// Keys are comparable value types and do not require cloning.
// Nil src returns (nil, nil). Empty non-nil src returns fresh empty map.
func CloneMap[K comparable, V any](src map[K]V, cloneVal func(V) (V, error)) (map[K]V, error)

// CloneMapOf is the infallible shorthand for primitive value types.
func CloneMapOf[K comparable, V any](src map[K]V, cloneVal func(V) V) map[K]V
```

**Examples**:
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
    return u.Clone()
})
```

---

### `manual.ClonePointer` / `ClonePointerOf`

```go
// ClonePointer allocates a new *T and fills it with cloneVal(*src).
// Original and clone never share a pointer address.
// Nil src returns (nil, nil) without calling cloneVal.
func ClonePointer[T any](src *T, cloneVal func(T) (T, error)) (*T, error)

// ClonePointerOf is the infallible shorthand for primitive pointed-to types.
func ClonePointerOf[T any](src *T, cloneVal func(T) T) *T
```

**Examples**:
```go
// Pointer to a struct
addr, err := manual.ClonePointer(u.Address, cloneAddress)

// Pointer to a primitive
label, err := manual.ClonePointer(u.Label, manual.Identity[string])
```

---

## Nil Safety Contract

All helpers treat nil consistently and without error:

| Input                       | Output                      |
|-----------------------------|-----------------------------|
| `nil *T` to `ClonePointer`  | `(nil, nil)`                |
| `nil []T` to `CloneSlice`   | `(nil, nil)`                |
| `nil map[K]V` to `CloneMap` | `(nil, nil)`                |
| Empty (non-nil) slice       | Fresh empty slice (not nil) |
| Empty (non-nil) map         | Fresh empty map (not nil)   |

✅ Clones faithfully preserve the nil-vs-empty distinction.

> 💡 **Remember**: Use the fallible versions (`CloneSlice`, etc.) when cloning complex types that can fail. Use the
> infallible `*Of` versions for primitives to reduce boilerplate. (◕‿◕)✧

<!--

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • API Reference
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap;">
        <div style="flex: 1; min-width: 200px;">
            <a href="core-concepts.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
            <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
            <span style="display: flex; flex-direction: column; line-height: 1.3;">
          <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
          <span style="font-size: 1rem; font-weight: 600;">Core Concepts</span>
        </span>
        </a>
        </div>
        <div style="flex: 1; min-width: 200px;">
            <a href="usage-guide.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
        <span style="display: flex; flex-direction: column; line-height: 1.3;">
          <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
          <span style="font-size: 1rem; font-weight: 600;">Usage Guide</span>
        </span>
            <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
        </a>
        </div>
    </div>
</div>
<!-- /Navigation -->


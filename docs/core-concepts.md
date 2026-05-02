# 🧠 Core Concepts

> The foundational interfaces and helpers that power `doppel`. Understand these, and you master the library. ✨

---

## `Cloner[T]` Interface

```go
// Cloner is the central extension interface for deep copying values of type T.
// Any value that can produce an independent deep copy of a T satisfies this contract.
type Cloner[T any] interface {
    Clone(src T) (T, error)
}
```

`Cloner[T]` is the contract that the registry (Phase 2) and field-level customization (Phase 3) build on.

### Creating a Cloner

Use `core.NewFuncCloner` to wrap a plain function:

```go
addressCloner := core.NewFuncCloner(func(src Address) (Address, error) {
    return Address{Street: src.Street, City: src.City}, nil
})
```

### When to Implement `Cloner[T]`

✅ The clone logic needs injected dependencies (DB handle, logger, feature flags)  
✅ You want different clone strategies at different call sites without touching the type  
✅ Cloning third-party types you cannot modify

---

## `SelfClonable[T]` Interface

```go
// SelfClonable is an optional interface types can implement to own their clone logic.
// When implemented, doppel.Clone(value) calls value.Clone() directly.
type SelfClonable[T any] interface {
    Clone() (T, error)
}
```

### Example Implementation

```go
type User struct {
    ID   int64
    Name string
    Tags []string
}

func (u *User) Clone() (*User, error) {
    if u == nil {
        return nil, nil
    }
    tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
    if err != nil {
        return nil, core.WrapError("User.Tags", err)
    }
    return &User{ID: u.ID, Name: u.Name, Tags: tags}, nil
}
```

### When to Implement `SelfClonable[T]`

✅ The type knows everything it needs to clone itself (most domain structs)  
✅ You want the fastest possible clone path (direct method call)  
✅ You prefer co-locating clone logic with the type definition

### Choosing Between `SelfClonable` and External `Cloner`

| Factor           | Prefer `SelfClonable[T]`    | Prefer External `Cloner[T]`         |
|------------------|-----------------------------|-------------------------------------|
| **Ownership**    | You own the type            | Third-party or shared type          |
| **Dependencies** | None needed                 | Needs injected context (DB, logger) |
| **Flexibility**  | One clone strategy per type | Multiple strategies per call site   |
| **Performance**  | ⚡ Fastest (direct call)     | ⚡ Fast (interface call)             |

---

## Identity Helpers

For primitive Go types (`bool`, integers, floats, `string`, `complex64/128`), a direct assignment is already a complete deep copy — they carry no pointers.

### `manual.Identity[T]` (Fallible)

```go
// Identity is a no-op pass-through for primitive types that returns (T, error).
// Use with CloneSlice, CloneMap, ClonePointer when the helper expects a fallible function.
func Identity[T any](src T) (T, error) {
    return src, nil
}
```

**Usage**:
```go
tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
```

### `manual.IdentityValue[T]` (Infallible)

```go
// IdentityValue is a no-op pass-through for primitive types that returns T directly.
// Use with CloneSliceOf, CloneMapOf, ClonePointerOf for infallible shorthand.
func IdentityValue[T any](src T) T {
    return src
}
```

**Usage**:
```go
// Infallible shorthand for primitive slices
tags := manual.CloneSliceOf(u.Tags, manual.IdentityValue[string])
```

### Why Two Versions?

| Helper             | Return Type  | Use Case                                                      |
|--------------------|--------------|---------------------------------------------------------------|
| `Identity[T]`      | `(T, error)` | When the helper expects fallible cloning (e.g., `CloneSlice`) |
| `IdentityValue[T]` | `T`          | When you want infallible shorthand (e.g., `CloneSliceOf`)     |

> 💡 **Pro Tip**: For primitives, always use the identity helpers — they make your intent explicit and avoid accidental
> shared references. ✧◝(⁰▿⁰)◜✧

<!--

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Core Concepts
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap;">
        <div style="flex: 1; min-width: 200px;">
            <a href="philosophy.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
            <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
            <span style="display: flex; flex-direction: column; line-height: 1.3;">
          <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
          <span style="font-size: 1rem; font-weight: 600;">Philosophy</span>
        </span>
        </a>
        </div>
        <div style="flex: 1; min-width: 200px;">
            <a href="api-reference.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
        <span style="display: flex; flex-direction: column; line-height: 1.3;">
          <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
          <span style="font-size: 1rem; font-weight: 600;">API Reference</span>
        </span>
            <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
        </a>
        </div>
    </div>
</div>
<!-- /Navigation -->


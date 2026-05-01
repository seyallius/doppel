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

<!-- Navigation -->
<div style="margin-top: 3rem; padding: 1.5rem; border-top: 2px solid #e1e4e8; border-radius: 6px; background: #f6f8fa;">
  <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 1rem;">
    <div style="flex: 1; min-width: 200px;">
      <a href="./philosophy.md" style="display: inline-flex; align-items: center; padding: 0.75rem 1rem; background: #0366d6; color: white; text-decoration: none; border-radius: 6px; font-weight: 600; transition: background 0.2s;">
        <span style="margin-right: 0.5rem;">←</span>
        <div style="display: flex; flex-direction: column;">
          <span style="font-size: 0.75rem; opacity: 0.9;">Previous:</span>
          <span>Philosophy</span>
        </div>
      </a>
    </div>
    <div style="flex: 1; min-width: 200px; text-align: right;">
      <a href="./api-reference.md" style="display: inline-flex; align-items: center; padding: 0.75rem 1rem; background: #28a745; color: white; text-decoration: none; border-radius: 6px; font-weight: 600; transition: background 0.2s;">
        <div style="display: flex; flex-direction: column; margin-right: 0.5rem;">
          <span style="font-size: 0.75rem; opacity: 0.9;">Next:</span>
          <span>API Reference →</span>
        </div>
      </a>
    </div>
  </div>
  <div style="margin-top: 1rem; text-align: center; color: #586069; font-size: 0.875rem;">
    <span>📚 doppel Documentation • Core Concepts</span>
  </div>
</div>

<style>
@media (max-width: 768px) {
  div[style*="margin-top: 3rem"] div[style*="display: flex"] {
    flex-direction: column !important;
  }
  div[style*="margin-top: 3rem"] div[style*="text-align: right"] {
    text-align: left !important;
  }
}
</style>

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

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->
<div class="doppel-nav-container"
     style="margin-top: 3rem; padding: 1.75rem; border-top: 1px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c); box-shadow: 0 8px 24px rgba(0,0,0,0.4);">
    <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 1rem;">
        <div style="flex: 1; min-width: 200px;">
            <a href="philosophy.md" class="doppel-nav-btn doppel-nav-prev">
            <span class="doppel-arrow">←</span>
            <div class="doppel-text">
                <span class="doppel-label">Previous</span>
                <span class="doppel-title">Philosophy</span>
            </div>
        </a>
        </div>
        <div style="flex: 1; min-width: 200px; text-align: right;">
            <a href="api-reference.md" class="doppel-nav-btn doppel-nav-next">
            <div class="doppel-text">
                <span class="doppel-label">Next</span>
                <span class="doppel-title">API Reference</span>
            </div>
            <span class="doppel-arrow">→</span>
        </a>
        </div>
    </div>
    <div style="margin-top: 1.25rem; text-align: center; color: #94a3b8; font-size: 0.8rem; letter-spacing: 0.03em;">
        <span>📚 doppel Documentation • Core Concepts</span>
    </div>
</div>

<style>
    .doppel-nav-btn {
        display: inline-flex;
        align-items: center;
        gap: 0.75rem;
        padding: 0.85rem 1.5rem;
        border-radius: 10px;
        font-weight: 500;
        text-decoration: none;
        color: #ffffff;
        position: relative;
        overflow: hidden;
        transition: transform 0.25s cubic-bezier(0.4, 0, 0.2, 1),
        box-shadow 0.25s cubic-bezier(0.4, 0, 0.2, 1),
        background 0.3s ease;
        box-shadow: 0 4px 10px rgba(0, 0, 0, 0.3);
    }
    
    /* Base Gradients */
    .doppel-nav-prev {
        background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
    }
    
    .doppel-nav-next {
        background: linear-gradient(135deg, #10b981 0%, #047857 100%);
    }
    
    /* Hover Animation */
    .doppel-nav-btn:hover {
        transform: translateY(-3px) scale(1.02);
        box-shadow: 0 12px 24px rgba(0, 0, 0, 0.4);
    }
    
    .doppel-nav-prev:hover {
        background: linear-gradient(135deg, #60a5fa 0%, #2563eb 100%);
    }
    
    .doppel-nav-next:hover {
        background: linear-gradient(135deg, #34d399 0%, #059669 100%);
    }
    
    /* Active/Click State */
    .doppel-nav-btn:active {
        transform: translateY(0) scale(0.97);
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
    }
    
    /* Focus for keyboard accessibility */
    .doppel-nav-btn:focus-visible {
        outline: 2px solid #60a5fa;
        outline-offset: 3px;
        border-radius: 12px;
    }
    
    /* Directional Arrow Slide */
    .doppel-arrow {
        font-size: 1.2rem;
        transition: transform 0.25s cubic-bezier(0.4, 0, 0.2, 1);
    }
    
    .doppel-nav-prev:hover .doppel-arrow {
        transform: translateX(-4px);
    }
    
    .doppel-nav-next:hover .doppel-arrow {
        transform: translateX(4px);
    }
    
    /* Typography */
    .doppel-text {
        display: flex;
        flex-direction: column;
        line-height: 1.25;
    }
    
    .doppel-label {
        font-size: 0.65rem;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        opacity: 0.85;
        margin-bottom: 2px;
    }
    
    .doppel-title {
        font-size: 0.95rem;
        font-weight: 600;
    }
    
    /* Mobile Responsiveness */
    @media (max-width: 768px) {
        .doppel-nav-container > div:first-child {
            flex-direction: column !important;
            gap: 1rem !important;
        }
        
        .doppel-nav-container > div:last-child {
            text-align: left !important;
        }
    }
</style>
<!-- /Navigation -->

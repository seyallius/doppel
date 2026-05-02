# 🔍 Reflection Engine (Fallback)

> When manual cloning isn't available: safe, correct, but slower. Use judiciously. ✨

---

## When Reflection is Used

The reflection engine (`engine.Engine`) is **never the default**. It is consulted only when:

1. ❌ No `Cloner[T]` is registered for the type, AND
2. ❌ The value does not implement `core.SelfClonable[T]`

```
Priority Chain:
Registered Cloner[T] → SelfClonable[T] → engine.Engine (reflection)
```

### When to Use Reflection Fallback

✅ Prototyping with third-party types you can't modify  
✅ Legacy code migration where manual `Clone()` isn't feasible yet  
✅ Dynamic scenarios where type isn't known at compile time  
⚠️ **Performance note**: Manual cloning is 3-6× faster; use reflection fallback judiciously

---

## Engine API

```go
// Engine performs reflection-based deep copying as the LAST resort.
type Engine struct { /* unexported */ }

// Options configures an Engine at construction time.
type Options struct {
    CyclePolicy CyclePolicy // zero value = PreserveShared
}

// CyclePolicy controls how Engine handles cyclic and shared pointer references.
type CyclePolicy int

const (
    PreserveShared CyclePolicy = iota // default: reproduce exact graph topology
    BreakCycles                       // break back-edges → nil, acyclic output
    ErrorOnCycle                      // return *CycleError on any back-edge
)

// CycleError is returned when CyclePolicy is ErrorOnCycle and a cycle is detected.
type CycleError struct {
    Addr     uintptr // pointer address where cycle detected
    TypeName string  // reflect.Type.String() for debugging
}

// New creates an Engine with default options (PreserveShared cycle policy).
func New(lookup TypeLookup) *Engine

// NewWithOptions creates an Engine with explicitly configured options.
func NewWithOptions(lookup TypeLookup, opts Options) *Engine

// Clone deep-copies src and returns a reflect.Value of the same type.
func (e *Engine) Clone(src reflect.Value) (reflect.Value, error)
```

---

## Cycle Policies

### `PreserveShared` (Default)

Reproduces exact graph topology. Shared pointers in the original remain shared in the clone.

```go
// Self-loop: n → n
n := &Node{Value: 42}
n.Next = n

eng := engine.New(nil) // default: PreserveShared
clonedVal, _ := eng.Clone(reflect.ValueOf(n))
cloned := clonedVal.Interface().(*Node)

// Cycle preserved: cloned.Next == cloned
```

✅ Use for: General-purpose cloning, faithful reproduction

### `BreakCycles`

Breaks back-edges by setting them to `nil`, producing an acyclic output safe for serialization.

```go
eng := engine.NewWithOptions(nil, engine.Options{
    CyclePolicy: engine.BreakCycles,
})
clonedVal, _ := eng.Clone(reflect.ValueOf(n))
cloned := clonedVal.Interface().(*Node)

// Cycle broken: cloned.Next == nil ← safe for JSON!
```

✅ Use for: JSON/YAML serialization, tree conversion, avoiding infinite loops

### `ErrorOnCycle`

Returns `*CycleError` on any cycle for strict validation during development.

```go
eng := engine.NewWithOptions(nil, engine.Options{
    CyclePolicy: engine.ErrorOnCycle,
})
_, err := eng.Clone(reflect.ValueOf(n))
// err.(*engine.CycleError) → "cycle detected at 0x... (type *GraphNode)"
```

✅ Use for: Development-time assertions, data model validation

---

## Features & Limitations

### ✅ Features

- Kind-specific cloning: `Ptr`, `Struct`, `Slice`, `Map`, `Array`, `Interface`, primitives
- Struct tag support: `doppel:"-"` (skip), `doppel:"shallow"` (share backing)
- Configurable cycle handling via `CyclePolicy`
- Shared reference preservation under `PreserveShared`
- Nil-safety: preserves `nil` vs `empty` distinction for slices/maps/pointers
- Error context: wraps failures with field-path via `core.WrapError`
- Concurrency: all mutable state is per-call; `Engine` is safe for concurrent use

### ❌ Limitations

- Unexported fields are skipped (use `SelfClonable[T]` to include them)
- `chan`, `func`, `unsafe.Pointer` are shallow-copied (reference semantics)
- Interface values cloned via concrete type; dynamic dispatch preserved

---

## Performance Considerations

### Benchmark Comparison (Indicative)

| Benchmark        | Manual (ns/op) | Reflection (ns/op) | Speedup   |
|------------------|----------------|--------------------|-----------|
| `Score`          | 22.03 ± 8%     | 131.8 ± 3%         | **~6×**   |
| `User`           | 309.9 ± 2%     | 1.193µ ± 4%        | **~4×**   |
| `Order`          | 615.9 ± 4%     | 2.183µ ± 4%        | **~3.5×** |
| `UserLargeSlice` | 8.363µ ± 3%    | 32.76µ ± 0%        | **~4×**   |
| `UserLargeMap`   | 29.26µ ± 0%    | 99.41µ ± 4%        | **~3.4×** |

### Key Takeaways

- 🚀 Manual cloning is **3-6× faster** than reflection-based cloning
- 🧠 Manual cloning uses **~40-95% fewer allocations**, especially with large maps
- ⚡ The gap grows with complexity — nested structs and large collections benefit most
- 🔁 Reflection fallback is still **correct and safe** — use when convenience outweighs performance

> 💡 **Pro Tip**: When using the reflection engine, register cloners for hot-path types via `TypeLookup`. A registry hit
> reduces reflection overhead to near-zero. (◕‿◕)✧

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->
<div class="doppel-nav-container"
     style="margin-top: 3rem; padding: 1.75rem; border-top: 1px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c); box-shadow: 0 8px 24px rgba(0,0,0,0.4);">
    <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 1rem;">
        <div style="flex: 1; min-width: 200px;">
            <a href="advanced.md" class="doppel-nav-btn doppel-nav-prev">
            <span class="doppel-arrow">←</span>
            <div class="doppel-text">
                <span class="doppel-label">Previous</span>
                <span class="doppel-title">Advanced</span>
            </div>
        </a>
        </div>
        <div style="flex: 1; min-width: 200px; text-align: right;">
            <a href="testing.md" class="doppel-nav-btn doppel-nav-next">
            <div class="doppel-text">
                <span class="doppel-label">Next</span>
                <span class="doppel-title">Testing & Benchmarks</span>
            </div>
            <span class="doppel-arrow">→</span>
        </a>
        </div>
    </div>
    <div style="margin-top: 1.25rem; text-align: center; color: #94a3b8; font-size: 0.8rem; letter-spacing: 0.03em;">
        <span>📚 doppel Documentation • Reflection Engine</span>
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

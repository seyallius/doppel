# 🗓️ Roadmap

> What's done, what's next, and where `doppel` is heading. ✨

---

## Phase Summary

| Phase | Focus                                                          | Status         |
|-------|----------------------------------------------------------------|----------------|
| **1** | Manual deep copy foundation (no reflection)                    | ✅ Complete     |
| **2** | Cloner registry — per-type override, thread-safe lookup        | ✅ Complete     |
| **3** | Field-level cloners — per-field override and conditional logic | 🔜 Next        |
| **4** | Reflection fallback — automatic clone for unregistered types   | ✅ Complete     |
| **5** | Cycle detection — configurable policies for cyclic graphs      | ✅ Complete     |
| **6** | Benchmarking suite — cross-strategy comparison                 | ✅ Complete (?) |
| **7** | API polish — `CloneWithOptions`, JSON-tag filtering, docs      | 📋 Planned     |

---

## Phase 1 — Manual Deep Copy Foundation ✅

**Goal**: Build the core cloning system using explicit/manual cloning only.

### Completed Requirements

- ✅ Go module `doppel` with clean package layout (`/core`, `/manual`)
- ✅ `Cloner[T]` interface and `NewFuncCloner` adapter
- ✅ Built-in manual helpers: `CloneSlice`, `CloneMap`, `ClonePointer` + `*Of` variants
- ✅ Zero reflection — all helpers resolved at compile time
- ✅ Nil safety contract preserved across all helpers
- ✅ Contextual error wrapping via `core.WrapError`
- ✅ Comprehensive unit tests for nested structs, maps, slices, pointers

---

## Phase 2 — Cloner Registry ✅

**Goal**: Allow users to plug in custom cloning logic per type.

### Completed Requirements

- ✅ Thread-safe `registry.Registry` with `Register[T]`, `Lookup[T]`
- ✅ `CloneWithRegistry` priority chain: Registered → SelfClonable → ErrNoCloner
- ✅ Reflection used only for type key derivation — never for field access
- ✅ Bridge interface `TypeLookup` for acyclic dependency with `engine/`

---

## Phase 3 — Field-Level Customization 🔜 (Next!)

**Goal**: Fine-grained control per struct field.

### Planned Requirements

- 🔄 `FieldCloner` function type: `func(value any) (any, error)`
- 🔄 `RegisterFieldCloner(structType, fieldName, cloner)` API
- 🔄 Priority: Field cloner → Type cloner → Manual clone → Reflection
- 🔄 Example: Clone map only if value satisfies condition
- 🔄 Reflection only for field discovery — cloning still manual/registry

---

## Phase 4 — Reflection Fallback ✅

**Goal**: Introduce reflection as a fallback only when manual clone or registry cloner is unavailable.

### Completed Requirements

- ✅ `engine.Engine` with kind-specific cloning (Struct, Slice, Map, Ptr, Array, Interface)
- ✅ Struct tag support: `doppel:"-"` (skip), `doppel:"shallow"` (share backing)
- ✅ Cycle safety via `visited` map (shared reference preservation)
- ✅ Concurrency safety via per-call `cloneState`
- ✅ Error context via `core.WrapError`
- ✅ Bridge to registry via `TypeLookup` interface

### When to Use

- Prototyping with third-party types you can't modify
- Legacy code migration where manual `Clone()` isn't feasible yet
- Dynamic scenarios where type isn't known at compile time

⚠️ **Performance note**: Manual cloning is 3-6× faster; use reflection fallback judiciously.

---

## Phase 5 — Cycle Detection ✅

**Goal**: Provide configurable cycle-handling strategies for pointer graphs.

### Completed Policies

- ✅ `PreserveShared` (default): reproduce exact topology, deduplicate shared refs
- ✅ `BreakCycles`: break back-edges → nil, produce acyclic output
- ✅ `ErrorOnCycle`: return `*CycleError` on any cycle for strict validation

### API

```go
engine.Options{CyclePolicy: engine.BreakCycles}
engine.NewWithOptions(lookup, opts)
engine.CycleError{Addr, TypeName}
```

### When to Use Each Policy

| Policy           | Use Case                                                            |
|------------------|---------------------------------------------------------------------|
| `PreserveShared` | General-purpose cloning, faithful reproduction                      |
| `BreakCycles`    | Serialization (JSON/YAML), tree conversion, avoiding infinite loops |
| `ErrorOnCycle`   | Development-time assertions, data model validation                  |

---

## Future Phases

### Phase 6 — Performance & Benchmarking 📋

- Cross-strategy benchmark suite (manual vs registry vs reflection)
- Allocation pattern optimization (map pre-sizing, slice copying)
- `benchstat` integration and `just` recipes for easy comparison

### Phase 7 — Developer Experience 📋

- `CloneWithOptions` for per-call configuration
- JSON-tag filtering for conditional field cloning
- CLI tool for generating `Clone()` stubs from struct definitions
- Enhanced godoc examples and IDE-friendly autocomplete

---

## Contributing to the Roadmap

Have an idea for Phase 3+?

1. Open an issue with your use case
2. Propose an API design following our [Philosophy](./philosophy.md)
3. Submit a PR with tests and benchmarks ✧◝(⁰▿⁰)◜✧

> 💡 **Remember**: Every feature must respect our core principle — explicit over magic. Reflection is a fallback, never
> the default. (◕‿◕)

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->
<div class="doppel-nav-container"
     style="margin-top: 3rem; padding: 1.75rem; border-top: 1px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c); box-shadow: 0 8px 24px rgba(0,0,0,0.4);">
    <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 1rem;">
        <div style="flex: 1; min-width: 200px;">
            <a href="testing.md" class="doppel-nav-btn doppel-nav-prev">
            <span class="doppel-arrow">←</span>
            <div class="doppel-text">
                <span class="doppel-label">Previous</span>
                <span class="doppel-title">Testing & Benchmarks</span>
            </div>
        </a>
        </div>
        <div style="flex: 1; min-width: 200px; text-align: right;">
            <a href="INDEX.md" class="doppel-nav-btn doppel-nav-next">
            <div class="doppel-text">
                <span class="doppel-label">Next</span>
                <span class="doppel-title">Back to Index</span>
            </div>
            <span class="doppel-arrow">→</span>
        </a>
        </div>
    </div>
    <div style="margin-top: 1.25rem; text-align: center; color: #94a3b8; font-size: 0.8rem; letter-spacing: 0.03em;">
        <span>📚 doppel Documentation • Roadmap</span>
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

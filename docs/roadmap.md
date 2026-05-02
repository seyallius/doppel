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

<!--

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Roadmap
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap;">
        <div style="flex: 1; min-width: 200px;">
            <a href="testing.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
            <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
            <span style="display: flex; flex-direction: column; line-height: 1.3;">
          <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
          <span style="font-size: 1rem; font-weight: 600;">Testing & Benchmarks</span>
        </span>
        </a>
        </div>
        <div style="flex: 1; min-width: 200px;">
            <a href="INDEX.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
        <span style="display: flex; flex-direction: column; line-height: 1.3;">
          <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
          <span style="font-size: 1rem; font-weight: 600;">Back to Index</span>
        </span>
            <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
        </a>
        </div>
    </div>
</div>
<!-- /Navigation -->


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

<!-- Navigation -->
<div style="margin-top: 3rem; padding: 1.5rem; border-top: 2px solid #e1e4e8; border-radius: 6px; background: #f6f8fa;">
  <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 1rem;">
    <div style="flex: 1; min-width: 200px;">
      <a href="./testing.md" style="display: inline-flex; align-items: center; padding: 0.75rem 1rem; background: #0366d6; color: white; text-decoration: none; border-radius: 6px; font-weight: 600; transition: background 0.2s;">
        <span style="margin-right: 0.5rem;">←</span>
        <div style="display: flex; flex-direction: column;">
          <span style="font-size: 0.75rem; opacity: 0.9;">Previous:</span>
          <span>Testing & Benchmarks</span>
        </div>
      </a>
    </div>
    <div style="flex: 1; min-width: 200px; text-align: right;">
      <a href="./INDEX.md" style="display: inline-flex; align-items: center; padding: 0.75rem 1rem; background: #28a745; color: white; text-decoration: none; border-radius: 6px; font-weight: 600; transition: background 0.2s;">
        <div style="display: flex; flex-direction: column; margin-right: 0.5rem;">
          <span style="font-size: 0.75rem; opacity: 0.9;">Back to</span>
          <span>INDEX.md →</span>
        </div>
      </a>
    </div>
  </div>
  <div style="margin-top: 1rem; text-align: center; color: #586069; font-size: 0.875rem;">
    <span>📚 doppel Documentation • Roadmap</span>
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

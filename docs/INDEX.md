# doppel

**Your data's doppelgänger — deep copies without side effects.**

doppel is a Go library for safe, explicit deep cloning of complex data structures. It provides a layered architecture
that prioritizes manual, zero-reflection cloning by default, with an optional reflection fallback for types you don't
control — all fully composable and extensible through a type-safe registry.

---

## Why doppel?

Go assignment is a shallow copy. Structs with pointer fields, slices, and maps silently share memory between "originals"
and "copies." This leads to subtle, hard-to-track bugs when code mutates what it believes is an independent copy. doppel
solves this by providing explicit, composable deep-copy primitives that give you full control over every field, every
allocation, and every edge case.

### Key principles

| Principle                  | What it means                                                                                                       |
|----------------------------|---------------------------------------------------------------------------------------------------------------------|
| **Manual first**           | No reflection, no magic, maximum speed. You write the clone logic using type-safe generics.                         |
| **Reflection as fallback** | The engine is consulted only when no manual clone exists — never the default, always the last resort.               |
| **Composable**             | `CloneSlice`, `CloneMap`, `ClonePointer` are generic helpers you wire together inside your type's `Clone()` method. |
| **Extensible**             | Register per-type or per-field cloners to override the default behavior without modifying the original type.        |
| **Explicit**               | Every clone path is visible and auditable — no hidden behavior, no surprises.                                       |

---

## Architecture overview

doppel is built in five phases, each adding capability on top of the previous:

```
Phase 1: Manual Deep Copy Foundation
  ├── core.Cloner[T]            — external clone logic interface
  ├── core.SelfClonable[T]      — self-cloning type interface
  ├── core.FuncCloner[T]        — function-to-Cloner adapter
  ├── manual.CloneSlice[T]      — generic slice deep copy
  ├── manual.CloneMap[K,V]      — generic map deep copy
  ├── manual.ClonePointer[T]    — generic pointer deep copy
  └── manual.Identity[T]        — no-op helper for primitives

Phase 2: Cloner Registry
  └── registry.Registry         — thread-safe, type-keyed cloner store

Phase 3: Field-Level Customization
  └── registry.RegisterField[T,F]  — per-field cloner overrides

Phase 4: Reflection Engine
  └── engine.Engine             — reflection-based deep copy fallback

Phase 5: Cycle & Sharing Policy
  └── engine.CyclePolicy        — PreserveShared / BreakCycles / ErrorOnCycle
```

### Clone dispatch chain

When you call `doppel.CloneDeep`, the library walks this priority chain, stopping at the first strategy that applies:

```
Registered Cloner[T]  →  SelfClonable[T]  →  Field Cloner  →  Reflection Engine
     (fastest)              (type-owned)      (per-field)      (automatic)
```

---

## Package map

| Package    | Role                                                                                  |
|------------|---------------------------------------------------------------------------------------|
| `doppel`   | Top-level API: `Clone`, `CloneWith`, `CloneWithRegistry`, `CloneDeep`                 |
| `core`     | Foundational interfaces: `Cloner[T]`, `SelfClonable[T]`, `FuncCloner[T]`, error types |
| `manual`   | Generic helpers: `CloneSlice`, `CloneMap`, `ClonePointer`, `Identity`                 |
| `registry` | Thread-safe type-keyed and field-keyed cloner store                                   |
| `engine`   | Reflection-based deep copy engine with cycle detection                                |

## Navigation flow

```
INDEX 
→ Getting Started 
→ SelfClonable 
→ Manual Helpers 
→ Registry 
→ Field Cloners
→ Reflection Engine 
→ Struct Tags 
→ Cycle Policy 
→ Error Handling 
→ Patterns
→ Benchmarks 
→ API Reference
```

---

## Documentation roadmap

The documentation is designed to be read in order. Each page builds on concepts introduced in the previous one.

1. **[Getting Started](getting-started.md)** — Install doppel, write your first clone in under 2 minutes.
2. **[SelfClonable Interface](self-clonable.md)** — The primary interface for types that own their clone logic.
3. **[Manual Helpers](manual-helpers.md)** — Generic slice, map, and pointer cloners you compose inside `Clone()`.
4. **[Cloner Registry](registry.md)** — Register cloners for types you don't control or need context for.
5. **[Field-Level Cloners](field-cloners.md)** — Override clone behavior for individual struct fields.
6. **[Reflection Engine](reflection-engine.md)** — How the automatic fallback works and what it supports.
7. **[Struct Tags](struct-tags.md)** — Control cloning per-field with `doppel:"..."` struct tags.
8. **[Cycle & Sharing Policy](cycle-policy.md)** — Handle cyclic and shared pointer graphs.
9. **[Error Handling](error-handling.md)** — Understand and work with doppel's error types.
10. **[Patterns & Best Practices](patterns.md)** — Proven patterns for real-world usage.
11. **[Benchmarks](benchmarks.md)** — Performance data comparing manual, registry, and reflection paths.
12. **[API Reference](api-reference.md)** — Complete function and type signatures.

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Documentation
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"></div>
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: center; align-items: center;">
            <a href="INDEX.md" style="display: flex; align-items: center; justify-content: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #8b5cf6, #6d28d9); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(139, 92, 246, 0.3); text-align: center;">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">⌂</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Return to</span>
                    <span style="font-size: 1rem; font-weight: 600;">Index</span>
                </span>
            </a>
        </div>
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"><a href="getting-started.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Getting Started</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
            </a></div>
    </div>
</div>
<!-- /Navigation -->


# 🏠 doppel Documentation Index

> Table of contents for the doppel Go library documentation.

**doppel** — Safe, explicit deep cloning of Go data structures.

*"Your data's doppelgänger — deep copies without side effects."*

---

## 📄 Documentation Files

| File                                                             | Description                                                             |
|------------------------------------------------------------------|-------------------------------------------------------------------------|
| [🗺️ Roadmap](../../../../Downloads/Documents/roadmap.md)        | Phased development plan — Phase 1 through Phase 5                       |
| [📖 Usage Guide](../../../../Downloads/Documents/usage-guide.md) | Step-by-step guide from simple to advanced cloning                      |
| [📐 API Reference](api-reference.md)                             | Complete reference for every public symbol                              |
| [⚡ Advanced Topics](advanced.md)                                 | Deep dives into priority chains, engine internals, and design decisions |

---

## 📑 Topic Index

### Getting Started

- [Implement SelfClonable](../../../../Downloads/Documents/usage-guide.md#step-1--implement-selfclonable)
- [Use doppel.Clone](../../../../Downloads/Documents/usage-guide.md#step-2--use-doppeclone)
- [Use Manual Helpers](../../../../Downloads/Documents/usage-guide.md#step-3--use-manual-helpers)
- [Use doppel.CloneWith](../../../../Downloads/Documents/usage-guide.md#step-4--use-doppelclonewith)

### Registry (Phase 2)

- [Create a Registry](../../../../Downloads/Documents/usage-guide.md#step-5--create-a-registry)
- [Register Type Cloners](../../../../Downloads/Documents/usage-guide.md#step-6--register-type-cloners)
- [Use doppel.CloneWithRegistry](../../../../Downloads/Documents/usage-guide.md#step-7--use-doppelclonewithregistry)
- [Type-Level Cloner Registration API](api-reference.md#type-level-cloner-registration-registry-package)

### Field-Level Cloning (Phase 3) 🆕

- [Field-Level Cloning with CloneDeep](../../../../Downloads/Documents/usage-guide.md#step-9--field-level-cloning-with-clonedeep) —
  The core Phase 3 workflow
- [Field-Level Cloner Registration API](api-reference.md#field-level-cloner-registration-registry-package) —
  RegisterField, LookupField, HasField, DeregisterField, FieldLen, LookupAnyField
- [doppel.CloneDeep](api-reference.md#doppelclonedeep) — Full priority chain: Type Cloner → SelfClonable → Engine
- [doppel.MustCloneDeep](api-reference.md#doppelmustclonedeep) — Panic-on-error variant
- [Field-Level Cloners (Advanced)](advanced.md#field-level-cloners) — Auto-discovery, priority, same-field-name
  independence
- [Struct Tag Directives](advanced.md#struct-tag-directives) — `doppel:"clone"`, `doppel:"deep"`, `doppel:"readonly"`
  and more
- [Per-Field Priority Chain](advanced.md#priority-chains) — How the engine resolves each field
- [Phase 3 Roadmap Entry](../../../../Downloads/Documents/roadmap.md) — What shipped in Phase 3

### Advanced

- [Priority Chains](advanced.md#priority-chains) — Top-level and per-field priority
- [Registry Internals](advanced.md#registry-internals) — Type key derivation, field key derivation, reflect-level
  bridges
- [Engine Architecture](advanced.md#engine-architecture) — Kind dispatch, stateless design
- [Concurrency Model](advanced.md#concurrency-model) — Thread safety for Registry, Engine, and doppel functions
- [Error Handling Patterns](advanced.md#error-handling-patterns) — Error-return vs. MustClone variants

### Core Interfaces

- [core.Cloner[T]](api-reference.md#corecloner) — The extension interface
- [core.SelfClonable[T]](api-reference.md#coreselfclonable) — Self-cloning types
- [core.NewFuncCloner](api-reference.md#corenewfunccloner) — Function-to-Cloner adapter
- [core.ErrNoCloner](api-reference.md#corerrnocloner) — "no strategy available" sentinel

---

## 🏷️ API Quick Reference

### doppel Package

| Function            | Signature                        | Phase |
|---------------------|----------------------------------|-------|
| `Clone`             | `(SelfClonable[T]) → (T, error)` | 1     |
| `MustClone`         | `(SelfClonable[T]) → T`          | 1     |
| `CloneWith`         | `(T, Cloner[T]) → (T, error)`    | 1     |
| `MustCloneWith`     | `(T, Cloner[T]) → T`             | 1     |
| `CloneWithRegistry` | `(T, *Registry) → (T, error)`    | 2     |
| `CloneDeep`         | `(T, *Registry) → (T, error)`    | 3 🆕  |
| `MustCloneDeep`     | `(T, *Registry) → T`             | 3 🆕  |

### registry Package — Type-Level

| Function     | Signature                         | Phase |
|--------------|-----------------------------------|-------|
| `New`        | `() → *Registry`                  | 2     |
| `Register`   | `(*Registry, Cloner[T])`          | 2     |
| `Lookup`     | `(*Registry) → (Cloner[T], bool)` | 2     |
| `Deregister` | `(*Registry)`                     | 2     |
| `Has`        | `(*Registry) → bool`              | 2     |
| `Len`        | `(*Registry) → int`               | 2     |
| `LookupAny`  | `(reflect.Type) → (func, bool)`   | 2     |

### registry Package — Field-Level 🆕

| Function          | Signature                                 | Phase |
|-------------------|-------------------------------------------|-------|
| `RegisterField`   | `(*Registry, string, Cloner[F])`          | 3     |
| `LookupField`     | `(*Registry, string) → (Cloner[F], bool)` | 3     |
| `HasField`        | `(*Registry, string) → bool`              | 3     |
| `DeregisterField` | `(*Registry, string) → bool`              | 3     |
| `FieldLen`        | `(*Registry) → int`                       | 3     |
| `LookupAnyField`  | `(reflect.Type, string) → (func, bool)`   | 3     |

### Struct Tags 🆕

| Tag                 | Behavior                               | Phase |
|---------------------|----------------------------------------|-------|
| `doppel:"-"`        | Skip field (zero value in clone)       | 4     |
| `doppel:"shallow"`  | Shallow copy (share reference)         | 4     |
| `doppel:"readonly"` | Shallow copy (immutability intent)     | 3 🆕  |
| `doppel:"clone"`    | Require field Cloner; error if missing | 3 🆕  |
| `doppel:"deep"`     | Explicit deep copy (same as default)   | 3 🆕  |

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Documentation Index
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="../README.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">README.md</span>
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


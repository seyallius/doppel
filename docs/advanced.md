# ⚡ doppel Advanced Topics

> Deep dives into the internals, priority chains, and design decisions behind doppel.

---

## Table of Contents

- [Priority Chains](#priority-chains)
- [Field-Level Cloners](#field-level-cloners)
- [Struct Tag Directives](#struct-tag-directives)
- [Registry Internals](#registry-internals)
- [Engine Architecture](#engine-architecture)
- [Concurrency Model](#concurrency-model)
- [Error Handling Patterns](#error-handling-patterns)

---

## Priority Chains

doppel uses a layered priority chain at multiple levels. Understanding these chains is key to
predicting exactly what happens when you call any of the public entry points.

### Top-Level Priority (per type)

When you call `CloneDeep[T](src, reg)`, the top-level dispatch works like this:

| Priority | Strategy | When it applies |
|---|---|---|
| 1 | Registered `Cloner[T]` | `reg` contains a Cloner for the exact type T |
| 2 | `SelfClonable[T]` | T implements `Clone() (T, error)` |
| 3 | Reflection engine | Fallback — recurse into the value graph |

This means a registered type Cloner **always wins** over SelfClonable and the engine. If you
register a `Cloner[User]`, it will be used for every `User` value, regardless of whether `User`
implements `SelfClonable`. This is by design: registration is an explicit override, and explicit
beats implicit.

### Per-Field Priority (inside the engine)

When the reflection engine clones a struct, each field is resolved through a five-level chain:

| Priority | Strategy | What happens |
|---|---|---|
| 1 | Struct tag directive | `doppel:"-"`, `doppel:"shallow"`, `doppel:"readonly"`, `doppel:"clone"` |
| 2 | Registered field Cloner | Auto-discovered by (structType, fieldName) pair |
| 3 | Registered type Cloner | For the field's type — e.g., `Cloner[*Address]` |
| 4 | SelfClonable on field value | Field value implements `Clone() (T, error)` |
| 5 | Reflection fallback | Recurse into the field value |

Struct tags are the highest priority because they represent explicit, compile-time declarations
about how a field should be treated. Field cloners come next because they're registered at runtime
but are more specific than type cloners (which apply to all occurrences of a type, not just one
field of one struct). SelfClonable and reflection are the fallbacks, consistent with the top-level
chain.

---

## Field-Level Cloners

🆕 **Phase 3** — Field-level cloners are the core innovation of Phase 3. They solve a specific
problem: you have a large struct (think 50–200+ fields) where only a handful need custom clone
logic, but you're forced to write a monolithic `Clone()` method that handles every single field.

### The Core Idea

Instead of writing this:

```go
func (m *BigModel) Clone() (*BigModel, error) {
    // 200 lines of field-by-field copying...
    return &BigModel{
        ID: m.ID,
        Name: m.Name,
        // ... 198 more fields ...
        Address: addr,   // custom logic
        Tags: tags,      // custom logic
    }, nil
}
```

You write this:

```go
reg := registry.New()

// Only the fields that need custom logic
registry.RegisterField[BigModel, *Address](reg, "Address", core.NewFuncCloner(cloneAddr))
registry.RegisterField[BigModel, []string](reg, "Tags", core.NewFuncCloner(
    func(src []string) ([]string, error) { return append([]string{}, src...), nil },
))

// Clone — reflection handles the other 198 fields
cloned, err := doppel.CloneDeep(bigModel, reg)
```

The reflection engine deep-copies every field automatically, but when it encounters a field with
a registered cloner, it uses that cloner instead of the default reflection path. This is the
**"default deep copy + selective override"** workflow.

### Auto-Discovery

Field cloners are **auto-discovered** by the engine during struct cloning. You don't need to
annotate your struct fields with tags for auto-discovery to work — just register the cloner and
the engine will find it. Here's the exact mechanism:

1. The engine's `cloneStruct` method iterates over each exported field of the struct.
2. Before falling through to the default reflection path, it checks whether the `FieldLookup`
   provider (typically `*registry.Registry`) has a cloner registered for the current
   `(structType, fieldName)` pair.
3. If a field cloner is found, it's invoked with the field's value. The result is used as the
   cloned field value.
4. If no field cloner is found, the engine proceeds to the type cloner → SelfClonable → reflection
   fallback chain for that field.

This means you can add or remove field cloners at any time without changing your struct definitions.
The engine adapts dynamically to whatever cloners are registered in the provided registry.

### Same Field Name, Different Struct Types

Field cloners are keyed by `(structType, fieldName)` pairs, not just field names. This means
two different struct types can have a field named `"Nested"` with completely independent cloners:

```go
registry.RegisterField[User, *Address](reg, "Nested", core.NewFuncCloner(cloneUserAddr))
registry.RegisterField[Order, *Address](reg, "Nested", core.NewFuncCloner(cloneOrderAddr))
```

The engine will use `cloneUserAddr` for `User.Nested` and `cloneOrderAddr` for `Order.Nested`
without any ambiguity.

### Field Cloner vs. Type Cloner Priority

When both a field cloner and a type cloner exist for the same field, the **field cloner wins**.
This is because field cloners are more specific — they target a particular field of a particular
struct, while type cloners apply to all occurrences of a type. Consider this scenario:

```go
// Type cloner for *Address — applies everywhere
registry.Register(reg, core.NewFuncCloner(func(src *Address) (*Address, error) {
    return &Address{Street: src.Street + "_type"}, nil
}))

// Field cloner for User.HomeAddr — applies only to this field
registry.RegisterField[User, *Address](reg, "HomeAddr", core.NewFuncCloner(func(src *Address) (*Address, error) {
    return &Address{Street: src.Street + "_field"}, nil
}))
```

When cloning a `User`, the `HomeAddr` field will use the field cloner (`"_field"` suffix). But
when cloning an `Order` that also has an `*Address` field (without a field cloner registered for
it), the type cloner is used (`"_type"` suffix).

### Deregistering Field Cloners

You can remove a field cloner at runtime with `DeregisterField`. After deregistration, the engine
falls back to the default reflection path for that field:

```go
registry.RegisterField[User, *Address](reg, "HomeAddr", core.NewFuncCloner(cloneAddr))

// Later — stop using the field cloner
removed := registry.DeregisterField[User](reg, "HomeAddr")
// removed == true, engine now uses default reflection for User.HomeAddr
```

`DeregisterField` returns a boolean indicating whether a cloner was actually removed. Calling it
on a field that has no registration returns `false` and is a safe no-op.

---

## Struct Tag Directives

The engine supports `doppel` struct tags that control how individual fields are treated during
cloning. Tags are the highest priority in the per-field chain — they're checked before any
cloner or fallback.

### Tag Reference

| Tag | Behavior | Use case |
|---|---|---|
| `doppel:"-"` | Skip the field entirely; clone receives the zero value | Sensitive data, computed fields, internal state |
| `doppel:"shallow"` | Assign without recursing; clone shares the field's value | Large immutable buffers, sync.Mutex, intentionally shared data |
| `doppel:"readonly"` | Same as shallow; communicates immutability intent | Conceptually immutable fields (config maps, constant slices) |
| `doppel:"clone"` | Require a registered field Cloner; errors if none found | Fields that must have custom clone logic (fail-safe) |
| `doppel:"deep"` | Explicit deep copy (same as default) | Documentation — marks fields as intentionally deep-copied |

### doppel:"clone" — The Fail-Safe Tag

The `doppel:"clone"` tag is particularly powerful. It makes the dependency on a field cloner
**explicit**: if no field cloner is registered for the tagged field, the clone operation returns
an error immediately. This catches configuration mistakes at runtime instead of silently falling
back to reflection.

```go
type Config struct {
    Name   string      `doppel:"deep"`     // explicit deep copy
    Secret string      `doppel:"-"`        // skip — zero value in clone
    Shared *Ref        `doppel:"shallow"`  // share the pointer
    Locked *Sensitive  `doppel:"clone"`    // MUST have a field cloner
    Env    map[string]string `doppel:"readonly"` // shallow, signals immutability
}
```

If you call `CloneDeep` on a `Config` without registering a field cloner for `Locked`, you get:

```
doppel: field Config.Locked tagged doppel:"clone" but no field cloner is registered
```

This is a **design choice**, not a limitation: in large codebases, silently falling back to
reflection for a field that was supposed to have custom logic is a bug waiting to happen. The
`clone` tag makes that dependency visible and enforceable.

### doppel:"deep" — Documentation Tag

The `doppel:"deep"` tag doesn't change behavior — it's identical to the default (no tag). Its
purpose is documentation: it makes it explicit that a field is intended to be deep-copied. This
is useful in code reviews and for future maintainers who might wonder why a particular field
doesn't have a tag.

### doppel:"readonly" — Semantic Shallow

The `doppel:"readonly"` tag is semantically identical to `doppel:"shallow"` — both result in a
simple assignment without recursion. The difference is communicative: `readonly` tells readers
that the field is conceptually immutable and safe to share between the original and the clone.
This is a convention, not an enforcement — the engine doesn't check immutability. But it makes
the intent clear, which is valuable in team settings.

---

## Registry Internals

The registry uses two separate maps under the hood:

| Map | Key | Value | Purpose |
|---|---|---|---|
| `typeCloners` | `reflect.Type` | `core.Cloner[T]` | Type-level cloners (Phase 2) |
| `fieldCloners` | `fieldKey{structType, fieldName}` | `core.Cloner[F]` | Field-level cloners (Phase 3) |

Both maps are protected by a single `sync.RWMutex`, ensuring thread safety for all operations.
The `reflect.Type` keys are derived using `reflect.TypeOf((*T)(nil)).Elem()`, which works correctly
even for interface types (where `reflect.TypeOf` on a nil interface value would return nil).

### Type Key Derivation

The `typeKeyFor[T]()` function creates a stable map key for type T:

```go
typedNilPtr := (*T)(nil)                      // Typed nil pointer — preserves type info
ptrReflectType := reflect.TypeOf(typedNilPtr) // Get *T reflect.Type
typeOfT := ptrReflectType.Elem()              // Dereference to T
```

This technique handles the edge case where T is an interface type. A plain `reflect.TypeOf(zero)`
on a nil interface value returns nil, which would break the map lookup. The typed nil pointer
approach always produces a valid `reflect.Type`.

### Field Key Derivation

For field cloners, the key is `(structType, fieldName)` where `structType` is the reflect.Type of
the enclosing struct (value type, not pointer). The `structTypeFor[T]()` helper dereferences any
pointer type to reach the underlying struct:

```go
// Both resolve to the same key:
registry.RegisterField[User, *Address](reg, "HomeAddr", cloner)
registry.RegisterField[*User, *Address](reg, "HomeAddr", cloner) // same key
```

### Reflect-Level Bridges

The `LookupAny` and `LookupAnyField` methods are the bridge between the type-safe generic registry
and the reflection engine. They wrap the stored `Cloner[T]` as a `func(reflect.Value) (reflect.Value, error)`
by calling the `Clone` method through reflection. This allows the engine to invoke registered
cloners without knowing T at compile time.

---

## Engine Architecture

The engine operates at `reflect.Value` level and uses a two-tier structure:

- **`Engine`** — stateless, safe for concurrent use, holds the `TypeLookup` / `FieldLookup`
  providers and options
- **`cloneState`** — per-call mutable state (visited map, DFS stack), never shared across
  goroutines

This design means a single `Engine` instance can serve unlimited concurrent clone operations.
Each call to `Engine.Clone` creates a fresh `cloneState`, ensuring isolation.

### Kind Dispatch

The engine dispatches based on `reflect.Kind`:

| Kind | Handler | Behavior |
|---|---|---|
| Ptr | `clonePointer` | Deep copy, cycle-aware |
| Struct | `cloneStruct` | Field-by-field with tag + field cloner support |
| Slice | `cloneSlice` | Element-by-element, preserves nil-vs-empty |
| Map | `cloneMap` | Key-value, preserves nil-vs-empty |
| Array | `cloneArray` | Fixed-length, element-by-element |
| Interface | `cloneInterface` | Unwrap concrete value, recurse |
| Primitives | inline | Assignment (already a complete deep copy) |
| Chan, Func, UnsafePointer | inline | Shallow copy (cannot be meaningfully deep-copied) |

---

## Concurrency Model

All public APIs in doppel are safe for concurrent use:

- **Registry** — all reads and writes are protected by `sync.RWMutex`
- **Engine** — stateless; per-call `cloneState` is never shared
- **doppel functions** — stateless; delegate to registry and engine

The recommended pattern is to create a single `Registry` at startup, register all cloners during
initialization, and then share it across goroutines for the lifetime of the application:

```go
var globalReg = registry.New()

func init() {
    registry.Register(globalReg, core.NewFuncCloner(cloneUser))
    registry.RegisterField[BigModel, *Address](globalReg, "Address", core.NewFuncCloner(cloneAddr))
}

func HandleRequest(src BigModel) (BigModel, error) {
    return doppel.CloneDeep(src, globalReg)
}
```

---

## Error Handling Patterns

doppel provides two error handling strategies:

### Error-Return (production)

```go
cloned, err := doppel.CloneDeep(src, reg)
if err != nil {
    // Handle gracefully — log, wrap, return to caller
    return BigModel{}, fmt.Errorf("clone failed: %w", err)
}
```

### Panic-on-Error (tests / init)

```go
cloned := doppel.MustCloneDeep(src, reg) // panics on error
```

The `Must*` variants follow the `template.Must` convention from the standard library. They're
appropriate when:

- You're in a test and a clone failure means the test setup is broken
- You're in `init()` and a clone failure means the program can't start
- You're building a value that's used as a constant throughout the program

**Don't use `Must*` in request handlers or any path where errors are recoverable.**

---

<!-- Navigation -->
<div align="center">

[🏠 INDEX](INDEX.md) · [🗺️ Roadmap](../../../../Downloads/Documents/roadmap.md) · [📖 Usage Guide](../../../../Downloads/Documents/usage-guide.md) · [📐 API Reference](api-reference.md)

</div>

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Advanced Topics
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="usage-guide.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Usage Guide</span>
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
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"><a href="reflection-engine.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Reflection Engine</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
            </a></div>
    </div>
</div>
<!-- /Navigation -->


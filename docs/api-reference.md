# 📐 doppel API Reference

> Complete reference for every public symbol in the doppel, registry, engine, and core packages.

---

## Table of Contents

- [Public Entry Points (doppel package)](#public-entry-points-doppel-package)
  - [doppel.Clone](#doppeclone)
  - [doppel.MustClone](#doppelmustclone)
  - [doppel.CloneWith](#doppelclonewith)
  - [doppel.MustCloneWith](#doppelmustclonewith)
  - [doppel.CloneWithRegistry](#doppelclonewithregistry)
  - [doppel.CloneDeep](#doppelclonedeep)
  - [doppel.MustCloneDeep](#doppelmustclonedeep)
- [Type-Level Cloner Registration (registry package)](#type-level-cloner-registration-registry-package)
- [Field-Level Cloner Registration (registry package)](#field-level-cloner-registration-registry-package)
- [Engine Configuration (engine package)](#engine-configuration-engine-package)
- [Core Interfaces (core package)](#core-interfaces-core-package)

---

## Public Entry Points (doppel package)

These are the functions you'll call most often. They live in the top-level `doppel` package and
dispatch to the appropriate strategy (SelfClonable, registry, or reflection engine).

### doppel.Clone

```go
// Clone produces a deep copy of src by calling src.Clone().
// The compiler enforces that src implements core.SelfClonable[T].
// Returns (T, error) where error includes contextual field-path on failure.
func Clone[T any](src core.SelfClonable[T]) (T, error)
```

Produces a deep copy of `src` by calling `src.Clone()`. The compiler enforces that `src` satisfies
`core.SelfClonable[T]`, so you get a type-safety guarantee at compile time. For types that do **not**
implement `SelfClonable`, use `CloneWith` with an external `Cloner[T]`, or `CloneDeep` for the full
reflection fallback chain.

**Example:**

```go
cloned, err := doppel.Clone(user) // user must implement SelfClonable[User]
```

### doppel.MustClone

```go
// MustClone is like Clone, but panics on error instead of returning it.
// Intended for tests and program initialization where clone failure is always a bug.
func MustClone[T any](src core.SelfClonable[T]) T
```

Like `Clone` but panics instead of returning an error. Intended for use in tests and program
initialization, where a cloning failure is always a programming error rather than a recoverable
condition. Use this when you're confident the clone cannot fail and want to avoid error-checking
boilerplate.

**Example:**

```go
cloned := doppel.MustClone(user) // panics on error
```

### doppel.CloneWith

```go
// CloneWith produces a deep copy of src using an external Cloner[T].
// Use when src does not implement SelfClonable, or when you need
// a different clone strategy at a specific call site.
func CloneWith[T any](src T, cloner core.Cloner[T]) (T, error)
```

Produces a deep copy of `src` using the provided external `Cloner`. Use this when the source type
does not implement `SelfClonable` — for example, when cloning logic lives in a separate struct with
injected dependencies, or when you want to override the default clone for a specific call site.

**Example:**

```go
cloner := core.NewFuncCloner(func(u User) (User, error) {
    return User{ID: u.ID, Name: u.Name}, nil
})
cloned, err := doppel.CloneWith(original, cloner)
```

### doppel.MustCloneWith

```go
// MustCloneWith is like CloneWith, but panics on error.
func MustCloneWith[T any](src T, cloner core.Cloner[T]) T
```

Like `CloneWith` but panics instead of returning an error. Same use case as `MustClone` — ideal for
tests and initialization code where failure is a bug.

### doppel.CloneWithRegistry

```go
// CloneWithRegistry produces a deep copy of src by walking a priority chain:
// 1. Registered Cloner[T] in reg (fastest)
// 2. core.SelfClonable[T] fallback (if T implements it)
// 3. core.ErrNoCloner if neither is available
// Reflection is used only for type key derivation — never for field access.
func CloneWithRegistry[T any](src T, reg *registry.Registry) (T, error)
```

Produces a deep copy of `src` by walking the following lookup chain, stopping at the first strategy
that applies:

1. **Registered `Cloner[T]`** — if `reg` contains a Cloner for type T, it is used. This is the
   fastest path and the whole point of the registry.
2. **`core.SelfClonable[T]` fallback** — if T implements `SelfClonable[T]`, its `Clone()` method is
   called. All existing SelfClonable types work out of the box, even without registration.
3. **`core.ErrNoCloner`** — returned when neither strategy is available. This is an explicit signal
   to either register a Cloner, implement SelfClonable, or use `CloneDeep` for automatic reflection
   fallback.

Reflection is used only inside the registry for type key derivation — **never** for field access
or value traversal.

**Example:**

```go
reg := registry.New()
registry.Register(reg, core.NewFuncCloner(func(u User) (User, error) {
    return User{ID: u.ID, Name: u.Name + "_cloned"}, nil
}))
cloned, err := doppel.CloneWithRegistry(user, reg)
```

### doppel.CloneDeep

```go
func CloneDeep[T any](src T, reg *registry.Registry) (T, error)
```

Produces a deep copy of `src` by walking the **full priority chain**:

1. **Registered `Cloner[T]`** — if `reg` contains a Cloner for type T, it is used. This is the
   fastest path and gives you full control over the clone logic for the entire type.
2. **`core.SelfClonable[T]`** — if T implements `SelfClonable[T]`, its `Clone()` method is called.
   All existing SelfClonable types work out of the box.
3. **Reflection engine** — the engine recursively clones the value, consulting registered
   field-level cloners (via `registry.RegisterField`) before falling through to default reflection
   for each struct field.

`CloneDeep` is the entry point for the **"default deep copy + selective override"** workflow
introduced in Phase 3. It is the recommended API when you have a struct with many fields but only
need custom clone logic for a few. Pass `nil` for `reg` to use pure reflection without any
registered cloners. The engine respects `doppel` struct tags (see the engine package documentation
for details).

**Example:**

```go
reg := registry.New()
registry.RegisterField[BigStruct, *Address](reg, "HomeAddr", core.NewFuncCloner(cloneAddr))
registry.RegisterField[BigStruct, []string](reg, "Tags", core.NewFuncCloner(
    func(src []string) ([]string, error) { return append([]string{}, src...), nil },
))
cloned, err := doppel.CloneDeep(bigStruct, reg)
```

### doppel.MustCloneDeep

```go
func MustCloneDeep[T any](src T, reg *registry.Registry) T
```

Like `CloneDeep` but panics instead of returning an error. Intended for use in tests and program
initialization, where a cloning failure is always a programming error rather than a recoverable
condition. This is the panic-on-error counterpart to `CloneDeep`, following the same convention as
`MustClone` and `MustCloneWith`.

**Example:**

```go
// Safe in tests — panics immediately if anything goes wrong
cloned := doppel.MustCloneDeep(bigStruct, reg)
```

---

## Type-Level Cloner Registration (registry package)

The `registry` package provides a thread-safe, type-keyed store of `Cloner[T]` values. Reflection
is used exclusively for type key derivation and the reflect-level bridge (`LookupAny`) — never for
cloning itself.

### registry.New

```go
func New() *Registry
```

Creates and returns an empty, ready-to-use Registry. The returned Registry is safe for concurrent
use immediately. You typically create one at startup and share it across goroutines.

### registry.Register

```go
func Register[T any](r *Registry, cloner core.Cloner[T])
```

Stores `cloner` as the `Cloner[T]` for type T. If a cloner is already registered for T it is
silently replaced, making `Register` safe to call multiple times during initialization. Safe for
concurrent use.

### registry.Lookup

```go
func Lookup[T any](r *Registry) (core.Cloner[T], bool)
```

Retrieves the registered `Cloner[T]` for type T. Returns `(cloner, true)` when found, `(nil, false)`
when not registered. Safe for concurrent use.

### registry.Deregister

```go
func Deregister[T any](r *Registry)
```

Removes the registered `Cloner` for type T, if any. Calling `Deregister` on a type that has no
registration is a no-op. Safe for concurrent use.

### registry.Has

```go
func Has[T any](r *Registry) bool
```

Reports whether a `Cloner` is registered for type T. Safe for concurrent use.

### registry.Len

```go
func (r *Registry) Len() int
```

Returns the total number of registered type-level cloners. Safe for concurrent use.

### registry.LookupAny

```go
func (r *Registry) LookupAny(t reflect.Type) (func(reflect.Value) (reflect.Value, error), bool)
```

Returns a reflect-level clone function for the given `reflect.Type`. This is the bridge between
the type-safe generic registry and the reflection engine in `engine/`, which operates at
`reflect.Value` level without knowing T at compile time. Returns `(nil, false)` when no Cloner is
registered for `t`. Safe for concurrent use.

---

## Field-Level Cloner Registration (registry package)

🆕 **Phase 3** — Field-level cloners provide fine-grained control over individual struct fields.
When the reflection engine clones a struct, it checks for registered field cloners before falling
through to default reflection-based cloning for each field. This enables a "default deep copy +
selective override" workflow where most fields are cloned automatically by reflection, but specific
fields use custom clone logic registered via `RegisterField`.

### registry.RegisterField

```go
func RegisterField[T any, F any](r *Registry, fieldName string, cloner core.Cloner[F])
```

Registers a `Cloner[F]` for a specific field of struct type T. The cloner is invoked when the
reflection engine encounters this field during a deep copy operation, overriding the default
reflection-based cloning. T must be a struct type (or pointer to a struct), and `fieldName` must
name an exported field of T whose type is compatible with F.

If a field cloner is already registered for the same struct type and field name, it is silently
replaced (consistent with `Register` behavior). Panics if T is not a struct type, if `fieldName`
doesn't exist on T, or if the field is unexported. Safe for concurrent use.

**Example:**

```go
registry.RegisterField[User, *Address](reg, "HomeAddr", core.NewFuncCloner(cloneAddr))
registry.RegisterField[User, []string](reg, "Tags", core.NewFuncCloner(
    func(src []string) ([]string, error) { return append([]string{}, src...), nil },
))
```

### registry.LookupField

```go
func LookupField[T any, F any](r *Registry, fieldName string) (core.Cloner[F], bool)
```

Retrieves the registered field-level `Cloner[F]` for a specific field of struct type T. Returns
`(cloner, true)` when found, `(nil, false)` when not registered. Safe for concurrent use.

**Example:**

```go
cloner, found := registry.LookupField[User, *Address](reg, "HomeAddr")
if found {
    cloned, err := cloner.Clone(user.HomeAddr)
}
```

### registry.HasField

```go
func HasField[T any](r *Registry, fieldName string) bool
```

Reports whether a field-level `Cloner` is registered for the given struct type T and field name.
Safe for concurrent use.

**Example:**

```go
if registry.HasField[User](reg, "HomeAddr") {
    // field cloner is available
}
```

### registry.DeregisterField

```go
func DeregisterField[T any](r *Registry, fieldName string) bool
```

Removes the registered field-level `Cloner` for the given struct type T and field name. Returns
`true` if a cloner was removed, `false` if no cloner was registered for that field. Safe for
concurrent use.

**Example:**

```go
removed := registry.DeregisterField[User](reg, "HomeAddr")
```

### registry.FieldLen

```go
func (r *Registry) FieldLen() int
```

Returns the total number of registered field-level cloners. This is independent of `Len()`, which
counts type-level cloners only. Safe for concurrent use.

### registry.LookupAnyField

```go
func (r *Registry) LookupAnyField(structType reflect.Type, fieldName string) (func(reflect.Value) (reflect.Value, error), bool)
```

Returns a reflect-level clone function for the given struct type and field name. This is the
field-level counterpart of `LookupAny`, enabling the reflection engine to discover and invoke
field-level cloners without knowing the concrete generic types at compile time. The returned
function accepts a `reflect.Value` of the field's type and returns a `reflect.Value` of the same
type (or an error). Returns `(nil, false)` when no field Cloner is registered. Safe for concurrent
use.

---

## Engine Configuration (engine package)

The engine package implements the reflection-based deep copy engine for doppel. It is the **last**
strategy in the priority chain: `Registered Cloner[T] → SelfClonable[T] → engine (reflection)`.

### engine.New

```go
func New(lookup TypeLookup) *Engine
```

Creates an Engine with default options (`PreserveShared` cycle policy). Pass a `*registry.Registry`
(or any `TypeLookup`) to have the engine consult registered `Cloner[T]`s at every node of the
value graph. If the provided `TypeLookup` also implements `FieldLookup`, the engine will
automatically consult field-level cloners during struct cloning. Pass `nil` to rely only on
SelfClonable detection and pure reflection.

### engine.NewWithOptions

```go
func NewWithOptions(lookup TypeLookup, opts Options) *Engine
```

Creates an Engine with explicitly configured options. Use this when you need a non-default
`CyclePolicy`:

```go
eng := engine.NewWithOptions(reg, engine.Options{CyclePolicy: engine.BreakCycles})
```

### engine.TypeLookup (interface)

```go
type TypeLookup interface {
    LookupAny(t reflect.Type) (func(reflect.Value) (reflect.Value, error), bool)
}
```

Implemented by `*registry.Registry`. Allows the engine to consult registered `Cloner[T]`s at
every node of the value graph without importing the registry package.

### engine.FieldLookup (interface)

```go
type FieldLookup interface {
    LookupAnyField(structType reflect.Type, fieldName string) (func(reflect.Value) (reflect.Value, error), bool)
}
```

Implemented by `*registry.Registry`. Allows the engine to consult registered field-level `Cloners`
during struct cloning without importing the registry package. Auto-detected when the `TypeLookup`
provided to `New` / `NewWithOptions` also implements `FieldLookup`.

---

## Core Interfaces (core package)

The core package defines the fundamental interfaces and error types used throughout doppel.

### core.Cloner

```go
type Cloner[T any] interface {
    Clone(src T) (T, error)
}
```

The extension interface for external clone logic. Register implementations with the registry or
pass them directly to `CloneWith`.

### core.SelfClonable

```go
type SelfClonable[T any] interface {
    Clone() (T, error)
}
```

Optional interface for types that can clone themselves. Implement this on your type to enable
`doppel.Clone` and registry-based fallback. The `Clone()` method signature must return `(T, error)`.

### core.NewFuncCloner

```go
func NewFuncCloner[T any](fn func(T) (T, error)) Cloner[T]
```

Convenience wrapper that adapts a plain function into a `Cloner[T]`. This is the most common way
to create cloners for registration.

### core.ErrNoCloner

```go
var ErrNoCloner error
```

Returned by `CloneWithRegistry` when neither a registered Cloner nor a SelfClonable implementation
exists for the requested type. This is an explicit signal to either register a Cloner, implement
SelfClonable, or switch to `CloneDeep` for automatic reflection fallback.

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • API Reference
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="core-concepts.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Core Concepts</span>
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
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"><a href="usage-guide.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Usage Guide</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
            </a></div>
    </div>
</div>
<!-- /Navigation -->


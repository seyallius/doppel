# Cloner Registry

The `registry` package provides a thread-safe, type-keyed store of `Cloner[T]` values. It decouples clone logic from type definitions — you can register cloners for types you don't own, swap implementations at runtime, and conditionally override behavior.

---

## Why a registry?

Not every type can implement `SelfClonable[T]`. You might need to clone:

- A type from a third-party package you can't modify.
- A type that needs external context (e.g., a database connection) to clone.
- A type where you want to register different clone strategies depending on the use case.

The registry solves all of these by mapping `reflect.Type` → `Cloner[T]` and looking up the right cloner at clone time.

---

## Creating a registry

```go
import "github.com/seyallius/doppel/registry"

reg := registry.New()
```

`New()` returns an empty, ready-to-use `Registry`. It is safe for concurrent use immediately. Each call to `New()` returns an independent instance — registering a cloner in one registry never affects another.

---

## Registering type-level cloners

### `registry.Register[T]`

Stores a `Cloner[T]` for type `T`:

```go
import (
    "github.com/seyallius/doppel/core"
    "github.com/seyallius/doppel/registry"
)

reg := registry.New()

// Register a cloner for Address
registry.Register(reg, core.NewFuncCloner(func(src Address) (Address, error) {
    return Address{
        Street: src.Street,
        City:   src.City + "_cloned",
        State:  src.State,
        Zip:    src.Zip,
    }, nil
}))
```

**Key behaviors:**
- Re-registering the same type silently replaces the previous cloner.
- Pointer and value types are stored under different keys: `Address` and `*Address` are distinct entries.
- Safe for concurrent use — backed by a `sync.RWMutex`.

### `registry.Lookup[T]`

Retrieves the `Cloner[T]` for type `T`:

```go
cloner, found := registry.Lookup[Address](reg)
if found {
    cloned, err := cloner.Clone(originalAddress)
}
```

Returns `(nil, false)` when no cloner is registered.

### `registry.Has[T]`

Quick check without retrieving the cloner:

```go
if registry.Has[Address](reg) {
    // ...
}
```

### `registry.Deregister[T]`

Remove a cloner. Calling this on a type with no registration is a safe no-op:

```go
registry.Deregister[Address](reg)
```

### `registry.Len()`

Count of registered type-level cloners:

```go
fmt.Println(reg.Len()) // e.g., 3
```

---

## Using the registry to clone

### `doppel.CloneWithRegistry[T]`

Walks this lookup chain:

```
Registered Cloner[T]  →  SelfClonable[T]  →  core.ErrNoCloner
```

If the type has a registered cloner, it is used directly (fastest path). If not, but the type implements `SelfClonable[T]`, its `Clone()` method is called. If neither applies, `ErrNoCloner` is returned.

```go
cloned, err := doppel.CloneWithRegistry(address, reg)
if err != nil {
    // Handle ErrNoCloner or clone failure
}
```

### `doppel.CloneDeep[T]`

Walks the full priority chain including the reflection engine:

```
Registered Cloner[T]  →  SelfClonable[T]  →  Field Cloner  →  Reflection Engine
```

If the type has no registered cloner and doesn't implement `SelfClonable[T]`, the reflection engine automatically deep-copies every exported field. This means `CloneDeep` **always succeeds** (unless the engine encounters an error):

```go
// Works even without any registration
cloned, err := doppel.CloneDeep(anyValue, reg)
```

Pass `nil` for the registry to use pure reflection with no registered cloners:

```go
cloned, err := doppel.CloneDeep(anyValue, nil)
```

---

## Registry priority over SelfClonable

When both a registered cloner and a `SelfClonable[T]` implementation exist for the same type, the **registry wins**. This is by design — it lets you override a type's default clone behavior for specific use cases:

```go
// User implements SelfClonable[*User], but we register a custom cloner
registry.Register(reg, core.NewFuncCloner(func(src *User) (*User, error) {
    // Custom logic — e.g., sanitize fields, fetch fresh data, etc.
    return &User{Name: "from_registry"}, nil
}))

// This uses the registry cloner, NOT User.Clone()
cloned, _ := doppel.CloneWithRegistry(originalUser, reg)
```

To restore the original `SelfClonable` behavior, deregister the type cloner:

```go
registry.Deregister[*User](reg)
// Now CloneWithRegistry falls through to User.Clone()
```

---

## Composing registry with manual helpers

Registry cloners can use manual helpers internally. This is the intended pattern for complex types:

```go
registry.Register(reg, core.NewFuncCloner(func(src Order) (Order, error) {
    // Use manual.CloneSlice for the Items field
    items, err := manual.CloneSlice(src.Items, func(s Score) (Score, error) {
        return Score{Label: s.Label, Value: s.Value}, nil
    })
    if err != nil {
        return Order{}, core.WrapError("Order.Items", err)
    }

    // Use manual.CloneMap for the Metadata field
    metadata, err := manual.CloneMap(src.Metadata, manual.Identity[string])
    if err != nil {
        return Order{}, core.WrapError("Order.Metadata", err)
    }

    return Order{
        ID:       src.ID,
        Customer: src.Customer, // could also be cloned
        Items:    items,
        Metadata: metadata,
    }, nil
}))
```

---

## Conditional cloning via registry

You can register cloners that implement conditional logic — for example, deep-copying only active users and replacing inactive ones with a placeholder:

```go
registry.Register(reg, core.NewFuncCloner(func(src *User) (*User, error) {
    if !src.Active {
        return &User{ID: src.ID, Name: "[inactive]"}, nil
    }
    return src.Clone() // delegate to SelfClonable
}))
```

---

## Multiple types in one registry

Each type is stored independently. Registering cloners for multiple types in the same registry is fully supported:

```go
reg := registry.New()

registry.Register(reg, core.NewFuncCloner(cloneAddress))
registry.Register(reg, core.NewFuncCloner(cloneScore))
registry.Register(reg, core.NewFuncCloner(cloneUser))

// Each type dispatches to its own cloner
addrCloned, _ := doppel.CloneWithRegistry(addr, reg)
scoreCloned, _ := doppel.CloneWithRegistry(score, reg)
userCloned, _ := doppel.CloneWithRegistry(user, reg)
```

---

## Thread safety

`Registry` is safe for concurrent use. All operations (`Register`, `Lookup`, `Has`, `Deregister`, `Len`) are protected by a `sync.RWMutex`. You can share a single `Registry` instance across goroutines without additional synchronization:

```go
// Safe to share across goroutines
reg := registry.New()
registry.Register(reg, cloner)

var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        cloned, _ := doppel.CloneWithRegistry(value, reg)
        _ = cloned
    }()
}
wg.Wait()
```

---

## Reflection usage in the registry

The registry uses reflection for exactly two purposes, both internal to the store:

1. **Deriving a stable map key** from `T` via `reflect.TypeOf((*T)(nil)).Elem()`. This works correctly even for interface types.
2. **Wrapping stored `Cloner[T]`s** as `reflect.Value`-level functions in `LookupAny`, so the reflection engine can call them without knowing `T` at compile time.

No reflection is used for field access, struct inspection, or value traversal. Actual cloning always delegates to the `Cloner[T]` you registered.

---

## What's next?

- **[Field-Level Cloners](field-cloners.md)** — Override clone behavior for individual struct fields, the core Phase 3 feature.
- **[Reflection Engine](reflection-engine.md)** — How the automatic fallback deep-copies any Go value.

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Cloner Registry
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="manual-helpers.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Manual Helpers</span>
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
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"><a href="field-cloners.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Field Cloners</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
            </a></div>
    </div>
</div>
<!-- /Navigation -->


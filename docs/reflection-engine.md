# Reflection Engine

The reflection engine is the **last strategy** in doppel's priority chain. It is consulted only when neither a registered `Cloner[T]` nor a `SelfClonable[T]` implementation exists for a given type. This means the engine is never the default — always the fallback.

The engine performs recursive, field-by-field deep copying using Go's `reflect` package. It handles all common Go types and integrates with the registry for per-field customization.

---

## When the engine is used

The engine is invoked in these situations:

1. **`doppel.CloneDeep`** with no registered `Cloner[T]` and no `SelfClonable[T]`.
2. **`engine.New(lookup).Clone(reflect.ValueOf(src))`** — direct usage.
3. **Nested fields** during struct cloning — each field goes through the same priority chain.

```
doppel.CloneDeep(value, reg)
    │
    ├─ Registered Cloner[T]?  → Use it (fastest)
    │
    ├─ SelfClonable[T]?       → Call Clone() (fast)
    │
    └─ engine.Clone()         → Recursive reflection (this page)
         │
         ├─ For each struct field:
         │    ├─ Struct tag?     → Apply tag directive
         │    ├─ Field cloner?   → Use it
         │    ├─ Type cloner?    → Use it
         │    ├─ SelfClonable?   → Call Clone()
         │    └─ Reflection      → Recurse
         │
         ├─ Slice elements → Recurse on each element
         ├─ Map values     → Recurse on each key/value pair
         └─ Pointer        → Recurse on pointed-to value
```

---

## Supported types

### Primitives (assignment copy)

These types need no special handling — assignment IS a complete deep copy:

`bool`, `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `uintptr`, `float32`, `float64`, `complex64`, `complex128`, `string`

```go
src := 42
cloned, _ := engine.New(nil).Clone(reflect.ValueOf(src))
// cloned.Int() == 42 — allocated via reflect.New, then Set
```

### Structs

Each exported field is cloned recursively. Unexported fields are **skipped** — they are inaccessible via reflection without `unsafe`. To include unexported fields, implement `SelfClonable[T]` on the type.

```go
type User struct {
    Name   string
    Active bool
    secret string // skipped — not exported
}
```

### Pointers

The pointed-to value is recursively cloned. A fresh allocation is created for the clone:

```go
type Parent struct {
    Child *Child
}

// Parent.Child in the clone points to a new *Child allocation
// Mutating original.Child does not affect clone.Child
```

### Slices

A new backing array is allocated. Each element is recursively cloned:

```go
// Original and clone have independent backing arrays
original.Tags[0] = "mutated"
// clone.Tags[0] still has the original value
```

Nil slices return nil. Empty slices return a fresh empty (non-nil) slice — the nil-vs-empty distinction is preserved.

### Maps

A new map is allocated. Each key and value is recursively cloned:

```go
// Original and clone are independent maps
original.Scores["math"] = 0
// clone.Scores["math"] still has the original value
```

Nil maps return nil. Empty maps return a fresh empty (non-nil) map.

### Arrays

Fixed-length arrays are cloned element-by-element into a new array:

```go
src := [4]int{10, 20, 30, 40}
cloned, _ := engine.New(nil).Clone(reflect.ValueOf(src))
// cloned is a [4]int with the same values
```

### Interfaces

The concrete value stored inside the interface is recursively cloned. The interface type is preserved:

```go
type Holder struct {
    Anything any
}

src := Holder{Anything: 42}
cloned, _ := engine.New(nil).Clone(reflect.ValueOf(src))
// cloned.Anything == 42 (int)
```

### Unsupported types

`Chan`, `Func`, and `UnsafePointer` carry reference semantics that cannot be meaningfully deep-copied. They are **shallow-copied** (the reference is shared between original and clone). This is a deliberate design choice — deep-copying a channel or function makes no semantic sense.

---

## SelfClonable detection

The engine detects `SelfClonable[T]` at runtime using reflection. It checks whether the value (or its pointer) has a `Clone() (T, error)` method with the correct signature:

```go
type MyType struct { Data string }

func (m *MyType) Clone() (*MyType, error) {
    return &MyType{Data: m.Data + "_cloned"}, nil
}

// The engine detects this automatically:
eng := engine.New(nil)
cloned, _ := eng.Clone(reflect.ValueOf(&MyType{Data: "test"}))
// cloned.Data == "test_cloned"
```

Detection works for both value receivers and pointer receivers:

| Receiver | Detection |
|----------|-----------|
| `func (m MyType) Clone() (MyType, error)` | Detected on the value |
| `func (m *MyType) Clone() (*MyType, error)` | Detected on `*MyType` or `MyType` (via `&val`) |

---

## Registry integration

The engine accepts an optional `TypeLookup` (typically `*registry.Registry`) at construction time. At every node of the value graph, it checks the registry before falling through to reflection:

```go
reg := registry.New()
registry.Register(reg, core.NewFuncCloner(myCustomCloner))

eng := engine.New(reg)
// For every value encountered during cloning:
//   1. Check registry.LookupAny(type)
//   2. Check SelfClonable
//   3. Reflect
```

The `FieldLookup` interface is auto-detected when the `TypeLookup` also implements it (as `*registry.Registry` does). This enables field-level cloner integration without additional configuration:

```go
reg := registry.New()
registry.RegisterField[User, *Address](reg, "Address", cloner)

eng := engine.New(reg)
// During struct cloning of User:
//   - Checks for field cloner on "Address" → found → uses it
//   - All other fields → default reflection
```

---

## Creating an engine

```go
// Default: PreserveShared cycle policy, no registry
eng := engine.New(nil)

// With registry (type + field cloners)
eng := engine.New(reg)

// With custom cycle policy
eng := engine.NewWithOptions(reg, engine.Options{
    CyclePolicy: engine.BreakCycles,
})
```

See [Cycle & Sharing Policy](cycle-policy.md) for details on the three policies.

---

## Concurrency

`Engine` is safe for concurrent use. All mutable state lives in per-call `cloneState` values, which are never shared across goroutines:

```go
eng := engine.New(nil)

var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        cloned, err := eng.Clone(reflect.ValueOf(src))
        _ = cloned
        _ = err
    }()
}
wg.Wait()
```

---

## Reflection usage summary

The engine uses reflection for legitimate deep-copy purposes only:

- Reading and writing struct fields.
- Allocating new slices, maps, arrays, and pointers.
- Detecting `SelfClonable` via method lookup.
- Dispatching to registered `Cloner[T]`s via `TypeLookup`/`FieldLookup`.

Reflection is **not** used for:
- Dynamic field access patterns beyond what is needed for deep copy.
- Serialization, deserialization, or any non-cloning purpose.
- Bypassing access controls (unexported fields are skipped).

---

## What's next?

- **[Struct Tags](struct-tags.md)** — Control per-field behavior with `doppel:"..."` annotations.
- **[Cycle & Sharing Policy](cycle-policy.md)** — Handle cyclic and shared pointer graphs.

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Reflection Engine
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="field-cloners.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Field Cloners</span>
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
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"><a href="struct-tags.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Struct Tags</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
            </a></div>
    </div>
</div>
<!-- /Navigation -->


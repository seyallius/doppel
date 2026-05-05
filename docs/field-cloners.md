# Field-Level Cloners

Field-level cloners provide fine-grained control over individual struct fields during a deep copy. Instead of writing a full `Clone()` method for a 200-field struct, register cloners for only the fields that need custom logic — the reflection engine handles the rest automatically.

This is the core **Phase 3** feature and the recommended approach for "default deep copy + selective override."

---

## The problem

Imagine a struct with 200 fields. Most are primitives that need no special handling, but a few need custom clone logic:

```go
type BigStruct struct {
    ID       int64
    Label    string
    Active   bool
    Priority int
    Weight   float64
    Address  *Address   // needs custom cloning (pointer with nested struct)
    Tags     []string   // needs custom cloning (must be independent)
    Scores   map[string]int
    // ... 192 more primitive fields
}
```

Writing a `Clone()` method with 200 fields is tedious, error-prone, and breaks every time you add a field. Field-level cloners solve this by letting you register custom cloners for only the fields that need them.

---

## Registering field cloners

### `registry.RegisterField[T, F]`

Register a `Cloner[F]` for a specific field `F` of struct type `T`:

```go
reg := registry.New()

// Custom clone for the Address pointer field
registry.RegisterField[BigStruct, *Address](reg, "Address", core.NewFuncCloner(
    func(src *Address) (*Address, error) {
        if src == nil {
            return nil, nil
        }
        return &Address{
            Street: src.Street,
            City:   src.City + "_cloned",
            State:  src.State,
            Zip:    src.Zip,
        }, nil
    },
))

// Custom clone for the Tags slice field
registry.RegisterField[BigStruct, []string](reg, "Tags", core.NewFuncCloner(
    func(src []string) ([]string, error) {
        return append([]string{}, src...), nil
    },
))
```

**Constraints enforced at registration time:**
- `T` must be a struct type (or pointer to a struct). Panics otherwise.
- `fieldName` must name an exported field of `T`. Panics for unexported fields or non-existent fields.
- Registering the same `(structType, fieldName)` pair twice silently replaces the first registration.

### Clone with field cloners

Use `doppel.CloneDeep` to trigger the full dispatch chain, including field cloners:

```go
src := BigStruct{
    ID:       1,
    Label:    "test",
    Active:   true,
    Priority: 5,
    Weight:   3.14,
    Address:  &Address{Street: "1 Main", City: "Metro"},
    Tags:     []string{"important", "urgent"},
    Scores:   map[string]int{"quality": 95},
}

cloned, err := doppel.CloneDeep(src, reg)
```

The engine:
1. Checks each field for a registered field cloner.
2. If found, uses it.
3. If not found, clones via default reflection (or type-level cloner, or SelfClonable).

---

## Per-field priority chain

Inside the reflection engine's `cloneStruct`, each field is processed in this priority order:

```
1. Struct tag directive    (doppel:"-", doppel:"shallow", doppel:"clone")
2. Registered field Cloner  (auto-discovered by field name)
3. Registered type Cloner   (for the field's type, via registry)
4. SelfClonable on the field value
5. Reflection fallback      (recursive deep copy)
```

This means:
- A struct tag always wins, regardless of registrations.
- A field cloner wins over a type cloner for the same field.
- A type cloner is checked before SelfClonable and reflection.

---

## Field cloner vs. type cloner priority

When both a field-level cloner and a type-level cloner exist for the same field, the **field cloner wins** for that specific field. Other fields of the same type continue using the type cloner:

```go
reg := registry.New()

// Type-level cloner: appends "_type" to Label
registry.Register(reg, core.NewFuncCloner(func(src *Nested) (*Nested, error) {
    return &Nested{Label: src.Label + "_type", Count: src.Count}, nil
}))

// Field-level cloner: appends "_field" to Label (wins for fieldHost.Nested)
registry.RegisterField[FieldHost, *Nested](reg, "Nested", core.NewFuncCloner(
    func(src *Nested) (*Nested, error) {
        return &Nested{Label: src.Label + "_field", Count: src.Count}, nil
    },
))
```

For `fieldHost.Nested`, the field cloner wins (Label becomes "inner_field"). For other structs that have a `*Nested` field but no registered field cloner, the type cloner applies (Label becomes "inner_type").

---

## Querying field cloners

| Function | Description |
|----------|-------------|
| `registry.LookupField[T, F](reg, "name")` | Retrieve the registered `Cloner[F]` for a field. Returns `(nil, false)` if not found. |
| `registry.HasField[T](reg, "name")` | Check if a field cloner is registered. |
| `registry.DeregisterField[T](reg, "name")` | Remove a field cloner. Returns `true` if one was removed. |
| `reg.FieldLen()` | Count of registered field-level cloners. |

### DeregisterField

After deregistration, the field falls back to the default reflection path:

```go
// Before: field cloner is used
cloned1, _ := doppel.CloneDeep(src, reg)
// cloned1.Nested.Label == "inner_field"

// Deregister
registry.DeregisterField[FieldHost](reg, "Nested")

// After: default reflection is used
cloned2, _ := doppel.CloneDeep(src, reg)
// cloned2.Nested.Label == "inner" (no "_field" suffix)
```

---

## Same field name, different struct types

Field cloners are keyed by `(structType, fieldName)` pairs. Two struct types with a field of the same name are independent:

```go
type HostA struct { Nested *Nested }
type HostB struct { Nested *Nested }

reg := registry.New()

registry.RegisterField[HostA, *Nested](reg, "Nested", core.NewFuncCloner(cloneA))
registry.RegisterField[HostB, *Nested](reg, "Nested", core.NewFuncCloner(cloneB))

// HostA.Nested uses cloneA
// HostB.Nested uses cloneB
// They are completely independent
```

---

## Pointer struct types

`RegisterField` accepts both value and pointer struct types. `*User` and `User` resolve to the same struct type key:

```go
// These are equivalent:
registry.RegisterField[User, *Address](reg, "HomeAddr", cloner)
registry.RegisterField[*User, *Address](reg, "HomeAddr", cloner)
```

---

## Multiple field cloners on the same struct

Register cloners for as many fields as needed on the same struct. Fields without registered cloners are cloned by default reflection:

```go
reg := registry.New()
registry.RegisterField[User, *Address](reg, "Address", cloneAddrCloner)
registry.RegisterField[User, []string](reg, "Tags", cloneTagsCloner)

// User.Name, User.ID, User.Active, User.Scores — all cloned by default reflection
// User.Address — cloned by registered field cloner
// User.Tags — cloned by registered field cloner
```

---

## Error propagation

If a field cloner returns an error, it propagates up with the field path annotated:

```go
registry.RegisterField[User, *Address](reg, "Address", core.NewFuncCloner(
    func(src *Address) (*Address, error) {
        return nil, errors.New("database unavailable")
    },
))

_, err := doppel.CloneDeep(user, reg)
// err.Error() contains "User.Address" and "database unavailable"
```

---

## Complete example: BigStruct scenario

This is the primary Phase 3 use case — a struct with many fields where only a few need custom cloning:

```go
reg := registry.New()

// Only register cloners for the 2 fields that need custom logic.
// The other 198 fields are cloned automatically by reflection.
registry.RegisterField[BigStruct, *Address](reg, "Address", core.NewFuncCloner(
    func(src *Address) (*Address, error) {
        if src == nil { return nil, nil }
        return &Address{Street: src.Street, City: src.City, State: src.State, Zip: src.Zip}, nil
    },
))
registry.RegisterField[BigStruct, []string](reg, "Tags", core.NewFuncCloner(
    func(src []string) ([]string, error) {
        return append([]string{}, src...), nil
    },
))

cloned, err := doppel.CloneDeep(bigStruct, reg)
if err != nil {
    panic(err)
}

// All primitive fields are correctly copied.
// Address and Tags use the custom field cloners.
// Everything is independent from the original.
```

---

## What's next?

- **[Reflection Engine](reflection-engine.md)** — Understand how the engine handles each Go type kind when no field cloner is registered.
- **[Struct Tags](struct-tags.md)** — Use `doppel:"..."` tags for declarative field-level control.

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Field-Level Cloners
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="registry.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Registry</span>
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


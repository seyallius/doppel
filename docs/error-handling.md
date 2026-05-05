# Error Handling

doppel provides a structured error type with field-path context, making it straightforward to identify which field or index triggered a cloning failure. All errors implement `Unwrap()` for compatibility with `errors.Is` and `errors.As`.

---

## Error types

### `core.CloneError`

The primary error type, carrying context about where the failure occurred:

```go
type CloneError struct {
    Context string // Human-readable path to the failing field, e.g. "User.Address"
    Cause   error  // The underlying error that triggered the failure
}
```

**Formatted message:** `doppel: error cloning User.Address: <cause>`

**Usage in your own `Clone()` methods:**

```go
func (u *User) Clone() (*User, error) {
    addr, err := manual.ClonePointer(u.Address, cloneAddress)
    if err != nil {
        return nil, core.WrapError("User.Address", err)
    }
    // ...
}
```

### `engine.CycleError`

Returned when `CyclePolicy` is `ErrorOnCycle` and a back-edge is detected:

```go
type CycleError struct {
    Addr     uintptr // Raw pointer address of the back-edge
    TypeName string  // reflect.Type.String() of the value at that address
}
```

**Formatted message:** `doppel/engine: cycle detected at address 0x... (type *pkg.Node); use BreakCycles or PreserveShared policy to handle cyclic graphs`

---

## Sentinel errors

### `core.ErrNoCloner`

Returned by `doppel.CloneWithRegistry` when neither a registered `Cloner[T]` nor a `SelfClonable[T]` implementation exists:

```go
_, err := doppel.CloneWithRegistry(someValue, emptyReg)
if errors.Is(err, core.ErrNoCloner) {
    // No cloner registered and type doesn't implement SelfClonable
    // Use CloneDeep instead for automatic reflection fallback
}
```

### `core.ErrNilSource`

Reserved for future strict-nil modes. Not used by default — all cloners propagate nil as nil (no error).

---

## Creating errors

### `core.WrapError`

The primary way to annotate errors with context:

```go
func (u *User) Clone() (*User, error) {
    tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
    if err != nil {
        return nil, core.WrapError("User.Tags", err)
    }

    scores, err := manual.CloneMap(u.Scores, manual.Identity[int])
    if err != nil {
        return nil, core.WrapError("User.Scores", err)
    }

    addr, err := manual.ClonePointer(u.Address, cloneAddress)
    if err != nil {
        return nil, core.WrapError("User.Address", err)
    }

    return &User{ID: u.ID, Name: u.Name, Tags: tags, Scores: scores, Address: addr}, nil
}
```

Each `WrapError` call adds one level to the error chain. When an error from a manual helper propagates, you get a nested path:

```
doppel: error cloning User.Address: doppel: pointer: <underlying error>
doppel: error cloning Order.Items: doppel: CloneSlice index [2]: <underlying error>
```

---

## Inspecting errors

### `errors.Is` — Check for specific errors

```go
sentinel := errors.New("validation failed")
_, err := manual.CloneSlice(src, func(s string) (string, error) {
    if s == "" {
        return "", sentinel
    }
    return s, nil
})

if errors.Is(err, sentinel) {
    fmt.Println("The sentinel error is in the chain")
}
```

`errors.Is` unwraps through `CloneError.Unwrap()`, so it works regardless of how many levels of `WrapError` are in the chain.

### `errors.As` — Extract typed errors

```go
var cloneErr *core.CloneError
if errors.As(err, &cloneErr) {
    fmt.Printf("Failed at: %s\n", cloneErr.Context)
    fmt.Printf("Caused by: %v\n", cloneErr.Cause)
}

var cycleErr *engine.CycleError
if errors.As(err, &cycleErr) {
    fmt.Printf("Cycle at 0x%x, type %s\n", cycleErr.Addr, cycleErr.TypeName)
}
```

### Error chain inspection

Because each `WrapError` wraps the previous error, you can walk the full chain:

```go
// Manual helper error:
//   doppel: CloneSlice index [2]: validation failed

// After WrapError("User.Tags"):
//   doppel: error cloning User.Tags: doppel: CloneSlice index [2]: validation failed

// After WrapError("Order.User"):
//   doppel: error cloning Order.User: doppel: error cloning User.Tags: doppel: CloneSlice index [2]: validation failed
```

---

## Error context from manual helpers

Each manual helper adds its own context to errors:

| Helper | Error context format |
|--------|---------------------|
| `CloneSlice` | `doppel: CloneSlice index [2]: <your error>` |
| `CloneMap` | `doppel: CloneMap key [myKey]: <your error>` |
| `ClonePointer` | `doppel: pointer: <your error>` |

The engine adds context for every struct field and map key it processes during reflection:

| Context | Example |
|---------|---------|
| Struct field | `doppel: error cloning User.Contact: ...` |
| Slice index | `doppel: error cloning [3]: ...` |
| Map key | `doppel: error cloning map[username]: ...` |
| Map value | `doppel: error cloning map[key].value: ...` |

---

## Error propagation patterns

### Sentinel errors in clone functions

```go
var ErrValidation = errors.New("field validation failed")

func cloneUser(src *User) (*User, error) {
    if src.Name == "" {
        return nil, ErrValidation
    }
    // ...
}

_, err := doppel.CloneDeep(user, reg)
if errors.Is(err, ErrValidation) {
    // A user with an empty name was cloned
}
```

### Registry cloner errors

```go
registry.Register(reg, core.NewFuncCloner(func(src *ExternalType) (*ExternalType, error) {
    result, err := fetchFromDatabase(src.ID)
    if err != nil {
        return nil, fmt.Errorf("database fetch: %w", err)
    }
    return result, nil
}))
```

### CycleError handling

```go
eng := engine.NewWithOptions(nil, engine.Options{
    CyclePolicy: engine.ErrorOnCycle,
})

_, err := eng.Clone(reflect.ValueOf(cyclicValue))
if errors.As(err, &engine.CycleError{}) {
    // Decide: use BreakCycles, or fix the data
    eng2 := engine.NewWithOptions(nil, engine.Options{
        CyclePolicy: engine.BreakCycles,
    })
    cloned, _ := eng2.Clone(reflect.ValueOf(cyclicValue))
}
```

---

## Best practices

1. **Always wrap errors from manual helpers.** Use `core.WrapError("StructName.FieldName", err)` to build a navigable error chain.
2. **Use sentinel errors for known failure modes.** `errors.Is` makes it easy to distinguish between different failure reasons.
3. **Handle nil inputs explicitly.** Return `nil, nil` for nil inputs — this is the library convention and avoids spurious errors.
4. **Let the engine handle context for you.** When using `CloneDeep`, the engine automatically annotates every field and index in the error path.
5. **Use `Must` variants in tests.** `doppel.MustClone`, `doppel.MustCloneWith`, and `doppel.MustCloneDeep` panic on error — appropriate for test fixtures and initialization code.

---

## What's next?

- **[Patterns & Best Practices](patterns.md)** — Complete patterns for composing errors, handling edge cases, and structuring clone methods.
- **[API Reference](api-reference.md)** — Full error type signatures.

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Error Handling
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="cycle-policy.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Cycle Policy</span>
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
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"><a href="patterns.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Patterns</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
            </a></div>
    </div>
</div>
<!-- /Navigation -->


# ⚙️ Advanced Topics

> Error handling, nil safety, struct tags, and pro tips for production use. ✨

---

## Error Handling

Every fallible helper returns `(T, error)`. Errors are wrapped with `core.WrapError` at each layer, building a context
path:

```
doppel: error cloning Order.Customer: doppel: error cloning User.Contact: doppel: error cloning ContactInfo.Address: pointer: <root cause>
```

### Using `errors.Is` and `errors.As`

Errors implement `Unwrap()`, so standard error inspection works:

```go
cloned, err := doppel.Clone(order)
if err != nil {
    var cloneErr *core.CloneError
    if errors.As(err, &cloneErr) {
        log.Printf("failed at: %s", cloneErr.Context)
    }
}
```

### When to Use `MustClone` vs `Clone`

| Scenario                  | Recommended                       | Why                                        |
|---------------------------|-----------------------------------|--------------------------------------------|
| Production business logic | `Clone` + explicit error handling | Graceful degradation                       |
| Test setup / fixtures     | `MustClone`                       | Fail fast on programming errors            |
| Program initialization    | `MustClone`                       | Clone failure = bug, not runtime condition |

---

## Nil Safety Contract

All helpers treat nil consistently and without error:

| Input                       | Output                      |
|-----------------------------|-----------------------------|
| `nil *T` to `ClonePointer`  | `(nil, nil)`                |
| `nil []T` to `CloneSlice`   | `(nil, nil)`                |
| `nil map[K]V` to `CloneMap` | `(nil, nil)`                |
| Empty (non-nil) slice       | Fresh empty slice (not nil) |
| Empty (non-nil) map         | Fresh empty map (not nil)   |

✅ Clones faithfully preserve the nil-vs-empty distinction. This matters for JSON marshaling, database NULLs, and API
contracts.

### Example: Preserving Nil vs Empty

```go
type Config struct {
    Tags []string // nil vs []string{} matters
}

// Original with nil slice
orig := Config{Tags: nil}
cloned, _ := orig.Clone()

fmt.Println(orig.Tags == nil) // true
fmt.Println(cloned.Tags == nil) // true ← preserved!

// Original with empty slice
orig2 := Config{Tags: []string{}}
cloned2, _ := orig2.Clone()

fmt.Println(len(orig2.Tags)) // 0
fmt.Println(len(cloned2.Tags)) // 0
fmt.Println(cloned2.Tags == nil) // false ← fresh empty slice
```

---

## Struct Tag Directives

The `engine.Engine` respects the following `doppel` struct tags for fine-grained control:

```go
type Example struct {
    SkipMe    string   `doppel:"-"`       // ← skipped; clone gets zero value
    ShareMe   []string `doppel:"shallow"` // ← shallow copy; shares backing array
    DeepClone string // ← default: deep copy recursively
}
```

| Tag                | Behavior                                                                                                                |
|--------------------|-------------------------------------------------------------------------------------------------------------------------|
| `doppel:"-"`       | Field is skipped entirely; clone receives zero value                                                                    |
| `doppel:"shallow"` | Field is assigned without recursing; clone shares the field's value (useful for immutable or reference-semantics types) |

> ⚠️ **Note**: Struct tags are **only processed by the reflection engine** (`engine/`). Manual clone methods and
> registry cloners ignore tags — you control the logic explicitly.

---

## Best Practices

### ✅ Do

- Wrap every error with `core.WrapError` for contextual debugging
- Use `Identity[T]` helpers for primitives to make intent explicit
- Implement `SelfClonable[T]` for your own domain types
- Benchmark manual vs reflection paths for hot code
- Use `MustClone` only in tests/init, not production logic

### ❌ Don't

- Use reflection fallback for performance-critical paths without benchmarking
- Skip nil checks in `Clone()` methods — handle `nil` receiver gracefully
- Forget that `CloneSlice` returns a fresh empty slice (not nil) for empty input
- Assume struct tags work with manual cloning — they don't!

### 🔁 Composition Pattern

```go
// Preferred: compose helpers, wrap errors, return early
func (o *Order) Clone() (*Order, error) {
    if o == nil {
        return nil, nil
    }
    customer, err := manual.ClonePointer(o.Customer, cloneCustomer)
    if err != nil {
        return nil, core.WrapError("Order.Customer", err)
    }
    items, err := manual.CloneSlice(o.Items, cloneItem)
    if err != nil {
        return nil, core.WrapError("Order.Items", err)
    }
    return &Order{Customer: customer, Items: items}, nil
}
```

> 💡 **Golden Rule**: Make every clone decision visible in code. If you can't see it, you can't debug it. (◕‿◕)✧

<!--

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Advanced Topics
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap;">
        <div style="flex: 1; min-width: 200px;">
            <a href="usage-guide.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
            <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
            <span style="display: flex; flex-direction: column; line-height: 1.3;">
          <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
          <span style="font-size: 1rem; font-weight: 600;">Usage Guide</span>
        </span>
        </a>
        </div>
        <div style="flex: 1; min-width: 200px;">
            <a href="reflection-engine.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
        <span style="display: flex; flex-direction: column; line-height: 1.3;">
          <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
          <span style="font-size: 1rem; font-weight: 600;">Reflection Engine</span>
        </span>
            <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
        </a>
        </div>
    </div>
</div>
<!-- /Navigation -->


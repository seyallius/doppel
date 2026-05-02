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

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->
<div class="doppel-nav-container"
     style="margin-top: 3rem; padding: 1.75rem; border-top: 1px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c); box-shadow: 0 8px 24px rgba(0,0,0,0.4);">
    <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 1rem;">
        <div style="flex: 1; min-width: 200px;">
            <a href="usage-guide.md" class="doppel-nav-btn doppel-nav-prev">
            <span class="doppel-arrow">←</span>
            <div class="doppel-text">
                <span class="doppel-label">Previous</span>
                <span class="doppel-title">Usage Guide</span>
            </div>
        </a>
        </div>
        <div style="flex: 1; min-width: 200px; text-align: right;">
            <a href="reflection-engine.md" class="doppel-nav-btn doppel-nav-next">
            <div class="doppel-text">
                <span class="doppel-label">Next</span>
                <span class="doppel-title">Reflection Engine</span>
            </div>
            <span class="doppel-arrow">→</span>
        </a>
        </div>
    </div>
    <div style="margin-top: 1.25rem; text-align: center; color: #94a3b8; font-size: 0.8rem; letter-spacing: 0.03em;">
        <span>📚 doppel Documentation • Advanced Topics</span>
    </div>
</div>

<style>
    .doppel-nav-btn {
        display: inline-flex;
        align-items: center;
        gap: 0.75rem;
        padding: 0.85rem 1.5rem;
        border-radius: 10px;
        font-weight: 500;
        text-decoration: none;
        color: #ffffff;
        position: relative;
        overflow: hidden;
        transition: transform 0.25s cubic-bezier(0.4, 0, 0.2, 1),
        box-shadow 0.25s cubic-bezier(0.4, 0, 0.2, 1),
        background 0.3s ease;
        box-shadow: 0 4px 10px rgba(0, 0, 0, 0.3);
    }
    
    /* Base Gradients */
    .doppel-nav-prev {
        background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
    }
    
    .doppel-nav-next {
        background: linear-gradient(135deg, #10b981 0%, #047857 100%);
    }
    
    /* Hover Animation */
    .doppel-nav-btn:hover {
        transform: translateY(-3px) scale(1.02);
        box-shadow: 0 12px 24px rgba(0, 0, 0, 0.4);
    }
    
    .doppel-nav-prev:hover {
        background: linear-gradient(135deg, #60a5fa 0%, #2563eb 100%);
    }
    
    .doppel-nav-next:hover {
        background: linear-gradient(135deg, #34d399 0%, #059669 100%);
    }
    
    /* Active/Click State */
    .doppel-nav-btn:active {
        transform: translateY(0) scale(0.97);
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
    }
    
    /* Focus for keyboard accessibility */
    .doppel-nav-btn:focus-visible {
        outline: 2px solid #60a5fa;
        outline-offset: 3px;
        border-radius: 12px;
    }
    
    /* Directional Arrow Slide */
    .doppel-arrow {
        font-size: 1.2rem;
        transition: transform 0.25s cubic-bezier(0.4, 0, 0.2, 1);
    }
    
    .doppel-nav-prev:hover .doppel-arrow {
        transform: translateX(-4px);
    }
    
    .doppel-nav-next:hover .doppel-arrow {
        transform: translateX(4px);
    }
    
    /* Typography */
    .doppel-text {
        display: flex;
        flex-direction: column;
        line-height: 1.25;
    }
    
    .doppel-label {
        font-size: 0.65rem;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        opacity: 0.85;
        margin-bottom: 2px;
    }
    
    .doppel-title {
        font-size: 0.95rem;
        font-weight: 600;
    }
    
    /* Mobile Responsiveness */
    @media (max-width: 768px) {
        .doppel-nav-container > div:first-child {
            flex-direction: column !important;
            gap: 1rem !important;
        }
        
        .doppel-nav-container > div:last-child {
            text-align: left !important;
        }
    }
</style>
<!-- /Navigation -->

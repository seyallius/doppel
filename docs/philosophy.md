# 💭 Design Philosophy

> The principles that guide every decision in `doppel`. Explicit over magic, always. ✨

---

## Core Principles

### 1. Manual Cloning is the Default
By default, no reflection is used — not even for type identification. Generic helpers (`CloneSlice`, `CloneMap`,
`ClonePointer`) are resolved entirely at compile time.

✅ You write the `Clone()` method
✅ `doppel` gives you concise, safe helpers
✅ Every copy decision is visible in your code

### 2. Explicit Over Magic

You write the `Clone()` method. `doppel` provides helpers to make it concise and safe, but the logic is always yours to
read and reason about.

❌ No hidden orchestration
❌ No global state
✅ Full control, full visibility

### 3. Composable, Not Monolithic

Each helper does exactly one thing. You wire them together inside your type's own `Clone()` method.

```go
// Example composition in a Clone() method
func (o *Order) Clone() (*Order, error) {
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

### 4. Open for Extension, Closed for Modification
The `Cloner[T]` interface is the single extension point. Registering custom cloners, adding per-field logic, or opting into reflection are all additive — the core manual cloning foundation never changes.
### 5. Errors Carry Context

Every helper wraps failures with a field-path string (`core.WrapError`) so that when a clone fails deep inside a nested
struct, the error message tells you exactly which field triggered it:

```
doppel: error cloning Order.Customer: doppel: error cloning User.Contact: doppel: error cloning ContactInfo.Address: pointer: <root cause>
```

### 6. Graceful Degradation

When manual clone logic isn't available, the reflection engine provides a safe, correct fallback — with clear
performance trade-offs documented.

---

## Priority Chain

| Priority | Strategy                                       | When Used                                   | Performance      |
|----------|------------------------------------------------|---------------------------------------------|------------------|
| 1        | **Manual clone** (`SelfClonable[T]`)           | Always, by default                          | ⚡ Fastest        |
| 2        | **External `Cloner[T]`** (`CloneWith`)         | Injected context, third-party types         | ⚡ Fast           |
| 3        | **Registry `Cloner[T]`** (`CloneWithRegistry`) | Type-level override, no source modification | ⚡ Fast           |
| 4        | **Reflection fallback** (`engine.Engine`)      | Last resort, prototyping, legacy            | 🐌 Slower (3-6×) |

> 🔄 The chain is evaluated in order. Reflection is **never** the default — always the last resort.

---

## When to Use What

| Scenario                           | Recommended Strategy                  | Why                                      |
|------------------------------------|---------------------------------------|------------------------------------------|
| Your own domain structs            | `SelfClonable[T]` + manual helpers    | Full control, best performance           |
| Third-party types you can't modify | External `Cloner[T]` via `CloneWith`  | No source changes needed                 |
| Many types, centralized config     | Registry via `CloneWithRegistry`      | Type-level overrides, thread-safe        |
| Prototyping / legacy migration     | Reflection fallback (`engine.Engine`) | "Just works", but document the trade-off |
| Performance-critical paths         | Always manual                         | 3-6× faster, fewer allocations           |

---

## Anti-Patterns to Avoid

❌ Using reflection fallback for hot-path types without benchmarking
❌ Skipping error wrapping — lose context on failures
❌ Forgetting nil safety contracts — preserve nil vs empty distinctions
❌ Overusing `MustClone` in production — prefer explicit error handling

> 💡 **Golden Rule**: If you can write the clone logic explicitly, do it. Reflection is a safety net, not a crutch. (◕‿◕)✧

<!--

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Design Philosophy
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="getting-started.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Getting Started</span>
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
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"><a href="core-concepts.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Core Concepts</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
            </a></div>
    </div>
</div>
<!-- /Navigation -->


# 💭 Design Philosophy

> The principles that guide every decision in `doppel`. Explicit over magic, always. ✨

---

## Core Principles

### 1. Manual Cloning is the Default

No reflection is used anywhere in Phase 1 — not even for type identification. Generic helpers (`CloneSlice`, `CloneMap`,
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

The `Cloner[T]` interface is the single extension point. Registering a custom cloner (Phase 2), adding per-field logic (
Phase 3), or opting into reflection (Phase 4) are all additive — nothing in Phase 1 changes.

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

<!-- Navigation -->
<div style="margin-top: 3rem; padding: 1.5rem; border-top: 2px solid #e1e4e8; border-radius: 6px; background: #f6f8fa;">
  <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 1rem;">
    <div style="flex: 1; min-width: 200px;">
      <a href="./getting-started.md" style="display: inline-flex; align-items: center; padding: 0.75rem 1rem; background: #0366d6; color: white; text-decoration: none; border-radius: 6px; font-weight: 600; transition: background 0.2s;">
        <span style="margin-right: 0.5rem;">←</span>
        <div style="display: flex; flex-direction: column;">
          <span style="font-size: 0.75rem; opacity: 0.9;">Previous:</span>
          <span>Getting Started</span>
        </div>
      </a>
    </div>
    <div style="flex: 1; min-width: 200px; text-align: right;">
      <a href="./core-concepts.md" style="display: inline-flex; align-items: center; padding: 0.75rem 1rem; background: #28a745; color: white; text-decoration: none; border-radius: 6px; font-weight: 600; transition: background 0.2s;">
        <div style="display: flex; flex-direction: column; margin-right: 0.5rem;">
          <span style="font-size: 0.75rem; opacity: 0.9;">Next:</span>
          <span>Core Concepts →</span>
        </div>
      </a>
    </div>
  </div>
  <div style="margin-top: 1rem; text-align: center; color: #586069; font-size: 0.875rem;">
    <span>📚 doppel Documentation • Design Philosophy</span>
  </div>
</div>

<style>
@media (max-width: 768px) {
  div[style*="margin-top: 3rem"] div[style*="display: flex"] {
    flex-direction: column !important;
  }
  div[style*="margin-top: 3rem"] div[style*="text-align: right"] {
    text-align: left !important;
  }
}
</style>

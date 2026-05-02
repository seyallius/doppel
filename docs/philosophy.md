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

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->
<div class="doppel-nav-container"
     style="margin-top: 3rem; padding: 1.75rem; border-top: 1px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c); box-shadow: 0 8px 24px rgba(0,0,0,0.4);">
    <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 1rem;">
        <div style="flex: 1; min-width: 200px;">
            <a href="getting-started.md" class="doppel-nav-btn doppel-nav-prev">
            <span class="doppel-arrow">←</span>
            <div class="doppel-text">
                <span class="doppel-label">Previous</span>
                <span class="doppel-title">Getting Started</span>
            </div>
        </a>
        </div>
        <div style="flex: 1; min-width: 200px; text-align: right;">
            <a href="core-concepts.md" class="doppel-nav-btn doppel-nav-next">
            <div class="doppel-text">
                <span class="doppel-label">Next</span>
                <span class="doppel-title">Core Concepts</span>
            </div>
            <span class="doppel-arrow">→</span>
        </a>
        </div>
    </div>
    <div style="margin-top: 1.25rem; text-align: center; color: #94a3b8; font-size: 0.8rem; letter-spacing: 0.03em;">
        <span>📚 doppel Documentation • Design Philosophy</span>
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

# 🔍 Reflection Engine (Fallback)

> When manual cloning isn't available: safe, correct, but slower. Use judiciously. ✨

---

## When Reflection is Used

The reflection engine (`engine.Engine`) is **never the default**. It is consulted only when:

1. ❌ No `Cloner[T]` is registered for the type, AND
2. ❌ The value does not implement `core.SelfClonable[T]`

```
Priority Chain:
Registered Cloner[T] → SelfClonable[T] → engine.Engine (reflection)
```

### When to Use Reflection Fallback

✅ Prototyping with third-party types you can't modify  
✅ Legacy code migration where manual `Clone()` isn't feasible yet  
✅ Dynamic scenarios where type isn't known at compile time  
⚠️ **Performance note**: Manual cloning is 3-6× faster; use reflection fallback judiciously

---

## Engine API

```go
// Engine performs reflection-based deep copying as the LAST resort.
type Engine struct { /* unexported */ }

// Options configures an Engine at construction time.
type Options struct {
    CyclePolicy CyclePolicy // zero value = PreserveShared
}

// CyclePolicy controls how Engine handles cyclic and shared pointer references.
type CyclePolicy int

const (
    PreserveShared CyclePolicy = iota // default: reproduce exact graph topology
    BreakCycles                       // break back-edges → nil, acyclic output
    ErrorOnCycle                      // return *CycleError on any back-edge
)

// CycleError is returned when CyclePolicy is ErrorOnCycle and a cycle is detected.
type CycleError struct {
    Addr     uintptr // pointer address where cycle detected
    TypeName string  // reflect.Type.String() for debugging
}

// New creates an Engine with default options (PreserveShared cycle policy).
func New(lookup TypeLookup) *Engine

// NewWithOptions creates an Engine with explicitly configured options.
func NewWithOptions(lookup TypeLookup, opts Options) *Engine

// Clone deep-copies src and returns a reflect.Value of the same type.
func (e *Engine) Clone(src reflect.Value) (reflect.Value, error)
```

---

## Cycle Policies

### `PreserveShared` (Default)

Reproduces exact graph topology. Shared pointers in the original remain shared in the clone.

```go
// Self-loop: n → n
n := &Node{Value: 42}
n.Next = n

eng := engine.New(nil) // default: PreserveShared
clonedVal, _ := eng.Clone(reflect.ValueOf(n))
cloned := clonedVal.Interface().(*Node)

// Cycle preserved: cloned.Next == cloned
```

✅ Use for: General-purpose cloning, faithful reproduction

### `BreakCycles`

Breaks back-edges by setting them to `nil`, producing an acyclic output safe for serialization.

```go
eng := engine.NewWithOptions(nil, engine.Options{
    CyclePolicy: engine.BreakCycles,
})
clonedVal, _ := eng.Clone(reflect.ValueOf(n))
cloned := clonedVal.Interface().(*Node)

// Cycle broken: cloned.Next == nil ← safe for JSON!
```

✅ Use for: JSON/YAML serialization, tree conversion, avoiding infinite loops

### `ErrorOnCycle`

Returns `*CycleError` on any cycle for strict validation during development.

```go
eng := engine.NewWithOptions(nil, engine.Options{
    CyclePolicy: engine.ErrorOnCycle,
})
_, err := eng.Clone(reflect.ValueOf(n))
// err.(*engine.CycleError) → "cycle detected at 0x... (type *GraphNode)"
```

✅ Use for: Development-time assertions, data model validation

---

## Features & Limitations

### ✅ Features

- Kind-specific cloning: `Ptr`, `Struct`, `Slice`, `Map`, `Array`, `Interface`, primitives
- Struct tag support: `doppel:"-"` (skip), `doppel:"shallow"` (share backing)
- Configurable cycle handling via `CyclePolicy`
- Shared reference preservation under `PreserveShared`
- Nil-safety: preserves `nil` vs `empty` distinction for slices/maps/pointers
- Error context: wraps failures with field-path via `core.WrapError`
- Concurrency: all mutable state is per-call; `Engine` is safe for concurrent use

### ❌ Limitations

- Unexported fields are skipped (use `SelfClonable[T]` to include them)
- `chan`, `func`, `unsafe.Pointer` are shallow-copied (reference semantics)
- Interface values cloned via concrete type; dynamic dispatch preserved

---

## Performance Considerations

### Benchmark Comparison (Indicative)

| Benchmark        | Manual (ns/op) | Reflection (ns/op) | Speedup   |
|------------------|----------------|--------------------|-----------|
| `Score`          | 22.03 ± 8%     | 131.8 ± 3%         | **~6×**   |
| `User`           | 309.9 ± 2%     | 1.193µ ± 4%        | **~4×**   |
| `Order`          | 615.9 ± 4%     | 2.183µ ± 4%        | **~3.5×** |
| `UserLargeSlice` | 8.363µ ± 3%    | 32.76µ ± 0%        | **~4×**   |
| `UserLargeMap`   | 29.26µ ± 0%    | 99.41µ ± 4%        | **~3.4×** |

### Key Takeaways

- 🚀 Manual cloning is **3-6× faster** than reflection-based cloning
- 🧠 Manual cloning uses **~40-95% fewer allocations**, especially with large maps
- ⚡ The gap grows with complexity — nested structs and large collections benefit most
- 🔁 Reflection fallback is still **correct and safe** — use when convenience outweighs performance

> 💡 **Pro Tip**: When using the reflection engine, register cloners for hot-path types via `TypeLookup`. A registry hit
> reduces reflection overhead to near-zero. (◕‿◕)✧

<!-- Navigation -->
<div style="margin-top: 3rem; padding: 1.5rem; border-top: 2px solid #e1e4e8; border-radius: 6px; background: #f6f8fa;">
  <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 1rem;">
    <div style="flex: 1; min-width: 200px;">
      <a href="./advanced.md" style="display: inline-flex; align-items: center; padding: 0.75rem 1rem; background: #0366d6; color: white; text-decoration: none; border-radius: 6px; font-weight: 600; transition: background 0.2s;">
        <span style="margin-right: 0.5rem;">←</span>
        <div style="display: flex; flex-direction: column;">
          <span style="font-size: 0.75rem; opacity: 0.9;">Previous:</span>
          <span>Advanced</span>
        </div>
      </a>
    </div>
    <div style="flex: 1; min-width: 200px; text-align: right;">
      <a href="./testing.md" style="display: inline-flex; align-items: center; padding: 0.75rem 1rem; background: #28a745; color: white; text-decoration: none; border-radius: 6px; font-weight: 600; transition: background 0.2s;">
        <div style="display: flex; flex-direction: column; margin-right: 0.5rem;">
          <span style="font-size: 0.75rem; opacity: 0.9;">Next:</span>
          <span>Testing & Benchmarks →</span>
        </div>
      </a>
    </div>
  </div>
  <div style="margin-top: 1rem; text-align: center; color: #586069; font-size: 0.875rem;">
    <span>📚 doppel Documentation • Reflection Engine</span>
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

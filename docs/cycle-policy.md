# Cycle & Sharing Policy

doppel provides three configurable policies for handling cyclic and shared pointer references during deep copy. Each policy trades off faithfulness, safety, and usability for different use cases.

---

## The problem

Go data structures can form graphs, not just trees:

- **Shared references:** Two fields point to the same allocation. Should the clone preserve this sharing?
- **Cyclic references:** A pointer forms a loop (e.g., `A.Next = B`, `B.Next = A`). Deep copying must avoid infinite recursion.

doppel handles both cases through configurable `CyclePolicy` values.

---

## The three policies

### PreserveShared (default)

Shared pointer allocations in the original are **preserved as shared** in the clone. Cyclic graphs are **reproduced faithfully** — the clone has the same cycle topology.

```go
shared := &Node{Name: "shared"}
type Diamond struct {
    Left  *Node
    Right *Node
}

src := Diamond{Left: shared, Right: shared}
cloned := doppel.MustCloneDeep(src, nil)

// clone.Left == clone.Right  (sharing preserved)
// clone.Left != src.Left     (different allocation)
```

**How it works:** The engine maintains a `visited` map from pointer address to the already-cloned `reflect.Value`. When it encounters a previously seen address, it returns the existing clone instead of recursing. This simultaneously handles cycles (back-edges return the in-progress clone) and shared references (both fields point to the same clone).

```go
// Self-referential cycle
n := &Node{ID: 1}
n.Next = n

cloned := doppel.MustCloneDeep(n, nil)
// cloned.ID == 1
// cloned.Next == cloned  (self-loop preserved)
// cloned != n             (different allocation)
```

**When to use:** General-purpose cloning. This is the safest default and matches the semantics most developers expect.

---

### BreakCycles

The first visit to a pointer is cloned normally. Any **back-edge** (a pointer to an address already on the current DFS call stack) is replaced with `nil` in the clone. This produces an **acyclic** clone of a cyclic graph.

Shared (non-cyclic) references are **not** deduplicated — each reference gets its own independent clone.

```go
n1 := &Node{ID: 1}
n2 := &Node{ID: 2}
n1.Next = n2
n2.Next = n1  // cycle: n1 → n2 → n1

eng := engine.NewWithOptions(nil, engine.Options{
    CyclePolicy: engine.BreakCycles,
})
clonedVal, _ := eng.Clone(reflect.ValueOf(n1))
cloned := clonedVal.Interface().(*Node)

// cloned.ID == 1
// cloned.Next.ID == 2
// cloned.Next.Next == nil  (cycle broken → nil)
```

**How it works:** The engine maintains an `inStack` map (DFS call stack). When it encounters an address on the current stack, it returns `nil`. After recursing, the address is removed from the stack so sibling references can still be cloned independently.

**When to use:**
- **Serialization:** When you need to serialize a potentially cyclic structure to a format that doesn't support cycles (JSON, protobuf).
- **Ayclic output:** When downstream code assumes a tree structure and would break on cycles.

**Trade-off:** Back-edges become `nil` in the clone. If you need those values, use `PreserveShared` instead.

---

### ErrorOnCycle

Returns a `*engine.CycleError` the moment a back-edge is detected. Non-cyclic shared references are cloned independently.

```go
eng := engine.NewWithOptions(nil, engine.Options{
    CyclePolicy: engine.ErrorOnCycle,
})

n := &Node{ID: 1}
n.Next = n  // self-loop

_, err := eng.Clone(reflect.ValueOf(n))
// err is *engine.CycleError
```

The `CycleError` contains the pointer address and type name for debugging:

```go
if cycleErr, ok := err.(*engine.CycleError); ok {
    fmt.Printf("Cycle at 0x%x (type %s)\n", cycleErr.Addr, cycleErr.TypeName)
}
```

**How it works:** Same stack tracking as `BreakCycles`, but returns an error instead of `nil` when a back-edge is detected.

**When to use:**
- **Strict validation:** When a cyclic graph represents a bug in your data model.
- **Assertion contexts:** When you want immediate, actionable feedback rather than silent behavior.

---

## Policy comparison

| Policy | Cycles | Shared refs | Back-edge behavior | Use case |
|--------|--------|-------------|-------------------|----------|
| `PreserveShared` | Reproduced faithfully | Deduplicated | Returns existing clone | General-purpose (default) |
| `BreakCycles` | Broken → `nil` | Independent (each cloned) | Returns `nil` | Serialization, acyclic output |
| `ErrorOnCycle` | Returns `*CycleError` | Independent (each cloned) | Returns error | Strict validation, assertion |

---

## Configuration

### Via `engine.NewWithOptions`

```go
eng := engine.NewWithOptions(reg, engine.Options{
    CyclePolicy: engine.BreakCycles,
})
cloned, err := eng.Clone(reflect.ValueOf(src))
```

### Via `doppel.CloneDeep`

`CloneDeep` creates an engine with `PreserveShared` by default. For non-default policies, create the engine directly:

```go
eng := engine.NewWithOptions(reg, engine.Options{
    CyclePolicy: engine.ErrorOnCycle,
})
clonedVal, err := eng.Clone(reflect.ValueOf(src))
cloned := clonedVal.Interface().(MyType)
```

---

## Shared references example

Shared references (non-cyclic) behave differently under each policy:

```go
shared := &Node{Name: "shared"}
src := Diamond{Left: shared, Right: shared}
```

| Policy | `clone.Left == clone.Right` | `clone.Left == src.Left` |
|--------|---------------------------|------------------------|
| `PreserveShared` | `true` (deduplicated) | `false` (new allocation) |
| `BreakCycles` | `false` (independent clones) | `false` (new allocation) |
| `ErrorOnCycle` | `false` (independent clones) | `false` (new allocation) |

---

## Self-referential maps

Maps can be self-referential (a map that contains itself as a value). The engine handles this under all policies:

- **PreserveShared:** The self-reference resolves to the same cloned map.
- **BreakCycles:** The self-reference is replaced with `nil`.
- **ErrorOnCycle:** Returns `*CycleError`.

---

## Backward compatibility

`PreserveShared` is the default policy and is fully backward-compatible with the Phase 4 engine. All existing code that doesn't specify a policy uses `PreserveShared` automatically.

---

## What's next?

- **[Error Handling](error-handling.md)** — Working with `CycleError`, `CloneError`, and `WrapError`.
- **[Patterns & Best Practices](patterns.md)** — Choosing the right policy for your use case.

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Cycle & Sharing Policy
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="struct-tags.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Struct Tags</span>
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
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"><a href="error-handling.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Error Handling</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
            </a></div>
    </div>
</div>
<!-- /Navigation -->


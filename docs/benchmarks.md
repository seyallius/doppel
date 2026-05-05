# Benchmarks

Performance data comparing doppel's manual clone helpers against shallow copy baselines. All benchmarks use Go's standard `testing.B` framework with `-benchmem` for allocation tracking.

*Hardware: Intel(R) Core(TM) i5-9500 CPU @ 3.00GHz, linux/amd64*

---

## How to run benchmarks

```bash
just bench

# Specific pattern
just bench "CloneSlice"

# Multiple runs for statistical stability
just bench-count "CloneSlice" 5
```

---

## Manual clone vs. shallow copy

### User (nested struct with pointer, slice, map)

| Benchmark          | ns/op | B/op | allocs/op |
|--------------------|-------|------|-----------|
| `ManualClone_User` | 359   | 528  | 6         |
| `ShallowCopy_User` | 2.75  | 0    | 0         |

The manual clone is roughly **130x slower** than a shallow copy, but produces a fully independent deep copy. For most applications, sub-microsecond clone times are negligible.

### Order (slice of structs + nested pointer)

| Benchmark           | ns/op | B/op | allocs/op |
|---------------------|-------|------|-----------|
| `ManualClone_Order` | 708   | 1136 | 11        |
| `ShallowCopy_Order` | 1.44  | 0    | 0         |

### Address (struct with nested pointer)

| Benchmark              | ns/op | B/op | allocs/op |
|------------------------|-------|------|-----------|
| `ManualClone_Address`  | 616   | 1056 | 12        |
| `ShallowCopy_Address`  | 1.33  | 0    | 0         |

### Large slices and maps

| Benchmark                                 | ns/op   | B/op    | allocs/op |
|-------------------------------------------|---------|---------|-----------|
| `ManualClone_UserLargeSlice` (1,000 tags) | 4,501   | 16,864  | 6         |
| `ManualClone_UserLargeMap` (500 entries)  | 1,690   | 1,256   | 8         |

> **Note**: `UserLargeMap` shows fewer allocations than expected because the map values are primitives (`int`), reducing per-entry overhead.

---

## Manual helpers vs. shallow copy

### Slices

| Benchmark                  | Size | ns/op  | B/op    | allocs/op |
|----------------------------|------|--------|---------|-----------|
| `CloneSlice_Strings_10`    | 10   | 74     | 160     | 1         |
| `CloneSlice_Strings_100`   | 100  | 447    | 1,792   | 1         |
| `CloneSlice_Strings_1000`  | 1000 | 3,918  | 16,384  | 1         |
| `ShallowCopy_Strings_1000` | 1000 | 2,215  | 16,384  | 1         |
| `CloneSlice_Ints_1000`     | 1000 | 3,632  | 8,192   | 1         |
| `ShallowCopy_Ints_1000`    | 1000 | 1,333  | 8,192   | 1         |

> **Insight**: Shallow copy of slices still allocates the new slice header and copies references/values, so it's not zero-cost. The benefit grows when element copying is cheap (e.g., `int` vs `string`).

### Maps

| Benchmark                   | Size | ns/op   | B/op    | allocs/op |
|-----------------------------|------|---------|---------|-----------|
| `CloneMap_StringInt_50`     | 50   | 1,737   | 1,880   | 4         |
| `CloneMap_StringInt_500`    | 500  | 5,021   | 6,616   | 4         |
| `ShallowCopy_StringInt_500` | 500  | 5,030   | 6,568   | 3         |

> **Insight**: For maps with primitive values, shallow copy performance is nearly identical to manual clone because both must iterate and re-insert entries. The advantage of shallow copy appears only when values are pointers or complex types where deep traversal is avoided.

### Pointers

| Benchmark                | ns/op | B/op | allocs/op |
|--------------------------|-------|------|-----------|
| `ClonePointer_Int`       | 14.01 | 8    | 1         |
| `ShallowPointerCopy_Int` | 1.27  | 0    | 0         |

Pointer cloning is extremely cheap — one `new(T)` allocation plus one function call. Shallow copy is faster still since it's just a reference assignment.

---

## Performance recommendations

| Scenario                          | Recommendation                                                         |
|-----------------------------------|------------------------------------------------------------------------|
| Hot path, type you own            | Manual `Clone()` + helpers                                             |
| Test fixtures                     | `MustClone`                                                            |
| Types cloned less than ~100k/sec  | Manual helpers are sufficient                                          |
| Large collections (10k+ elements) | Profile first; shallow copy may suffice if immutability isn't required |
| Maps with primitive values        | Shallow copy offers minimal gain; prefer clarity                       |

---

## Running your own benchmarks

```bash
just bench-save "Benchmark" 5 before.txt
# ... make changes ...
just bench-save "Benchmark" 5 after.txt
just benchstat-cmp before.txt after.txt
```

Use `benchstat` (from `golang.org/x/perf/cmd/benchstat`) for statistically significant comparisons.

---

## What's next?

- **[API Reference](api-reference.md)** — Complete function and type signatures.

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        doppel Documentation &bull; Benchmarks
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;">
            <a href="struct-tags.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">&#8592;</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Struct Tags</span>
                </span>
            </a>
        </div>
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: center; align-items: center;">
            <a href="INDEX.md" style="display: flex; align-items: center; justify-content: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #8b5cf6, #6d28d9); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(139, 92, 246, 0.3); text-align: center;">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">&#8962;</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Return to</span>
                    <span style="font-size: 1rem; font-weight: 600;">Index</span>
                </span>
            </a>
        </div>
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;">
            <a href="api-reference.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">API Reference</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">&#8594;</span>
            </a>
        </div>
    </div>
</div>
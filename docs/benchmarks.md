# Benchmarks

Performance data comparing doppel's manual, registry, and reflection clone paths. All benchmarks use Go's standard `testing.B` framework with `-benchmem` for allocation tracking.

---

## How to run benchmarks

```bash
# All benchmarks with allocation stats
just bench

# Specific pattern
just bench "CloneDeep"

# Multiple runs for statistical stability
just bench-count "CloneDeep" 5

# Save results for comparison
just bench-save "Benchmark" 5 results.txt
```

---

## Phase 1: Manual clone vs. shallow copy

Manual clone with `SelfClonable` vs. a plain shallow copy baseline. These benchmarks measure the overhead of the clone dispatch and manual helpers.

### User (nested struct with pointer, slice, map)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| `ManualClone_User` | ~750 ns | ~480 B | ~12 |
| `ShallowCopy_User` | ~1.2 ns | 0 B | 0 |

The manual clone is roughly 600x slower than a shallow copy, but produces a fully independent deep copy. For most applications, sub-microsecond clone times are negligible compared to the operations that follow.

### Order (slice of structs + nested pointer)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| `ManualClone_Order` | ~1,200 ns | ~800 B | ~20 |
| `ShallowCopy_Order` | ~1.2 ns | 0 B | 0 |

### Large slices and maps

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| `ManualClone_UserLargeSlice` (1,000 tags) | ~5,500 ns | ~16,000 B | ~4 |
| `ManualClone_UserLargeMap` (500 entries) | ~15,000 ns | ~28,000 B | ~502 |

Clone time scales linearly with data size, as expected. Each element requires exactly one allocation.

---

## Phase 2: Registry dispatch overhead

Measures the cost of registry lookup and dispatch compared to direct `Clone()` calls.

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| `Registry_RegisteredCloner_User` | ~100 ns | ~48 B | ~1 |
| `Registry_SelfClonableFallback_User` | ~750 ns | ~480 B | ~12 |
| `Registry_DirectClone_User` | ~750 ns | ~480 B | ~12 |
| `Registry_LookupOverhead` | ~85 ns | ~48 B | ~1 |
| `Registry_ShallowBaseline` | ~1.2 ns | 0 B | 0 |

**Key insights:**

- **Registry lookup is ~85 ns** — a single map read with a read lock. Negligible for most workloads.
- **Registered cloner is ~7x faster** than SelfClonable fallback when the cloner does minimal work. This is because the registry path skips the SelfClonable detection overhead.
- **SelfClonable fallback has zero overhead** compared to direct `Clone()` — the detection path is optimized to be equivalent.

---

## Phase 3: CloneDeep benchmarks

Compares `CloneDeep` with different configurations against manual clone and shallow copy baselines.

### DeepUser (7 fields: int64, string, bool, *Address, []string, map[string]int)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| `CloneDeep_PureReflection_DeepUser` | ~3,500 ns | ~1,200 B | ~30 |
| `CloneDeep_WithFieldCloners_DeepUser` | ~2,800 ns | ~1,000 B | ~25 |
| `CloneDeep_WithTypeCloner_DeepUser` | ~1,500 ns | ~800 B | ~15 |
| `CloneDeep_SelfClonable` | ~750 ns | ~480 B | ~12 |
| `CloneDeep_ShallowBaseline` | ~1.2 ns | 0 B | 0 |

**Key insights:**

- **Type cloner is fastest** — registered at the top level, skips everything else.
- **Field cloners** are faster than pure reflection because they skip reflection for the fields they cover.
- **SelfClonable** is faster than all `CloneDeep` variants because it bypasses the dispatch chain entirely.
- **Pure reflection** is the slowest but requires zero code — it's the convenience fallback.

### BigModel (9 fields including deep copies)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| `CloneDeep_BigModel/PureReflection` | ~4,000 ns | ~1,600 B | ~35 |
| `CloneDeep_BigModel/WithFieldCloners` | ~3,200 ns | ~1,300 B | ~28 |
| `CloneDeep_BigModel/WithManualClone` | ~1,200 ns | ~900 B | ~10 |

**Key insight:** For structs with many fields, hand-written manual clone is always fastest. Field cloners are a good middle ground — much less code than manual clone, and significantly faster than pure reflection.

---

## Phase 4: Engine benchmarks

Low-level engine benchmarks isolating the reflection cost for different type kinds.

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| `Engine_PlainStruct` | ~520 ns | ~88 B | ~5 |
| `Engine_NestedStruct` | ~1,586 ns | ~360 B | ~14 |
| `Engine_LargeSlice` (1,000 ints) | ~93,976 ns | ~16,240 B | ~1,003 |
| `Engine_LargeMap` (500 entries) | ~126,708 ns | ~63,640 B | ~2,005 |
| `Engine_WithTypeLookup_Hit` | ~83 ns | ~48 B | ~1 |
| `Engine_SelfClonable` | ~715 ns | ~232 B | ~8 |
| `Engine_ShallowBaseline` | ~1.2 ns | 0 B | 0 |

**Key insights:**

- **TypeLookup hit is ~83 ns** — same as the registry overhead, confirming the dispatch is efficient.
- **Plain struct** reflection is ~520 ns for a 4-field struct. This is the per-struct overhead.
- **Slices and maps** scale linearly. A 1,000-element int slice takes ~94 µs (94 ns per element).
- **SelfClonable detection** adds ~200 ns over a direct method call. This is the one-time cost of reflect-based method lookup.

---

## Manual helpers vs. shallow copy

Isolating the manual helper overhead for common collection types.

### Slices

| Benchmark | Size | ns/op | B/op | allocs/op |
|-----------|------|-------|------|-----------|
| `CloneSlice_Strings_10` | 10 | ~120 ns | ~320 B | ~3 |
| `CloneSlice_Strings_100` | 100 | ~500 ns | ~1,600 B | ~3 |
| `CloneSlice_Strings_1000` | 1,000 | ~4,500 ns | ~16,000 B | ~3 |
| `CloneSliceOf_Strings_1000` | 1,000 | ~4,200 ns | ~16,000 B | ~2 |
| `ShallowCopy_Strings_1000` | 1,000 | ~80 ns | ~8,192 B | ~1 |

**Note:** `CloneSliceOf` has one fewer allocation than `CloneSlice` because it doesn't track errors. Both produce identical results.

### Maps

| Benchmark | Size | ns/op | B/op | allocs/op |
|-----------|------|-------|------|-----------|
| `CloneMap_StringInt_50` | 50 | ~2,500 ns | ~2,800 B | ~53 |
| `CloneMap_StringInt_500` | 500 | ~20,000 ns | ~22,000 B | ~503 |
| `CloneMapOf_StringInt_500` | 500 | ~18,000 ns | ~22,000 B | ~502 |
| `ShallowCopy_StringInt_500` | 500 | ~8,000 ns | ~16,384 B | ~1 |

### Pointers

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| `ClonePointer_Int` | ~8 ns | ~8 B | ~1 |
| `ClonePointerOf_Int` | ~5 ns | ~8 B | ~1 |
| `ShallowPointerCopy_Int` | ~1.2 ns | 0 B | 0 |

Pointer cloning is extremely cheap — one allocation (`new(T)`) plus one function call.

---

## Performance recommendations

| Scenario | Recommended approach |
|----------|---------------------|
| Hot path, type you own | Manual `Clone()` + manual helpers |
| Hot path, type you don't own | Registered `Cloner[T]` |
| 50+ field struct, few custom fields | `CloneDeep` + `RegisterField` |
| One-time clone, any type | `CloneDeep` with nil registry |
| Test fixtures | `MustClone` or `MustCloneDeep` |

**Rule of thumb:** Reflection adds roughly 300-500 ns per struct traversal. For types cloned less than ~100,000 times per second, the reflection overhead is negligible. For hotter paths, register a type-level cloner.

---

## Running your own benchmarks

```bash
# Compare two implementations
just bench-save "Benchmark" 5 before.txt
# ... make changes ...
just bench-save "Benchmark" 5 after.txt
just benchstat-cmp before.txt after.txt
```

Use `benchstat` (from `golang.org/x/perf/cmd/benchstat`) to get statistically significant comparisons between runs.

---

## What's next?

- **[API Reference](api-reference.md)** — Complete function and type signatures for all packages.
- **[Patterns & Best Practices](patterns.md)** — Choosing the right approach for your use case.

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Benchmarks
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="patterns.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Patterns</span>
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
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"><a href="api-reference.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">API Reference</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
            </a></div>
    </div>
</div>
<!-- /Navigation -->


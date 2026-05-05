# 🧪 Testing & Benchmarks

> Run tests, measure performance, and interpret results like a pro. ✨

---

## Running Tests

### Basic Commands

```bash
# All tests
go test ./...

# With race detector (recommended for CI)
go test -race ./...

# Verbose output per package
go test -v ./...
go test -v ./manual/...

# Run specific test
go test -v -run TestCloneSlice ./manual/...
```

### Benchmark Commands

```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Benchmark a specific package
go test -bench=. -benchmem ./manual/...

# Filter benchmarks by name
go test -bench=BenchmarkCloneSlice -benchmem ./manual/...

# Compare with previous results using benchstat
just bench-save output=bench.txt
just benchstat file=bench.txt
```

### Using `just` Recipes (If Available)

```bash
# Save benchmarks
just bench-save output=bench.txt

# Compare with benchstat
just benchstat file=bench.txt

# Run CI-style tests
just ci
```

---

## Benchmark Results

Indicative results on **11th Gen Intel® Core™ i5-11400H @ 2.70GHz**.

### Doppel (Manual) vs Reflection Comparison

| Benchmark        | Doppel (ns/op) | Reflection (ns/op) | Speedup   | Doppel (B/op) | Reflection (B/op) | Doppel (allocs/op) | Reflection (allocs/op) |
|------------------|----------------|--------------------|-----------|---------------|-------------------|--------------------|------------------------|
| `Score`          | 22.03 ± 8%     | 131.8 ± 3%         | **~6×**   | 24            | 96                | 1                  | 4                      |
| `User`           | 309.9 ± 2%     | 1.193µ ± 4%        | **~4×**   | 528           | 968               | 6                  | 18                     |
| `Order`          | 615.9 ± 4%     | 2.183µ ± 4%        | **~3.5×** | 1.109Ki       | 1.742Ki           | 11                 | 34                     |
| `UserLargeSlice` | 8.363µ ± 3%    | 32.76µ ± 0%        | **~4×**   | 32.44Ki       | 32.87Ki           | 6                  | 18                     |
| `UserLargeMap`   | 29.26µ ± 0%    | 99.41µ ± 4%        | **~3.4×** | 53.64Ki       | 101.4Ki           | 10                 | 2,016                  |

### Engine-Specific Benchmarks

```
BenchmarkEngine_PlainStruct            2272365     520.8 ns/op     88 B/op     5 allocs/op
BenchmarkEngine_NestedStruct            736201    1586 ns/op    360 B/op    14 allocs/op
BenchmarkEngine_LargeSlice               12734   93976 ns/op  16240 B/op  1003 allocs/op
BenchmarkEngine_LargeMap                  8905  126708 ns/op  63640 B/op  2005 allocs/op
BenchmarkEngine_WithTypeLookup_Hit    14146173    83.50 ns/op     48 B/op     1 allocs/op  ← registry hit!
BenchmarkEngine_SelfClonable           1671700   715.6 ns/op    232 B/op     8 allocs/op
BenchmarkEngine_ShallowBaseline      975549486   1.229 ns/op      0 B/op     0 allocs/op
```

---

## Interpreting Benchmarks

### Key Metrics

| Metric      | What It Means                  | Why It Matters                |
|-------------|--------------------------------|-------------------------------|
| `ns/op`     | Nanoseconds per operation      | Lower = faster cloning        |
| `B/op`      | Bytes allocated per operation  | Lower = less GC pressure      |
| `allocs/op` | Heap allocations per operation | Lower = better cache locality |

### Reading the Results

1. **Speedup column**: Manual cloning is 3-6× faster than reflection
2. **Allocation columns**: Manual uses 40-95% fewer allocations
3. **Large collections**: The performance gap grows with complexity

### When to Benchmark

✅ Before optimizing a hot path  
✅ When choosing between manual and reflection for a type  
✅ After refactoring clone logic to ensure no regression  
✅ When adding new helper functions

### Benchmarking Tips

- Use `-benchmem` to see allocation stats
- Run benchmarks multiple times for stable results
- Compare on the same hardware/Go version
- Focus on realistic data sizes (not just micro-benchmarks)

> 💡 **Pro Tip**: Always benchmark manual vs reflection for your specific use case. The 3-6× speedup is indicative — your
> mileage may vary based on struct complexity and data size. (◕‿◕)✧

<!--

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Testing & Benchmarks
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="reflection-engine.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Reflection Engine</span>
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
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"><a href="roadmap.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Roadmap</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
            </a></div>
    </div>
</div>
<!-- /Navigation -->


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

<!-- Navigation -->
<div style="margin-top: 3rem; padding: 1.5rem; border-top: 2px solid #e1e4e8; border-radius: 6px; background: #f6f8fa;">
  <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 1rem;">
    <div style="flex: 1; min-width: 200px;">
      <a href="./reflection-engine.md" style="display: inline-flex; align-items: center; padding: 0.75rem 1rem; background: #0366d6; color: white; text-decoration: none; border-radius: 6px; font-weight: 600; transition: background 0.2s;">
        <span style="margin-right: 0.5rem;">←</span>
        <div style="display: flex; flex-direction: column;">
          <span style="font-size: 0.75rem; opacity: 0.9;">Previous:</span>
          <span>Reflection Engine</span>
        </div>
      </a>
    </div>
    <div style="flex: 1; min-width: 200px; text-align: right;">
      <a href="./roadmap.md" style="display: inline-flex; align-items: center; padding: 0.75rem 1rem; background: #28a745; color: white; text-decoration: none; border-radius: 6px; font-weight: 600; transition: background 0.2s;">
        <div style="display: flex; flex-direction: column; margin-right: 0.5rem;">
          <span style="font-size: 0.75rem; opacity: 0.9;">Next:</span>
          <span>Roadmap →</span>
        </div>
      </a>
    </div>
  </div>
  <div style="margin-top: 1rem; text-align: center; color: #586069; font-size: 0.875rem;">
    <span>📚 doppel Documentation • Testing & Benchmarks</span>
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

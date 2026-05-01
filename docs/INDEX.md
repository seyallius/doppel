# 🗺️ doppel Documentation Index

> Your central navigation hub for all `doppel` documentation. Jump to any section below! ✧◝(⁰▿⁰)◜✧

---

## 📖 Table of Contents

### 🎯 Getting Started

- [Why doppel?](./getting-started.md#why-doppel)
- [Installation](./getting-started.md#installation)
- [Quick Example](./getting-started.md#quick-example)
- [Go Version Requirements](./getting-started.md#go-version-requirements)

### 💭 Design Philosophy

- [Core Principles](./philosophy.md#core-principles)
- [Priority Chain](./philosophy.md#priority-chain)
- [When to Use What](./philosophy.md#when-to-use-what)

### 🧠 Core Concepts

- [`Cloner[T]` Interface](./core-concepts.md#clonert-interface)
- [`SelfClonable[T]` Interface](./core-concepts.md#selfclonablet-interface)
- [Identity Helpers](./core-concepts.md#identity-helpers)
- [Choosing Your Strategy](./core-concepts.md#choosing-your-strategy)

### 🔧 API Reference

- [Public Entry Points](./api-reference.md#public-entry-points)
    - [`Clone[T]`](./api-reference.md#doppelclone)
    - [`MustClone[T]`](./api-reference.md#doppelmustclone)
    - [`CloneWith[T]`](./api-reference.md#doppelclonewith)
    - [`MustCloneWith[T]`](./api-reference.md#doppelmustclonewith)
    - [`CloneWithRegistry[T]`](./api-reference.md#doppelclonewithregistry)
- [Manual Helpers](./api-reference.md#manual-helpers)
    - [`CloneSlice` / `CloneSliceOf`](./api-reference.md#manualcloneslice--clonesliceof)
    - [`CloneMap` / `CloneMapOf`](./api-reference.md#manualclonemap--clonemapof)
    - [`ClonePointer` / `ClonePointerOf`](./api-reference.md#manualclonepointer--clonepointerof)

### 🛠️ Usage Guide

- [Step 1: Primitives Only](./usage-guide.md#step-1--simple-struct-primitives-only)
- [Step 2: Pointer Fields](./usage-guide.md#step-2--struct-with-a-pointer-field)
- [Step 3: Slices & Maps](./usage-guide.md#step-3--struct-with-slices-and-maps)
- [Step 4: Nested Aggregates](./usage-guide.md#step-4--full-aggregate-with-nested-structs)
- [Step 5: External Cloner](./usage-guide.md#step-5--external-cloner-no-selfclonable)
- [Step 6: Conditional Cloning](./usage-guide.md#step-6--conditional--filtered-cloning)
- [Step 7: Cloner Registry](./usage-guide.md#step-7--cloner-registry-phase-2)
- [Step 8: Reflection Fallback](./usage-guide.md#step-8--reflection-fallback-with-cycle-policies-phase-45)

### ⚙️ Advanced Topics

- [Error Handling & Context](./advanced.md#error-handling)
- [Nil Safety Contract](./advanced.md#nil-safety-contract)
- [Struct Tag Directives](./advanced.md#struct-tag-directives)
- [Best Practices](./advanced.md#best-practices)

### 🔍 Reflection Engine (Fallback)

- [When Reflection is Used](./reflection-engine.md#when-reflection-is-used)
- [Engine API](./reflection-engine.md#engine-api)
- [Cycle Policies](./reflection-engine.md#cycle-policies)
- [Features & Limitations](./reflection-engine.md#features--limitations)
- [Performance Considerations](./reflection-engine.md#performance-considerations)

### 🧪 Testing & Benchmarks

- [Running Tests](./testing.md#running-tests)
- [Benchmark Commands](./testing.md#benchmark-commands)
- [Performance Results](./testing.md#benchmark-results)
- [Interpreting Benchmarks](./testing.md#interpreting-benchmarks)

### 🗓️ Roadmap

- [Phase Summary](./roadmap.md#phase-summary)
- [Phase 1: Manual Foundation ✅](./roadmap.md#phase-1--manual-deep-copy-foundation)
- [Phase 2: Cloner Registry ✅](./roadmap.md#phase-2--cloner-registry)
- [Phase 3: Field-Level Customization 🔜](./roadmap.md#phase-3--field-level-customization)
- [Phase 4: Reflection Fallback ✅](./roadmap.md#phase-4--reflection-fallback)
- [Phase 5: Cycle Detection ✅](./roadmap.md#phase-5--cycle-detection)
- [Future Phases](./roadmap.md#future-phases)

---

## 🔗 Quick Links

| Need to...               | Go to...                                                                          |
|--------------------------|-----------------------------------------------------------------------------------|
| Install doppel           | [Getting Started → Installation](./getting-started.md#installation)               |
| Clone a simple struct    | [Usage Guide → Step 1](./usage-guide.md#step-1--simple-struct-primitives-only)    |
| Handle pointer fields    | [Usage Guide → Step 2](./usage-guide.md#step-2--struct-with-a-pointer-field)      |
| Clone slices/maps        | [API Reference → Manual Helpers](./api-reference.md#manual-helpers)               |
| Use reflection fallback  | [Reflection Engine → When to Use](./reflection-engine.md#when-reflection-is-used) |
| Configure cycle handling | [Reflection Engine → Cycle Policies](./reflection-engine.md#cycle-policies)       |
| Run benchmarks           | [Testing → Benchmark Commands](./testing.md#benchmark-commands)                   |
| See what's next          | [Roadmap](./roadmap.md)                                                           |

<!-- Navigation -->
<div style="margin-top: 3rem; padding: 1.5rem; border-top: 2px solid #e1e4e8; border-radius: 6px; background: #f6f8fa;">
  <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 1rem;">
    <div style="flex: 1; min-width: 200px;">
      <a href="../README.md" style="display: inline-flex; align-items: center; padding: 0.75rem 1rem; background: #0366d6; color: white; text-decoration: none; border-radius: 6px; font-weight: 600; transition: background 0.2s;">
        <span style="margin-right: 0.5rem;">←</span>
        <div style="display: flex; flex-direction: column;">
          <span style="font-size: 0.75rem; opacity: 0.9;">Back to</span>
          <span>README.md</span>
        </div>
      </a>
    </div>
    <div style="flex: 1; min-width: 200px; text-align: right;">
      <a href="./getting-started.md" style="display: inline-flex; align-items: center; padding: 0.75rem 1rem; background: #28a745; color: white; text-decoration: none; border-radius: 6px; font-weight: 600; transition: background 0.2s;">
        <div style="display: flex; flex-direction: column; margin-right: 0.5rem;">
          <span style="font-size: 0.75rem; opacity: 0.9;">Next:</span>
          <span>Getting Started →</span>
        </div>
      </a>
    </div>
  </div>
  <div style="margin-top: 1rem; text-align: center; color: #586069; font-size: 0.875rem;">
    <span>📚 doppel Documentation</span>
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

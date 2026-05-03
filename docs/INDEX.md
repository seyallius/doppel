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

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Documentation Index
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="../README.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">README.md</span>
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
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"><a href="getting-started.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Getting Started</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
            </a></div>
    </div>
</div>
<!-- /Navigation -->


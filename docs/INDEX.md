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
<div class="doppel-nav-container"
     style="margin-top: 3rem; padding: 1.75rem; border-top: 1px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c); box-shadow: 0 8px 24px rgba(0,0,0,0.4);">
    <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 1rem;">
        <div style="flex: 1; min-width: 200px;">
            <a href="../README.md" class="doppel-nav-btn doppel-nav-prev">
            <span class="doppel-arrow">←</span>
            <div class="doppel-text">
                <span class="doppel-label">Previous</span>
                <span class="doppel-title">README.md</span>
            </div>
        </a>
        </div>
        <div style="flex: 1; min-width: 200px; text-align: right;">
            <a href="getting-started.md" class="doppel-nav-btn doppel-nav-next">
            <div class="doppel-text">
                <span class="doppel-label">Next</span>
                <span class="doppel-title">Getting Started</span>
            </div>
            <span class="doppel-arrow">→</span>
        </a>
        </div>
    </div>
    <div style="margin-top: 1.25rem; text-align: center; color: #94a3b8; font-size: 0.8rem; letter-spacing: 0.03em;">
        <span>📚 doppel Documentation • Documentation Index</span>
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

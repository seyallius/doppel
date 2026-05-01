# doppel Docs Sidebar

- **🏠 Home**
    - [README](../README.md)
    - [Documentation Index](./INDEX.md)

- **🎯 Getting Started**
    - [Why doppel?](./getting-started.md#why-doppel)
    - [Installation](./getting-started.md#installation)
    - [Quick Example](./getting-started.md#quick-example)

- **💭 Philosophy**
    - [Core Principles](./philosophy.md#core-principles)
    - [Priority Chain](./philosophy.md#priority-chain)

- **🧠 Core Concepts**
    - [`Cloner[T]`](./core-concepts.md#clonert-interface)
    - [`SelfClonable[T]`](./core-concepts.md#selfclonablet-interface)
    - [Identity Helpers](./core-concepts.md#identity-helpers)

- **🔧 API Reference**
    - [Public Entry Points](./api-reference.md#public-entry-points)
    - [Manual Helpers](./api-reference.md#manual-helpers)

- **🛠️ Usage Guide**
    - [Step 1: Primitives](./usage-guide.md#step-1--simple-struct-primitives-only)
    - [Step 2: Pointers](./usage-guide.md#step-2--struct-with-a-pointer-field)
    - [Step 3: Slices/Maps](./usage-guide.md#step-3--struct-with-slices-and-maps)
    - [Step 4: Nested](./usage-guide.md#step-4--full-aggregate-with-nested-structs)
    - [Step 5: External Cloner](./usage-guide.md#step-5--external-cloner-no-selfclonable)
    - [Step 6: Conditional](./usage-guide.md#step-6--conditional--filtered-cloning)
    - [Step 7: Registry](./usage-guide.md#step-7--cloner-registry-phase-2)
    - [Step 8: Reflection](./usage-guide.md#step-8--reflection-fallback-with-cycle-policies-phase-45)

- **⚙️ Advanced**
    - [Error Handling](./advanced.md#error-handling)
    - [Nil Safety](./advanced.md#nil-safety-contract)
    - [Struct Tags](./advanced.md#struct-tag-directives)

- **🔍 Reflection Engine**
    - [When to Use](./reflection-engine.md#when-reflection-is-used)
    - [Cycle Policies](./reflection-engine.md#cycle-policies)
    - [Performance](./reflection-engine.md#performance-considerations)

- **🧪 Testing**
    - [Commands](./testing.md#running-tests)
    - [Benchmarks](./testing.md#benchmark-results)

- **🗓️ Roadmap**
    - [Phase Summary](./roadmap.md#phase-summary)
    - [Next: Phase 3](./roadmap.md#phase-3--field-level-customization)
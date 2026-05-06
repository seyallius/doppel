# doppel Documentation

> "Your data's doppelgänger — deep copies without side effects."

doppel is a safe, explicit deep-cloning library for Go. It provides
zero-overhead cloning with no reflection, using composable generic helpers and
an optional code generator.

## Architecture

doppel is built around three layers:

1. **Core interfaces** (`core` package) — defines `Cloner[T]`, `SelfClonable[T]`,
   struct tag types, and error types. These are pure data structures with no
   runtime reflection.

2. **Manual helpers** (`manual` package) — generic functions like
   `CloneSlice`, `CloneMap`, and `ClonePointer` that you compose inside your
   type's own `Clone()` method. Everything is explicit and auditable.

3. **Code generator** (`doppelgen` CLI) — reads Go source files with `doppel`
   struct tags and emits `Clone()` method implementations automatically. This
   eliminates manual boilerplate while still producing readable, no-reflection
   code. See [Getting Started — Code Generator](getting-started.md#code-generator).

## Design principles

- **No reflection at runtime** — cloning uses direct function calls and
  generics only.
- **Explicit over magic** — every clone path is visible and auditable in your
  source.
- **Composable** — `CloneSlice`, `CloneMap`, and `ClonePointer` are building
  blocks you wire together.
- **Error-aware** — every helper returns `(T, error)` so cloning failures are
  never silently swallowed.

## Navigation

| Page                                                | Description                                                        |
|-----------------------------------------------------|--------------------------------------------------------------------|
| [Getting Started](getting-started.md)               | Installation, first clone, and code generator setup.               |
| [Struct Tags](struct-tags.md)                       | All `doppel:"..."` tag directives and their effects.               |
| [API Reference](api-reference.md)                   | Complete reference for all public types, constants, and functions. |
| [Manual Helpers](manual-helpers.md)                 | `manual.CloneSlice`, `CloneMap`, `ClonePointer`, and `Identity`.   |
| [Code Generator](getting-started.md#code-generator) | `doppelgen` CLI usage, flags, and `go:generate` integration.       |

---

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Documentation
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"></div>
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


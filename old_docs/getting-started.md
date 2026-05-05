# 🎯 Getting Started with doppel

> Quick onboarding: why doppel exists, how to install it, and your first clone. ✨

---

## Why doppel?

Most deep-copy libraries in Go use `reflect` as their primary engine. Reflection works for any type automatically, but
it comes with real costs:

| Issue           | Impact                                                                                       |
|-----------------|----------------------------------------------------------------------------------------------|
| **Performance** | Reflection bypasses compiler optimizations, paying allocation overhead on every field access |
| **Opacity**     | Silent skipping of unexported fields, mishandled interfaces, unexpected shared references    |
| **No Control**  | Can't conditionally clone fields or mix shallow/deep strategies per-field                    |

`doppel` inverts the priority order:

| Priority | Strategy                                           | When Used                                       |
|----------|----------------------------------------------------|-------------------------------------------------|
| 1        | **Manual clone** (your `Clone()` method)           | Always, by default — fastest path               |
| 2        | **External `Cloner[T]`** (via `CloneWith`)         | When clone logic needs injected context         |
| 3        | **Registry `Cloner[T]`** (via `CloneWithRegistry`) | Type-level override without modifying source    |
| 4        | **Reflection fallback** (`engine.Engine`)          | Last resort — only when none of the above exist |
By default, reflection is not used at all. Every copy decision is written explicitly by you. ✧◝(⁰▿⁰)◜✧
---

## Installation

```bash
go get github.com/seyallius/doppel
```

**Requirements**: Go 1.25 or later (for generic type inference and range-over-integer improvements).

---

## Quick Example

```go
// Package cmd. main.go - Quick start example for doppel library.
package main

import (
    "fmt"
    "github.com/seyallius/doppel"
    "github.com/seyallius/doppel/manual"
    "github.com/seyallius/doppel/core"
)

// User represents a simple domain entity.
type User struct {
    ID   int64
    Name string
    Tags []string
}

// Clone implements SelfClonable[*User] for explicit deep copying.
// Returns an independent copy with no shared references.
func (u *User) Clone() (*User, error) {
    if u == nil {
        return nil, nil
    }
    // Clone slice of primitives using Identity helper
    tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
    if err != nil {
        return nil, core.WrapError("User.Tags", err)
    }
    return &User{
        ID:   u.ID,
        Name: u.Name,
        Tags: tags, // independent slice
    }, nil
}

func main() {
    original := &User{
        ID:   1,
        Name: "Alice",
        Tags: []string{"admin", "dev"},
    }
    
    // Deep clone using doppel's public API
    cloned, err := doppel.Clone(original)
    if err != nil {
        panic(err)
    }
    
    // Verify independence
    cloned.Tags = append(cloned.Tags, "modified")
    fmt.Println("Original tags:", original.Tags) // [admin dev]
    fmt.Println("Cloned tags:  ", cloned.Tags)   // [admin dev modified] ✧
}
```

---

## Next Steps

- 🧠 Understand the [Core Concepts](./core-concepts.md)
- 🛠️ Follow the [Usage Guide](./usage-guide.md) step-by-step
- 🔍 Learn about the [Reflection Fallback](./reflection-engine.md) for edge cases

> 💡 **Remember**: Manual cloning is always the default. Reflection is a controlled fallback — never the first choice. (
> ◕‿◕)

<!--

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Getting Started
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="INDEX.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">INDEX.md</span>
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
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"><a href="philosophy.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Philosophy</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
            </a></div>
    </div>
</div>
<!-- /Navigation -->


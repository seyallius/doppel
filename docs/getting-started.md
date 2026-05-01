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

In Phase 1, reflection is not present at all. Every copy decision is written explicitly by you. ✧◝(⁰▿⁰)◜✧

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

<!-- Navigation -->
<div style="margin-top: 3rem; padding: 1.5rem; border-top: 2px solid #e1e4e8; border-radius: 6px; background: #f6f8fa;">
  <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 1rem;">
    <div style="flex: 1; min-width: 200px;">
      <a href="./INDEX.md" style="display: inline-flex; align-items: center; padding: 0.75rem 1rem; background: #0366d6; color: white; text-decoration: none; border-radius: 6px; font-weight: 600; transition: background 0.2s;">
        <span style="margin-right: 0.5rem;">←</span>
        <div style="display: flex; flex-direction: column;">
          <span style="font-size: 0.75rem; opacity: 0.9;">Previous:</span>
          <span>INDEX.md</span>
        </div>
      </a>
    </div>
    <div style="flex: 1; min-width: 200px; text-align: right;">
      <a href="./philosophy.md" style="display: inline-flex; align-items: center; padding: 0.75rem 1rem; background: #28a745; color: white; text-decoration: none; border-radius: 6px; font-weight: 600; transition: background 0.2s;">
        <div style="display: flex; flex-direction: column; margin-right: 0.5rem;">
          <span style="font-size: 0.75rem; opacity: 0.9;">Next:</span>
          <span>Philosophy →</span>
        </div>
      </a>
    </div>
  </div>
  <div style="margin-top: 1rem; text-align: center; color: #586069; font-size: 0.875rem;">
    <span>📚 doppel Documentation • Getting Started</span>
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

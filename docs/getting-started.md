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

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->
<div class="doppel-nav-container"
     style="margin-top: 3rem; padding: 1.75rem; border-top: 1px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c); box-shadow: 0 8px 24px rgba(0,0,0,0.4);">
    <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 1rem;">
        <div style="flex: 1; min-width: 200px;">
            <a href="INDEX.md" class="doppel-nav-btn doppel-nav-prev">
            <span class="doppel-arrow">←</span>
            <div class="doppel-text">
                <span class="doppel-label">Previous</span>
                <span class="doppel-title">INDEX.md</span>
            </div>
        </a>
        </div>
        <div style="flex: 1; min-width: 200px; text-align: right;">
            <a href="philosophy.md" class="doppel-nav-btn doppel-nav-next">
            <div class="doppel-text">
                <span class="doppel-label">Next</span>
                <span class="doppel-title">Philosophy</span>
            </div>
            <span class="doppel-arrow">→</span>
        </a>
        </div>
    </div>
    <div style="margin-top: 1.25rem; text-align: center; color: #94a3b8; font-size: 0.8rem; letter-spacing: 0.03em;">
        <span>📚 doppel Documentation • Getting Started</span>
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

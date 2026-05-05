# Getting Started

Get doppel installed and write your first deep clone in under two minutes.

---

## Installation

```bash
go get github.com/seyallius/doppel
```

doppel requires **Go 1.25** or later. There are zero external dependencies — only the Go standard library is used.

---

## Your first clone

The simplest way to use doppel is to implement the `core.SelfClonable[T]` interface on your type. This means adding a single method — `Clone() (T, error)` — that returns an independent deep copy.

### Step 1: Define your type

```go
package main

type User struct {
    ID     int64
    Name   string
    Active bool
    Tags   []string
    Scores map[string]int
}
```

### Step 2: Implement `Clone()`

Use doppel's manual helpers to clone each field. For primitive fields (`int64`, `string`, `bool`), assignment is already a deep copy — no helper needed.

```go
import (
    "github.com/seyallius/doppel/core"
    "github.com/seyallius/doppel/manual"
)

func (u *User) Clone() (*User, error) {
    if u == nil {
        return nil, nil
    }

    // Clone the slice — Identity[string] means "strings don't need deep copy"
    tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
    if err != nil {
        return nil, core.WrapError("User.Tags", err)
    }

    // Clone the map — Identity[int] means "ints don't need deep copy"
    scores, err := manual.CloneMap(u.Scores, manual.Identity[int])
    if err != nil {
        return nil, core.WrapError("User.Scores", err)
    }

    // Return a new User with all fields independently copied
    return &User{
        ID:     u.ID,
        Name:   u.Name,
        Active: u.Active,
        Tags:    tags,
        Scores: scores,
    }, nil
}
```

### Step 3: Call `doppel.Clone`

```go
import "github.com/seyallius/doppel"

func main() {
    original := &User{
        ID:     1,
        Name:   "Alice",
        Active: true,
        Tags:   []string{"admin", "editor"},
        Scores: map[string]int{"math": 95, "english": 88},
    }

    cloned, err := doppel.Clone(original)
    if err != nil {
        panic(err)
    }

    // Mutate the original — the clone is unaffected
    original.Tags[0] = "mutated"
    original.Scores["math"] = 0

    fmt.Println(cloned.Tags[0])    // "admin" (not "mutated")
    fmt.Println(cloned.Scores["math"]) // 95 (not 0)
}
```

---

## Quick reference: choosing the right helper

| You have...                      | Use                                               |
|----------------------------------|----------------------------------------------------|
| A pointer field                   | `manual.ClonePointer(u.Addr, cloneAddr)`         |
| A slice of primitives            | `manual.CloneSlice(u.Tags, manual.Identity[string])` |
| A slice of structs               | `manual.CloneSlice(u.Items, item.Clone)`           |
| A map with primitive values       | `manual.CloneMap(u.Scores, manual.Identity[int])`   |
| A map with struct values          | `manual.CloneMap(u.Lookup, cloneValue)`           |
| A primitive field                 | Direct assignment (no helper needed)               |

---

## MustClone

`MustClone` panics instead of returning an error. Use this in tests and initialization code where a cloning failure is always a programming error:

```go
cloned := doppel.MustClone(original)
```

---

## Struct tags (future generator)

You can annotate struct fields with `doppel:"..."` tags. These are currently informational only — a future code generator will read them to automatically emit `Clone()` implementations:

```go
type User struct {
    Name    string
    Secret  string           `doppel:"-"`       // skip in clone
    Config  map[string]string `doppel:"readonly"` // shared (not deep-copied)
    Address *Address         `doppel:"clone"`    // custom clone logic
    Tags    []string         `doppel:"deep"`     // explicit deep copy (default)
}
```

See [Struct Tags](struct-tags.md) for the full reference.

---

## What's next?

- **[SelfClonable Interface](self-clonable.md)** — Deep dive into the `Clone()` method pattern, including nested structs and pointer fields.
- **[Manual Helpers](manual-helpers.md)** — Detailed reference for `CloneSlice`, `CloneMap`, `ClonePointer`, and `Identity`.

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        doppel Documentation &bull; Getting Started
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;">
            <a href="INDEX.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">&#8592;</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Home</span>
                </span>
            </a>
        </div>
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: center; align-items: center;">
            <a href="INDEX.md" style="display: flex; align-items: center; justify-content: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #8b5cf6, #6d28d9); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(139, 92, 246, 0.3); text-align: center;">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">&#8962;</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Return to</span>
                    <span style="font-size: 1rem; font-weight: 600;">Index</span>
                </span>
            </a>
        </div>
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;">
            <a href="self-clonable.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">SelfClonable</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">&#8594;</span>
            </a>
        </div>
    </div>
</div>

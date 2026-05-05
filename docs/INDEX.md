# doppel

**Your data's doppelgänger — deep copies without side effects.**

---

## What is doppel?

doppel is a Go library for safe, explicit deep cloning of complex data structures. It provides a minimal, zero-reflection API built around composable generic helpers that you wire together inside your type's `Clone()` method.

Go assignment is a shallow copy. Structs with pointer fields, slices, and maps silently share memory between "originals" and "copies," leading to subtle bugs. doppel solves this by giving you full control over every field, every allocation, and every edge case.

## Quick start

```go
type User struct {
    ID     int64
    Name   string
    Tags   []string
    Scores map[string]int
}

func (u *User) Clone() (*User, error) {
    if u == nil { return nil, nil }

    tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
    if err != nil {
        return nil, core.WrapError("User.Tags", err)
    }

    scores, err := manual.CloneMap(u.Scores, manual.Identity[int])
    if err != nil {
        return nil, core.WrapError("User.Scores", err)
    }

    return &User{ID: u.ID, Name: u.Name, Tags: tags, Scores: scores}, nil
}

cloned, err := doppel.Clone(original)
```

## Architecture

```
core          — Cloner[T], SelfClonable[T], FuncCloner[T], CloneError, tag contract
manual        — CloneSlice[T], CloneMap[K,V], ClonePointer[T], Identity[T], IdentityValue[T]
doppel        — Clone[T], MustClone[T]
```

## Navigation

| #  | Page                                      | Topic                              |
|----|-------------------------------------------|-------------------------------------|
| 1  | [Getting Started](getting-started.md) | Install, first clone, API guide   |
| 2  | [SelfClonable](self-clonable.md)     | The `Clone()` method pattern       |
| 3  | [Manual Helpers](manual-helpers.md)  | CloneSlice, CloneMap, ClonePointer |
| 4  | [Struct Tags](struct-tags.md)        | Tag directives for future generator|
| 5  | [Benchmarks](benchmarks.md)         | Performance data                   |
| 6  | [API Reference](api-reference.md)   | Complete function signatures       |

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        doppel Documentation
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"></div>
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
            <a href="getting-started.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Getting Started</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">&#8594;</span>
            </a>
        </div>
    </div>
</div>

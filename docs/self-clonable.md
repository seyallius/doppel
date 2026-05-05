# SelfClonable Interface

The `core.SelfClonable[T]` interface is the primary contract for types that manage their own deep-copy logic. When a type implements this interface, `doppel.Clone` dispatches directly to its `Clone()` method — no reflection, no magic, maximum speed.

---

## The interface

```go
type SelfClonable[T any] interface {
    Clone() (T, error)
}
```

The generic parameter `T` is the type being cloned. For pointer receivers, `T` is the pointer type:

```go
func (u *User) Clone() (*User, error) { ... }
```

---

## When to use SelfClonable

Choose `SelfClonable[T]` when:

- **The type owns all the state it needs to copy.** The `Clone()` method can reach every field directly.
- **You want the fastest possible clone path.** The compiler knows the exact type at call time.

---

## Basic implementation

### Value types (all primitive fields)

For a struct with only primitive fields, the clone is a plain struct literal. No helpers are needed:

```go
type Score struct {
    Label string
    Value float64
}

func (s Score) Clone() (Score, error) {
    return Score{Label: s.Label, Value: s.Value}, nil
}
```

### Nested structs with pointers

When your struct contains pointer fields, slices, or maps, use the `manual` package helpers to create independent copies:

```go
func cloneAddress(src Address) (Address, error) {
    return Address{
        Street: src.Street,
        City:   src.City,
        State:  src.State,
        Zip:    src.Zip,
    }, nil
}

func (u *User) Clone() (*User, error) {
    if u == nil {
        return nil, nil
    }

    addr, err := manual.ClonePointer(u.Address, cloneAddress)
    if err != nil {
        return nil, core.WrapError("User.Address", err)
    }

    tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
    if err != nil {
        return nil, core.WrapError("User.Tags", err)
    }

    return &User{ID: u.ID, Name: u.Name, Address: addr, Tags: tags}, nil
}
```

---

## Nil handling

Always handle nil at the top of `Clone()`. This is consistent across the library — all helpers return `nil, nil` when given nil input:

```go
func (u *User) Clone() (*User, error) {
    if u == nil {
        return nil, nil
    }
    // ...
}
```

---

## Slice of structs

When cloning a slice of structs that also implement `Clone()`, pass a transformation function to `CloneSlice`:

```go
items, err := manual.CloneSlice(o.Items, func(s Score) (Score, error) {
    return s.Clone()
})
```

---

## Independence guarantee

Every clone must be fully independent from the source. Mutating the original must never affect the clone:

```go
original := newUser()
cloned, _ := doppel.Clone(original)

original.Name = "mutated"
// cloned.Name still has the original value
```

---

## What's next?

- **[Manual Helpers](manual-helpers.md)** — Detailed reference for `CloneSlice`, `CloneMap`, `ClonePointer`, and `Identity`.

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        doppel Documentation &bull; SelfClonable Interface
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;">
            <a href="getting-started.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">&#8592;</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Getting Started</span>
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
            <a href="manual-helpers.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Manual Helpers</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">&#8594;</span>
            </a>
        </div>
    </div>
</div>

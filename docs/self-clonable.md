# SelfClonable Interface

The `core.SelfClonable[T]` interface is the primary contract for types that manage their own deep-copy logic. When a type implements this interface, doppel dispatches directly to its `Clone()` method — no reflection, no registry lookup, maximum speed.

---

## The interface

```go
// SelfClonable is an optional interface a type can implement so that
// doppel.Clone can dispatch directly to it without an external Cloner.
type SelfClonable[T any] interface {
    Clone() (T, error)
}
```

The generic parameter `T` is the type being cloned. For pointer receivers, `T` is the pointer type:

```go
// *User implements SelfClonable[*User]
func (u *User) Clone() (*User, error) { ... }
```

For value receivers, `T` is the value type:

```go
// Score implements SelfClonable[Score]
func (s Score) Clone() (Score, error) { ... }
```

---

## When to use SelfClonable

Choose `SelfClonable[T]` when:

- **The type owns all the state it needs to copy.** The `Clone()` method can reach every field directly.
- **You want the fastest possible clone path.** The compiler knows the exact type at call time — no interface dispatch, no reflection.
- **You control the type definition.** You can add a `Clone()` method to the struct.

Choose `Cloner[T]` (external) when:

- **You don't own the type** and can't add methods to it (e.g., a type from a third-party package).
- **Cloning requires external context** (e.g., re-fetching a lazy-loaded field from a database).

---

## Basic implementation

### Value types (all primitive fields)

For a struct with only primitive fields, the clone is a plain struct literal. No helpers are needed because Go assignment for primitives is already a complete deep copy:

```go
type Score struct {
    Label string
    Value float64
}

func (s Score) Clone() (Score, error) {
    return Score{Label: s.Label, Value: s.Value}, nil
}

// Usage
cloned, _ := doppel.Clone(score)
```

### Nested structs with pointers

When your struct contains pointer fields, slices, or maps, use the `manual` package helpers to create independent copies:

```go
type User struct {
    ID      int64
    Name    string
    Active  bool
    Contact ContactInfo
    Tags    []string
    Scores  map[string]int
}

type ContactInfo struct {
    Email   string
    Phone   string
    Address *Address
}

type Address struct {
    Street string
    City   string
    State  string
    Zip    string
}
```

Clone from the innermost type outward:

```go
func cloneAddress(src Address) (Address, error) {
    return Address{
        Street: src.Street,
        City:   src.City,
        State:  src.State,
        Zip:    src.Zip,
    }, nil
}

func cloneContactInfo(src ContactInfo) (ContactInfo, error) {
    // ClonePointer wraps cloneAddress to handle nil safely
    clonedAddr, err := manual.ClonePointer(src.Address, cloneAddress)
    if err != nil {
        return ContactInfo{}, core.WrapError("ContactInfo.Address", err)
    }
    return ContactInfo{
        Email:   src.Email,
        Phone:   src.Phone,
        Address: clonedAddr,
    }, nil
}

func (u *User) Clone() (*User, error) {
    if u == nil {
        return nil, nil
    }

    contact, err := cloneContactInfo(u.Contact)
    if err != nil {
        return nil, core.WrapError("User.Contact", err)
    }

    tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
    if err != nil {
        return nil, core.WrapError("User.Tags", err)
    }

    scores, err := manual.CloneMap(u.Scores, manual.Identity[int])
    if err != nil {
        return nil, core.WrapError("User.Scores", err)
    }

    return &User{
        ID:      u.ID,
        Name:    u.Name,
        Active:  u.Active,
        Contact: contact,
        Tags:    tags,
        Scores:  scores,
    }, nil
}
```

---

## Nil handling

Always handle nil at the top of `Clone()`. This is consistent across the library — all cloners return `nil, nil` when given nil input:

```go
func (u *User) Clone() (*User, error) {
    if u == nil {
        return nil, nil
    }
    // ... clone logic
}
```

The `manual.ClonePointer`, `manual.CloneSlice`, and `manual.CloneMap` helpers all follow this same convention: nil input produces nil output with no error.

---

## Slice of structs

When cloning a slice of structs that also implement `SelfClonable`, pass a transformation function to `CloneSlice`:

```go
type Order struct {
    ID       string
    Customer *User
    Items    []Score
}

func (o *Order) Clone() (*Order, error) {
    if o == nil {
        return nil, nil
    }

    // Clone the pointer to *User
    customer, err := manual.ClonePointer(o.Customer, func(u User) (User, error) {
        cloned, err := u.Clone() // calls *User.Clone()
        if err != nil || cloned == nil {
            return User{}, err
        }
        return *cloned, nil
    })
    if err != nil {
        return nil, core.WrapError("Order.Customer", err)
    }

    // Clone the slice of Score structs
    items, err := manual.CloneSlice(o.Items, func(s Score) (Score, error) {
        return s.Clone() // calls Score.Clone()
    })
    if err != nil {
        return nil, core.WrapError("Order.Items", err)
    }

    return &Order{
        ID:       o.ID,
        Customer: customer,
        Items:    items,
    }, nil
}
```

---

## Independence guarantee

Every clone produced by a `SelfClonable` implementation must be fully independent from the source. This means mutating the original must never affect the clone, and vice versa:

```go
original := newUser()
cloned, _ := doppel.Clone(original)

// These mutations must NOT affect the clone:
original.Name = "mutated"
original.Tags[0] = "mutated"
original.Contact.Address.City = "mutated"
original.Scores["math"] = 0

// clone still has the original values
```

---

## Dispatch priority

When you call `doppel.CloneDeep` (the full priority chain), `SelfClonable` is checked **after** the registry but **before** the reflection engine:

```
Registered Cloner[T]  →  SelfClonable[T]  →  Reflection Engine
```

This means:
- If you register a `Cloner[T]` for the same type, the registry wins.
- If neither a registry entry nor `SelfClonable` exists, the reflection engine handles it automatically.

---

## What's next?

- **[Manual Helpers](manual-helpers.md)** — Detailed reference for `CloneSlice`, `CloneMap`, `ClonePointer`, and `Identity`.
- **[Cloner Registry](registry.md)** — Register external cloners for types you don't own.

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • SelfClonable Interface
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="getting-started.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Getting Started</span>
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
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"><a href="manual-helpers.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Manual Helpers</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
            </a></div>
    </div>
</div>
<!-- /Navigation -->


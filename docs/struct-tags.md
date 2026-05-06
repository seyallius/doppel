# Struct Tags

doppel uses the `doppel` struct tag key to control how individual fields are
handled during cloning. Tags are consumed by the code generator
([`doppelgen`](getting-started.md#code-generator)) to emit the correct
`Clone()` implementation for each field.

## Available directives

| Tag                | Effect                                                        |
|--------------------|---------------------------------------------------------------|
| *(no tag)*         | Deep copy (default behavior).                                 |
| `doppel:"-"`       | Skip the field entirely; the clone receives the zero value.   |
| `doppel:"shallow"` | Assign without recursing; the clone shares the field's value. |
| `doppel:"deep"`    | Explicit deep copy (same as default).                         |
| `doppel:"clone"`   | Emit a deep clone using a convention-based clone function.    |
| `doppel:"empty"`   | Produce empty-but-non-nil for collections and pointers.       |

> **Tip:** `doppelgen` parses these tags at code-generation time — no
> reflection is used at runtime.

## Usage

```go
type User struct {
    Name   string             // no tag → deep copy (primitive, direct assignment)
    Secret string             `doppel:"-"`          // skipped in clone
    Config map[string]string  `doppel:"shallow"`    // shared reference
    Address *Address          `doppel:"clone"`      // custom clone function
    Tags   []string           `doppel:"deep"`       // explicit deep copy
    Cache  []string           `doppel:"empty"`      // empty-but-non-nil slice
}
```

## `doppel:"-"` — Skip

The field is excluded from the clone entirely. The cloned struct receives the
zero value for that field:

- `string` → `""`
- `int` → `0`
- `*T` / `[]T` / `map[K]V` → `nil`

Use this for sensitive data (passwords, tokens) or fields that should not be
propagated to the clone.

## `doppel:"shallow"` — Shallow copy

The field is copied by direct assignment. The clone **shares** the original's
backing data. This is useful for large, immutable values where a deep copy
would be wasteful.

```go
type Config struct {
    StaticData map[string]string `doppel:"shallow"` // shared, never mutated
}
```

> **Warning:** If the shared data is later mutated, both the original and the
> clone will see the changes.

## `doppel:"deep"` — Deep copy

A full recursive deep copy is performed. This is the default when no tag is
specified, but you can make it explicit for clarity:

```go
Tags []string `doppel:"deep"`
```

The code generator emits calls to `manual.CloneSlice`, `manual.CloneMap`, or
`manual.ClonePointer` depending on the field type.

## `doppel:"clone"` — Custom clone function

Instructs the code generator to call a **convention-named** function instead
of using a built-in helper. The function must be named
`clone<StructName><FieldName>` and be visible in the same package:

```go
type User struct {
    Profile *Profile `doppel:"clone"`
}

// Convention name: clone + struct name + field name
func cloneUserProfile(src *Profile) (*Profile, error) {
    // custom clone logic here
}
```

## `doppel:"empty"` — Empty-but-non-nil

Produces an empty-but-non-nil value for collections and pointer types. This is
useful when you need a clone with initialized but empty containers:

| Source type | Clone value | Effect                       |
|-------------|-------------|------------------------------|
| `[]T`       | `[]T{}`     | Non-nil empty slice          |
| `map[K]V`   | `map[K]V{}` | Non-nil empty map            |
| `*Struct`   | `&Struct{}` | Pointer to zero-value struct |

For **primitive types**, `empty` is silently ignored — the value is assigned
normally because primitive assignment is already non-nil:

```go
type Cache struct {
    Data     map[string]string  `doppel:"empty"` // → map[string]string{}
    Entries  []Entry            `doppel:"empty"` // → []Entry{}
    Default  *Config            `doppel:"empty"` // → &Config{}
    Count    int                `doppel:"empty"` // → direct assignment (ignored)
}
```

## Using tags with `doppelgen`

When you run the code generator, it reads the struct tags and emits the
appropriate `Clone()` method. See [Getting Started](getting-started.md) for
how to run `doppelgen`.


<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Struct Tags
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="manual-helpers.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Manual Helpers</span>
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
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"><a href="benchmarks.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Benchmarks</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
            </a></div>
    </div>
</div>
<!-- /Navigation -->


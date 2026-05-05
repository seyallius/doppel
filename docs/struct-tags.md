# Struct Tags

The `doppel:"..."` struct tag system provides declarative annotations for fields. Currently these tags serve as **documentation and future code generator input** — no runtime reflection parsing occurs.

The generator (planned for a future release) will read these tags and automatically emit `SelfClonable[T].Clone()` implementations.

---

## Available directives

| Tag              | Behavior                                    |
|------------------|----------------------------------------------|
| `doppel:"-"`      | Skip the field; clone receives zero value.  |
| `doppel:"shallow"` | Shared reference; clone shares the value.   |
| `doppel:"readonly"` | Same as shallow; semantic alias.             |
| `doppel:"clone"`   | Custom clone logic (generator-specific).    |
| `doppel:"deep"`    | Full deep copy (default behavior).            |

---

## Usage

```go
type User struct {
    Name    string
    Secret  string           `doppel:"-"`       // excluded from clone
    Config  map[string]string `doppel:"readonly"` // shared with original
    Address *Address         `doppel:"clone"`    // generator emits custom code
    Tags    []string         `doppel:"deep"`     // explicit deep copy
}
```

---

## Tag semantics

### `doppel:"-"` — Skip

The field is excluded from the clone. The cloned struct will have the zero value for that field:

```go
type Config struct {
    Name   string `doppel:"-"`
    Secret string `doppel:"-"`
}

// cloned.Secret == ""
```

Use this for credentials, internal caches, or computed values that should not appear in clones.

### `doppel:"shallow"` / `doppel:"readonly"` — Shared reference

The field value is assigned without deep copying. The clone and original share the same underlying data:

```go
// original.Config["port"] = "9090" → also visible in cloned.Config
```

Use this for large, immutable data or config maps that will never be modified after initialization.

### `doppel:"clone"` — Custom clone logic

Tells the generator to emit custom clone code for this field. This is useful for fields with complex clone requirements that go beyond simple deep copy:

```go
type SecureUser struct {
    Password *Credentials `doppel:"clone"`
}
```

### `doppel:"deep"` — Explicit deep copy

Same behavior as the default (no tag). The field is recursively deep-copied. Use this tag to document intent explicitly:

```go
type CachedResponse struct {
    Data   []byte `doppel:"shallow"` // shared
    Meta   *Meta  `doppel:"deep"`    // explicit deep copy
}
```

---

## Tag parsing (for generators)

The `core.ParseTag` function provides a pure, zero-reflection parser that generators can use:

```go
directive := core.ParseTag("shallow")
// directive.Shallow == true
// directive.Deep == false

directive = core.ParseTag("-")
// directive.Skip == true
```

See [API Reference](api-reference.md) for the full `TagDirective` struct.

---

## What's next?

- **[Benchmarks](benchmarks.md)** — Performance data for manual clone helpers.

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        doppel Documentation &bull; Struct Tags
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;">
            <a href="manual-helpers.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">&#8592;</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Manual Helpers</span>
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
            <a href="benchmarks.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Benchmarks</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">&#8594;</span>
            </a>
        </div>
    </div>
</div>

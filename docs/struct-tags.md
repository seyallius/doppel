# Struct Tags

The reflection engine respects `doppel:"..."` struct tags to control per-field clone behavior. Tags provide a declarative, zero-code way to skip fields, share backing data, or require field cloners.

---

## Available directives

| Tag | Behavior |
|-----|----------|
| `doppel:"-"` | Skip the field entirely. The clone receives the zero value for that field. |
| `doppel:"shallow"` | Shallow copy — assign without recursing. The clone shares the field's value with the original. |
| `doppel:"readonly"` | Same behavior as `shallow`. Communicates that the field is conceptually immutable and safe to share. |
| `doppel:"clone"` | Require a registered field `Cloner`. Errors at runtime if no field cloner is registered. |
| `doppel:"deep"` | Explicit deep copy marker. Same behavior as the default (no tag), but documents the intent. |

---

## Priority within cloneStruct

Struct tags have the **highest priority** in the per-field chain:

```
1. Struct tag directive    (doppel:"-", "shallow", "readonly", "clone", "deep")
2. Registered field Cloner  (auto-discovered)
3. Registered type Cloner   (via registry)
4. SelfClonable on the field
5. Reflection fallback
```

A struct tag always wins over registered field cloners, type cloners, and SelfClonable implementations.

---

## `doppel:"-"` — Skip a field

The field is excluded from the clone entirely. The cloned struct will have the zero value for that field:

```go
type Config struct {
    Name    string
    Secret  string `doppel:"-"`
    Tags    []string
}

src := Config{Name: "app", Secret: "password123", Tags: []string{"v1"}}
cloned, _ := doppel.CloneDeep(src, nil)

// cloned.Name  == "app"
// cloned.Secret == ""       (zero value — skipped)
// cloned.Tags  == []string{"v1"}
```

Use this for fields that should not appear in clones: credentials, internal caches, computed values, or fields that are re-initialized after cloning.

---

## `doppel:"shallow"` — Share the backing data

The field value is assigned without recursing. The clone and original share the same underlying data:

```go
type Document struct {
    Title   string
    Content []byte `doppel:"shallow"`
}

src := Document{Title: "report", Content: []byte("hello")}
cloned, _ := doppel.CloneDeep(src, nil)

src.Content[0] = 'H'
// cloned.Content[0] == 'H' — they share the same []byte backing array
```

### When to use `shallow`

- **Immutable data:** Large read-only byte slices or strings that will never be mutated.
- **Performance-critical paths:** When you are certain the shared data will not be modified.
- **Intentional sharing:** When the clone should reference the same resource (e.g., a shared buffer pool).

---

## `doppel:"readonly"` — Semantic alias for shallow

`readonly` has identical behavior to `shallow` — the field is assigned without recursing. The difference is purely semantic: `readonly` communicates that the field is **conceptually immutable** and safe to share:

```go
type Service struct {
    Name      string
    Config    map[string]string `doppel:"readonly"` // never modified after init
    Cache     map[string]any     // deep-copied (default)
}

src := Service{Name: "api", Config: map[string]string{"port": "8080"}}
cloned, _ := doppel.CloneDeep(src, nil)

src.Config["port"] = "9090"
// cloned.Config["port"] == "9090" — shared (readonly)
```

Use `readonly` instead of `shallow` when the primary reason for sharing is immutability, not performance. This makes the intent clear to readers.

---

## `doppel:"clone"` — Require a registered field cloner

Marks a field as requiring a registered field `Cloner`. If no field cloner is registered, the engine returns an error at clone time. This catches configuration mistakes early:

```go
type User struct {
    Name    string
    Address *Address `doppel:"clone"`
}

reg := registry.New()
registry.RegisterField[User, *Address](reg, "Address", core.NewFuncCloner(cloneAddr))

// Works — field cloner is registered
cloned, _ := doppel.CloneDeep(user, reg)

// Fails — no field cloner registered for "Address"
emptyReg := registry.New()
_, err := doppel.CloneDeep(user, emptyReg)
// Error: "doppel: field User.Address tagged doppel:\"clone\" but no field cloner is registered"
```

### Without a FieldLookup provider

If no `FieldLookup` is available at all (e.g., `engine.New(nil)`), the `doppel:"clone"` tag also errors:

```
doppel: field User.Address tagged doppel:"clone" but no FieldLookup provider is available
```

### When to use `doppel:"clone"`

- **Contract enforcement:** Make it explicit that this field requires custom clone logic.
- **Safety net:** Catch missing registrations early rather than silently falling through to default reflection.
- **Documentation:** Tag the field so readers know it has special clone requirements.

---

## `doppel:"deep"` — Explicit deep copy

Same behavior as the default (no tag). The field is recursively deep-copied. Use this tag to **document intent** — it tells readers "this field is intentionally deep-copied" without changing any behavior:

```go
type Profile struct {
    Name    string
    History []Event `doppel:"deep"` // explicit: this is deep-copied (same as default)
}
```

This is most useful in structs that mix `shallow`/`readonly`/`-` tags with the default behavior. Tagging the default fields with `doppel:"deep"` makes the mixed strategy visually clear:

```go
type CachedResponse struct {
    Data    []byte `doppel:"shallow"` // shared — large, immutable payload
    Headers map[string]string         // deep-copied (default, implicit)
    Meta    *ResponseMeta `doppel:"deep"` // deep-copied (explicit, documented)
}
```

---

## Combining tags

Each field can have exactly one `doppel` tag. You cannot combine directives on a single field. If you need both "require a field cloner" and "the result should be shared," register a field cloner that returns the same pointer.

---

## Tag priority over field cloners

A struct tag always wins over a registered field cloner. This is the full per-field chain:

```
doppel:"-"        → skip (field cloner is ignored)
doppel:"shallow"  → share (field cloner is ignored)
doppel:"clone"    → require field cloner (field cloner IS used)
doppel:"deep"     → deep copy (field cloner is ignored; falls through to type cloner / SelfClonable / reflection)
(no tag)          → check field cloner → type cloner → SelfClonable → reflection
```

The exception is `doppel:"clone"`, which specifically **requires** a field cloner and uses it.

---

## Summary table

| Tag | Field cloner consulted? | Behavior |
|-----|------------------------|----------|
| `-` | No | Skip; zero value in clone |
| `shallow` | No | Share backing data |
| `readonly` | No | Share backing data (semantic alias) |
| `clone` | **Yes (required)** | Use registered field cloner; error if missing |
| `deep` | No | Deep copy (same as no tag) |
| *(no tag)* | Yes (optional) | Use field cloner if registered, else deep copy |

---

## What's next?

- **[Cycle & Sharing Policy](cycle-policy.md)** — How the engine handles cyclic and shared pointer graphs.
- **[Patterns & Best Practices](patterns.md)** — Real-world struct tag patterns and field cloner strategies.

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Struct Tags
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="reflection-engine.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Reflection Engine</span>
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
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"><a href="cycle-policy.md" style="display: flex; align-items: center; justify-content: flex-end; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #10b981, #047857); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; text-align: right; box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);">
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Next</span>
                    <span style="font-size: 1rem; font-weight: 600;">Cycle Policy</span>
                </span>
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">→</span>
            </a></div>
    </div>
</div>
<!-- /Navigation -->


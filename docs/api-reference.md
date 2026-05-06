# API Reference

This page documents all public types, constants, and functions exported by the
doppel library.

---

## Package `doppel`

> `import "github.com/seyallius/doppel"`

### `func Clone[T any](src core.SelfClonable[T]) (T, error)`

Produces a deep copy of `src` by calling `src.Clone()`. The generic parameter
`T` is constrained to `core.SelfClonable[T]`, which means the type must
implement a `Clone() (T, error)` method. This is a zero-overhead dispatch —
no reflection, no registry lookup.

```go
cloned, err := doppel.Clone(user)
```

### `func MustClone[T any](src core.SelfClonable[T]) T`

Like `Clone`, but **panics** instead of returning an error. Intended for use
in tests and program initialization, where a cloning failure is always a
programming error.

```go
cloned := doppel.MustClone(user)
```

---

## Package `core`

> `import "github.com/seyallius/doppel/core"`

### Types

#### `type TagValue string`

A type-safe alias for doppel struct tag directives. Using this alias in
switch statements prevents accidental typos.

#### `type TagDirective struct`

Represents the parsed result of a doppel struct tag. Each boolean field is
mutually exclusive — at most one will be `true`.

| Field     | Type   | Tag                | Description                                |
|-----------|--------|--------------------|--------------------------------------------|
| `Skip`    | `bool` | `doppel:"-"`       | Exclude from clone                         |
| `Shallow` | `bool` | `doppel:"shallow"` | Shared reference                           |
| `Deep`    | `bool` | `doppel:"deep"`    | Full deep copy (default)                   |
| `Clone`   | `bool` | `doppel:"clone"`   | User-provided clone function               |
| `Empty`   | `bool` | `doppel:"empty"`   | Empty-but-non-nil for collections/pointers |

#### `type CloneError struct`

Carries contextual path information about a cloning failure, making it
straightforward to identify which field triggered the error.

| Field     | Type     | Description                                  |
|-----------|----------|----------------------------------------------|
| `Context` | `string` | Human-readable path (e.g., `"User.Address"`) |
| `Cause`   | `error`  | Underlying error that triggered the failure  |

#### `type Cloner[T any] interface`

The central extension interface. Any value that can produce a deep copy of a
`T` satisfies `Cloner[T]`.

```go
type Cloner[T any] interface {
    Clone(src T) (T, error)
}
```

#### `type FuncCloner[T any] struct`

Adapts a plain function to the `Cloner[T]` interface. Construct with
`NewFuncCloner`.

#### `type SelfClonable[T any] interface`

Optional interface that types implement so `doppel.Clone` can dispatch
directly without an external `Cloner`.

```go
type SelfClonable[T any] interface {
    Clone() (T, error)
}
```

### Constants

#### Tag key

| Constant | Value      | Description                                         |
|----------|------------|-----------------------------------------------------|
| `TagKey` | `"doppel"` | The struct tag key consulted by the code generator. |

#### Tag values (`TagValue`)

| Constant     | Value       | Description                                         |
|--------------|-------------|-----------------------------------------------------|
| `TagSkip`    | `"-"`       | Omit the field; clone receives zero value.          |
| `TagShallow` | `"shallow"` | Copy by direct assignment; shares backing data.     |
| `TagClone`   | `"clone"`   | Use a user-provided `clone<Type><Field>` function.  |
| `TagDeep`    | `"deep"`    | Full recursive deep copy (default).                 |
| `TagEmpty`   | `"empty"`   | Produce empty-but-non-nil for collections/pointers. |

#### Sentinel errors

| Constant       | Description                                                      |
|----------------|------------------------------------------------------------------|
| `ErrNilSource` | Reserved for future strict nil-rejection mode.                   |
| `ErrNoCloner`  | No cloner registered and type does not implement `SelfClonable`. |

### Functions

#### `func ParseTagValue(raw string) TagValue`

Validates and converts a raw tag string to a `TagValue`. Returns `TagDeep`
as default for empty or unrecognized values.

```go
tv := core.ParseTagValue("shallow") // → core.TagShallow
tv := core.ParseTagValue("unknown") // → core.TagDeep (default)
```

#### `func ParseTag(tagValue string) TagDirective`

Parses a doppel struct tag value into a `TagDirective`. Returns a directive
where at most one boolean field is `true`. Empty or unrecognized values
default to `Deep`.

```go
dir := core.ParseTag("-")       // → TagDirective{Skip: true}
dir := core.ParseTag("empty")   // → TagDirective{Empty: true}
dir := core.ParseTag("")        // → TagDirective{Deep: true}
```

#### `func NewFuncCloner[T any](cloneFn func(T) (T, error)) *FuncCloner[T]`

Wraps `cloneFn` as a `Cloner[T]`.

```go
cloner := core.NewFuncCloner(func(src MyType) (MyType, error) {
    return src, nil
})
```

#### `func WrapError(context string, cause error) error`

Creates a `CloneError` that annotates `cause` with a field-path `context`.
Use inside manual `Clone()` implementations:

```go
addr, err := manual.ClonePointer(u.Address, cloneAddress)
if err != nil {
    return User{}, core.WrapError("User.Address", err)
}
```

#### `func (*CloneError) Error() string`

Returns a descriptive error string including the context path.

#### `func (*CloneError) Unwrap() error`

Exposes the underlying cause for `errors.Is` / `errors.As` inspection.

#### `func (*FuncCloner[T]) Clone(src T) (T, error)`

Delegates to the wrapped function.

---

## Package `manual`

> `import "github.com/seyallius/doppel/manual"`

### Functions

#### `func CloneSlice[T any](src []T, cloneElem func(T) (T, error)) ([]T, error)`

Creates an independent deep copy of `src`. `cloneElem` is called once per
element. For primitive types, pass `manual.Identity[T]`.

- Nil `src` → `(nil, nil)` — nil is preserved.
- Empty non-nil `src` → fresh empty slice.

On error, returns `nil` and a contextual error identifying the offending index.

#### `func CloneMap[K comparable, V any](src map[K]V, cloneVal func(V) (V, error)) (map[K]V, error)`

Creates an independent deep copy of `src`. Keys (comparable value types) are
copied automatically; only values are cloned via `cloneVal`.

- Nil `src` → `(nil, nil)`.
- Empty non-nil `src` → fresh empty map.

#### `func ClonePointer[T any](src *T, cloneVal func(T) (T, error)) (*T, error)`

Creates an independent deep copy of the value `src` points to. The result is
placed into freshly allocated memory.

- Nil `src` → `(nil, nil)` — preserved without error.

#### `func Identity[T any](src T) (T, error)`

Returns `src` unchanged with a `nil` error. The correct element/value cloner
for primitive types.

#### `func IdentityValue[T any](src T) T`

Returns `src` unchanged without an error return. Useful when callers need an
infallible `func(T) T` signature.

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • API Reference
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="benchmarks.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Benchmarks</span>
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
        <div style="flex: 1; min-width: 200px; display: flex; justify-content: flex-end;"></div>
    </div>
</div>
<!-- /Navigation -->


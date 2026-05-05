# Patterns & Best Practices

Proven patterns for using doppel effectively in real-world Go projects. These patterns cover the most common scenarios and help you avoid pitfalls.

---

## Pattern 1: Full manual clone (maximum performance)

**When to use:** Performance-critical hot paths where you control the type and want zero reflection.

```go
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

**Dispatch:** `doppel.Clone(user)` — goes directly to `User.Clone()`, no registry lookup, no reflection.

---

## Pattern 2: Big struct with selective override

**When to use:** Structs with many fields (50+) where only a few need custom cloning. This is the primary Phase 3 use case.

```go
// Imagine 200 fields. Only 2 need custom logic.
reg := registry.New()

registry.RegisterField[BigStruct, *Address](reg, "Address", core.NewFuncCloner(
    func(src *Address) (*Address, error) {
        if src == nil { return nil, nil }
        cp := *src
        return &cp, nil
    },
))

registry.RegisterField[BigStruct, []string](reg, "Tags", core.NewFuncCloner(
    func(src []string) ([]string, error) {
        return append([]string{}, src...), nil
    },
))

cloned, err := doppel.CloneDeep(bigStruct, reg)
```

**Result:** 198 fields are deep-copied by reflection. 2 fields use custom cloners. No `Clone()` method needed.

---

## Pattern 3: Third-party type cloning

**When to use:** You need to clone a type from a third-party package that you can't modify.

```go
// thirdparty.User — you can't add a Clone() method to this

reg := registry.New()
registry.Register(reg, core.NewFuncCloner(func(u thirdparty.User) (thirdparty.User, error) {
    tags := make([]string, len(u.Tags))
    copy(tags, u.Tags)
    return thirdparty.User{
        ID:   u.ID,
        Name: u.Name,
        Tags: tags,
    }, nil
}))

cloned, err := doppel.CloneWithRegistry(externalUser, reg)
```

---

## Pattern 4: Conditional cloning

**When to use:** You want different clone behavior depending on the data.

```go
registry.Register(reg, core.NewFuncCloner(func(src *User) (*User, error) {
    if !src.Active {
        // Inactive users get a lightweight placeholder
        return &User{ID: src.ID, Name: "[inactive]"}, nil
    }
    // Active users get a full deep copy
    return src.Clone()
}))
```

---

## Pattern 5: Mixed shallow/deep with struct tags

**When to use:** Structs with a mix of shared and independent fields.

```go
type CachedResponse struct {
    Payload  []byte            `doppel:"shallow"` // large, immutable — share
    Headers  map[string]string `doppel:"readonly"` // config — safe to share
    Metadata *ResponseMeta     // deep-copied (default)
    AuditLog []AuditEntry      `doppel:"deep"`    // explicit deep copy
}
```

---

## Pattern 6: Struct tag enforcement

**When to use:** You want to catch missing registrations at runtime rather than silently falling through to reflection.

```go
type SecureUser struct {
    Name     string
    Password *Credentials `doppel:"clone"` // MUST have a registered field cloner
    Settings *Settings    `doppel:"clone"` // MUST have a registered field cloner
}

reg := registry.New()
registry.RegisterField[SecureUser, *Credentials](reg, "Password", core.NewFuncCloner(
    func(src *Credentials) (*Credentials, error) {
        // Custom clone: hash the password, rotate the salt, etc.
        return &Credentials{Hash: hash(src.Plain), Salt: newSalt()}, nil
    },
))

// If you forget to register Settings:
_, err := doppel.CloneDeep(user, reg)
// Error: "doppel: field SecureUser.Settings tagged doppel:\"clone\" but no field cloner is registered"
```

---

## Pattern 7: Registry as application singleton

**When to use:** You want to configure cloners once at startup and share them across the application.

```go
// setup.go — called once at init
var GlobalRegistry *registry.Registry

func InitRegistry() {
    GlobalRegistry = registry.New()

    registry.Register(GlobalRegistry, core.NewFuncCloner(cloneUser))
    registry.Register(GlobalRegistry, core.NewFuncCloner(cloneOrder))
    registry.RegisterField[User, *Address](GlobalRegistry, "HomeAddr", cloneAddr)
}

// handler.go — uses the global registry
func HandleCloneRequest(w http.ResponseWriter, r *http.Request) {
    user := loadUser(r.Context())
    cloned, err := doppel.CloneDeep(user, GlobalRegistry)
    // ...
}
```

**Thread safety:** `Registry` is safe for concurrent use. A single instance can be shared across all goroutines without additional synchronization.

---

## Pattern 8: Cloning for safe concurrent access

**When to use:** You need to pass data to a goroutine without race conditions.

```go
func ProcessUser(u *User) {
    // Create an independent copy for this goroutine
    snapshot := doppel.MustClone(u)

    go func() {
        // Safe: modifications to snapshot don't affect the original
        snapshot.Name = "processing"
        doWork(snapshot)
    }()
}
```

---

## Pattern 9: Test fixtures

**When to use:** Creating reusable test data that can be mutated without affecting other tests.

```go
func TestUserUpdate(t *testing.T) {
    original := newUser() // shared fixture

    t.Run("update name", func(t *testing.T) {
        user := doppel.MustClone(original)
        user.Name = "updated"
        assert.Equal(t, "updated", user.Name)
    })

    t.Run("add tag", func(t *testing.T) {
        user := doppel.MustClone(original)
        user.Tags = append(user.Tags, "new")
        assert.Equal(t, 4, len(user.Tags))
    })
}
```

Each sub-test gets its own independent copy. Mutations in one test don't affect others.

---

## Pattern 10: Handling cyclic data

**When to use:** Your data structures form graphs with cycles or shared references.

```go
// Default: preserve sharing and cycles
cloned := doppel.MustCloneDeep(graph, nil)
// clone.A.Next == clone.B.Prev (sharing preserved)

// Serialization: break cycles
eng := engine.NewWithOptions(nil, engine.Options{
    CyclePolicy: engine.BreakCycles,
})
clonedVal, _ := eng.Clone(reflect.ValueOf(graph))
// Back-edges become nil — safe for JSON marshaling

// Validation: detect cycles early
eng2 := engine.NewWithOptions(nil, engine.Options{
    CyclePolicy: engine.ErrorOnCycle,
})
_, err := eng2.Clone(reflect.ValueOf(graph))
if errors.As(err, &engine.CycleError{}) {
    log.Printf("data has cycles — not suitable for export")
}
```

---

## Anti-patterns to avoid

### Don't forget nil checks

```go
// BAD — panics if u is nil
func (u *User) Clone() (*User, error) {
    tags, _ := manual.CloneSlice(u.Tags, manual.Identity[string])
    // ...
}

// GOOD — nil is handled
func (u *User) Clone() (*User, error) {
    if u == nil { return nil, nil }
    tags, _ := manual.CloneSlice(u.Tags, manual.Identity[string])
    // ...
}
```

### Don't skip WrapError

```go
// BAD — loses context
func (u *User) Clone() (*User, error) {
    tags, err := manual.CloneSlice(u.Tags, cloneElem)
    if err != nil { return nil, err } // no context

    // GOOD — preserves field path
    if err != nil { return nil, core.WrapError("User.Tags", err) }
}
```

### Don't assume reflection covers unexported fields

```go
type User struct {
    Name   string
    secret string // NOT cloned by reflection
}

// Using CloneDeep: secret will be zero in the clone.
// To include it, implement SelfClonable:
func (u *User) Clone() (*User, error) {
    return &User{
        Name:   u.Name,
        secret: u.secret, // now included
    }, nil
}
```

### Don't share the same manual function across unrelated types

```go
// OK — different types, different registrations
registry.Register(reg, core.NewFuncCloner(cloneUser))
registry.Register(reg, core.NewFuncCloner(cloneOrder))

// BAD — reusing a cloner for an incompatible type
// This would panic or produce wrong results
```

---

## What's next?

- **[Benchmarks](benchmarks.md)** — Performance data to help you choose between manual, registry, and reflection approaches.
- **[API Reference](api-reference.md)** — Complete function and type signatures.

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->

---

<div style="margin-top: 3rem; margin-bottom: 1rem; padding: 2rem 1.5rem; border-top: 2px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c);">
    <div style="margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #e1e4e8; text-align: center; color: #586069; font-size: 0.85rem;">
        📚 doppel Documentation • Patterns & Best Practices
    </div>
    <div style="display: flex; justify-content: space-between; align-items: stretch; gap: 1.5rem; flex-wrap: wrap; margin-top: 1.5rem;">
        <div style="flex: 1; min-width: 200px;"><a href="error-handling.md" style="display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1.5rem; background: linear-gradient(135deg, #3b82f6, #1d4ed8); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 14px; line-height: 1.4; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);">
                <span style="font-size: 1.2rem; font-weight: 700; line-height: 1;">←</span>
                <span style="display: flex; flex-direction: column; line-height: 1.3;">
                    <span style="font-size: 0.7rem; opacity: 0.85; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 2px;">Previous</span>
                    <span style="font-size: 1rem; font-weight: 600;">Error Handling</span>
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


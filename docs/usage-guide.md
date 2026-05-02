# 🛠️ Usage Guide

> Step-by-step patterns for cloning increasingly complex data structures. Follow along! ✨

---

## Step 1 — Simple Struct (Primitives Only)

A struct whose fields are all primitive types needs no helpers. A plain struct literal in `Clone()` is sufficient.

```go
type Address struct {
    Street string
    City   string
    State  string
    Zip    string
}

func (a Address) Clone() (Address, error) {
    return Address{
        Street: a.Street,
        City:   a.City,
        State:  a.State,
        Zip:    a.Zip,
    }, nil
}
```

✅ Primitives are value types — no heap allocation to share.

---

## Step 2 — Struct with a Pointer Field

Use `manual.ClonePointer` to allocate a new pointer and clone the pointed-to value independently.

```go
type ContactInfo struct {
    Email   string
    Phone   string
    Address *Address
}

func cloneContactInfo(src ContactInfo) (ContactInfo, error) {
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
```

✅ Always wrap errors with `core.WrapError` for contextual debugging.

---

## Step 3 — Struct with Slices and Maps

Use `manual.CloneSlice` and `manual.CloneMap`. For primitive element/value types, pass `manual.Identity[T]`.

```go
type Profile struct {
    Tags   []string
    Scores map[string]int
    Badges []string
}

func (p Profile) Clone() (Profile, error) {
    tags, err := manual.CloneSlice(p.Tags, manual.Identity[string])
    if err != nil {
        return Profile{}, core.WrapError("Profile.Tags", err)
    }

    scores, err := manual.CloneMap(p.Scores, manual.Identity[int])
    if err != nil {
        return Profile{}, core.WrapError("Profile.Scores", err)
    }

    // Infallible shorthand for primitive slices
    badges := manual.CloneSliceOf(p.Badges, manual.IdentityValue[string])

    return Profile{Tags: tags, Scores: scores, Badges: badges}, nil
}
```

✅ Use `*Of` helpers for infallible primitive cloning to reduce boilerplate.

---

## Step 4 — Full Aggregate with Nested Structs

Everything composes. Each layer calls the clone function of the layer below it.

```go
type User struct {
    ID      int64
    Name    string
    Active  bool
    Contact ContactInfo
    Tags    []string
    Scores  map[string]int
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

// Call site:
cloned, err := doppel.Clone(user)

```

✅ Composition is key — each type owns its clone logic.

---

## Step 5 — External Cloner (No SelfClonable)

When a type does not implement `SelfClonable` — e.g., a third-party struct — use `core.NewFuncCloner` and `doppel.CloneWith`.

```go
// ThirdPartyConfig is a type you don't own.
type ThirdPartyConfig struct {
    Host    string
    Port    int
    Options map[string]string
}

configCloner := core.NewFuncCloner(func(src ThirdPartyConfig) (ThirdPartyConfig, error) {
    opts, err := manual.CloneMap(src.Options, manual.Identity[string])
    if err != nil {
        return ThirdPartyConfig{}, core.WrapError("ThirdPartyConfig.Options", err)
    }
    return ThirdPartyConfig{Host: src.Host, Port: src.Port, Options: opts}, nil
})

cloned, err := doppel.CloneWith(cfg, configCloner)
```

✅ External cloners let you clone types you can't modify.

---

## Step 6 — Conditional / Filtered Cloning

Because you supply the clone function, you have full control over what goes into the clone.

```go
// Clone a map, but only carry over entries whose value is above a threshold.
aboveThreshold, err := manual.CloneMap(rawScores, func(score int) (int, error) {
    if score < passingGrade {
        return 0, nil // zero-out failing scores
    }
    return score, nil
})

// Clone a slice, skipping nil pointers entirely.
validUsers, err := manual.CloneSlice(allUsers, func(u *User) (*User, error) {
    if u == nil {
        return nil, nil // preserved as nil in clone
    }
    return u.Clone()
})
```

✅ This is the preview of field-level customization (Phase 3).

---

## Step 7 — Cloner Registry (Phase 2)

Register custom clone logic for types at application startup. The registry is thread-safe.

```go
reg := registry.New()

// Register a custom cloner for Address
registry.Register(reg, core.NewFuncCloner(func(src Address) (Address, error) {
    return Address{City: strings.ToUpper(src.City)}, nil // transform on clone
}))

// Use CloneWithRegistry — it will find the registered cloner automatically
cloned, err := doppel.CloneWithRegistry(addr, reg)
```

**Priority Chain in `CloneWithRegistry`**:
```
1. Registered Cloner[T] → 2. SelfClonable[T] → 3. core.ErrNoCloner
```

✅ Registries enable centralized clone config without modifying types.

---

## Step 8 — Reflection Fallback with Cycle Policies (Phase 4+5)

When you have a type with no manual clone and no registered cloner, use the reflection engine as a safe fallback.

```go
type GraphNode struct {
    ID    int
    Links []*GraphNode  // may contain cycles
}

src := &GraphNode{ID: 1}
src.Links = []*GraphNode{src}  // self-referential cycle

// Option A: PreserveShared (default) — reproduce exact topology
eng1 := engine.New(nil)
cloned1, _ := eng1.Clone(reflect.ValueOf(src))
// cloned1.(*GraphNode).Links[0] == cloned1  ← cycle preserved

// Option B: BreakCycles — produce acyclic output for serialization
eng2 := engine.NewWithOptions(nil, engine.Options{
    CyclePolicy: engine.BreakCycles,
})
cloned2, _ := eng2.Clone(reflect.ValueOf(src))
// cloned2.(*GraphNode).Links[0] == nil  ← cycle broken, safe for JSON

// Option C: ErrorOnCycle — strict validation during development
eng3 := engine.NewWithOptions(nil, engine.Options{
    CyclePolicy: engine.ErrorOnCycle,
})
_, err := eng3.Clone(reflect.ValueOf(src))
// err.(*engine.CycleError) → "cycle detected at 0x... (type *GraphNode)"
```

> 💡 **Pro Tip**: Prefer manual `Clone()` implementations for performance-critical paths. Use reflection fallback for prototyping or legacy integration. Benchmarks show manual cloning is ~3-6× faster. (◕‿◕)✧

<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->
<div class="doppel-nav-container"
     style="margin-top: 3rem; padding: 1.75rem; border-top: 1px solid #1e293b; border-radius: 12px; background: linear-gradient(145deg, #0f172a, #0b111c); box-shadow: 0 8px 24px rgba(0,0,0,0.4);">
    <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 1rem;">
        <div style="flex: 1; min-width: 200px;">
            <a href="api-reference.md" class="doppel-nav-btn doppel-nav-prev">
            <span class="doppel-arrow">←</span>
            <div class="doppel-text">
                <span class="doppel-label">Previous</span>
                <span class="doppel-title">API Reference</span>
            </div>
        </a>
        </div>
        <div style="flex: 1; min-width: 200px; text-align: right;">
            <a href="advanced.md" class="doppel-nav-btn doppel-nav-next">
            <div class="doppel-text">
                <span class="doppel-label">Next</span>
                <span class="doppel-title">Advanced</span>
            </div>
            <span class="doppel-arrow">→</span>
        </a>
        </div>
    </div>
    <div style="margin-top: 1.25rem; text-align: center; color: #94a3b8; font-size: 0.8rem; letter-spacing: 0.03em;">
        <span>📚 doppel Documentation • Usage Guide</span>
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

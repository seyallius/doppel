// Package doppel provides safe, explicit deep cloning of Go data structures.
//
// "Your data's doppelgänger — deep copies without side effects."
//
// # Design philosophy
//
// doppel is built around a performance-first, explicit-over-magic mindset:
//
//  1. Manual cloning is the DEFAULT — no reflection, no magic, maximum speed.
//  2. Reflection is a fallback — introduced in Phase 4, only when no manual
//     clone exists for a type.
//  3. Everything is composable — CloneSlice, CloneMap, and ClonePointer are
//     generic helpers you wire together inside your type's own Clone() method.
//  4. Everything is extensible — register a per-type or per-field Cloner to
//     override the default behavior (Phase 2 / Phase 3).
//
// # Phase 1 — Manual Deep Copy Foundation
//
// In this phase the library ships with:
//   - core.Cloner[T]          — the extension interface.
//   - core.SelfClonable[T]    — optional interface for self-cloning types.
//   - manual.CloneSlice / manual.CloneSliceOf     — generic slice deep copy.
//   - manual.CloneMap   / manual.CloneMapOf       — generic map deep copy.
//   - manual.ClonePointer / manual.ClonePointerOf — generic pointer deep copy.
//   - manual.Identity   / manual.IdentityValue    — no-op helpers for primitives.
//   - doppel.Clone                                — dispatches to src.Clone() for SelfClonable types.
//   - doppel.CloneWith                            — dispatches to an external Cloner[T].
//
// # Phase 2 — Cloner Registry
//
// The registry package adds a thread-safe, type-keyed store of Cloner[T] values.
// Reflection is used only for type key derivation — never for cloning itself.
//   - registry.New()         	— create an empty Registry.
//   - registry.Register[T]   	— store a Cloner[T] for type T.
//   - registry.Lookup[T]     	— retrieve the Cloner[T] for type T.
//   - registry.Deregister[T] 	— remove the Cloner for type T.
//   - registry.Has[T]        	— report whether a Cloner is registered for T.
//   - doppel.CloneWithRegistry — dispatches via registry → SelfClonable → ErrNoCloner.
//
// # Phase 3 — Field-Level Customization
//
// Field-level cloners provide fine-grained control over individual struct fields.
// Instead of writing a full Clone() method for a 200-field struct, register cloners
// for only the fields that need custom logic. The reflection engine handles the rest.
//   - registry.RegisterField[T, F]  — store a Cloner[F] for a specific field of T.
//   - registry.LookupField[T, F]    — retrieve the field Cloner.
//   - registry.HasField[T]          — check if a field Cloner exists.
//   - registry.DeregisterField[T]   — remove a field Cloner.
//   - doppel.CloneDeep              — full priority chain including reflection.
//
// # Quick start
//
//	// 1. Implement Clone() on your type.
//	func (u *User) Clone() (*User, error) {
//	    tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
//	    if err != nil {
//	        return nil, core.WrapError("User.Tags", err)
//	    }
//	    return &User{ID: u.ID, Name: u.Name, Tags: tags}, nil
//	}
//
//	// 2. Call doppel.Clone.
//	cloned, err := doppel.Clone(user)
//
//	// 3. Or use CloneDeep for "default deep copy + selective override":
//	reg := registry.New()
//	registry.RegisterField[BigStruct, *Address](reg, "HomeAddr", core.NewFuncCloner(cloneAddr))
//	cloned, err := doppel.CloneDeep(bigStruct, reg)
package doppel

import (
	"reflect"

	"github.com/seyallius/doppel/core"
	"github.com/seyallius/doppel/engine"
	"github.com/seyallius/doppel/registry"
)

// Clone produces a deep copy of src by calling src.Clone().
// src must satisfy core.SelfClonable[T]; the compiler enforces this.
//
// For types that do not implement SelfClonable, use CloneWith together
// with a core.Cloner[T] or core.FuncCloner[T].
func Clone[T any](src core.SelfClonable[T]) (T, error) {
	return src.Clone()
}

// MustClone is like Clone but panics instead of returning an error.
// Intended for use in tests and program initialization, where a cloning
// failure is always a programming error rather than a recoverable condition.
func MustClone[T any](src core.SelfClonable[T]) T {
	result, err := src.Clone()
	if err != nil {
		panic("doppel.MustClone: " + err.Error())
	}
	return result
}

// CloneWith produces a deep copy of src using the provided external Cloner.
// Use this when the source type does not implement SelfClonable — for example,
// when cloning logic lives in a separate struct with injected dependencies, or
// when you want to override the default clone for a specific call site.
//
//	cloner := core.NewFuncCloner(func(u User) (User, error) { ... })
//	cloned, err := doppel.CloneWith(original, cloner)
func CloneWith[T any](src T, cloner core.Cloner[T]) (T, error) {
	return cloner.Clone(src)
}

// MustCloneWith is like CloneWith but panics instead of returning an error.
func MustCloneWith[T any](src T, cloner core.Cloner[T]) T {
	result, err := cloner.Clone(src)
	if err != nil {
		panic("doppel.MustCloneWith: " + err.Error())
	}
	return result
}

// CloneWithRegistry produces a deep copy of src by walking the following
// lookup chain in order, stopping at the first strategy that applies:
//
//  1. Registered Cloner[T] — if r contains a Cloner for type T, it is used.
//     This is the fastest path and the whole point of the registry: you control
//     the clone logic for the types you care about most.
//
//  2. core.SelfClonable[T] fallback — if T implements SelfClonable[T], its
//     Clone() method is called. This means all existing SelfClonable types
//     work with CloneWithRegistry out of the box, even without registration.
//
//  3. core.ErrNoCloner — returned when neither strategy is available.
//     This is an explicit signal to either register a Cloner, implement
//     SelfClonable, or use CloneDeep for automatic reflection fallback.
//
// Reflection is used only inside the registry for type key derivation —
// never for field access or value traversal.
func CloneWithRegistry[T any](src T, reg *registry.Registry) (T, error) {
	// 1. Registry fast path — O(1) map lookup with a read lock.
	if cloner, found := registry.Lookup[T](reg); found {
		return cloner.Clone(src)
	}

	// 2. SelfClonable fallback — no registry entry for T, but the value
	//    might carry its own Clone() method.
	if selfClonable, ok := any(src).(core.SelfClonable[T]); ok {
		return selfClonable.Clone()
	}

	var zero T
	return zero, core.ErrNoCloner
}

// CloneDeep produces a deep copy of src by walking the full priority chain:
//
//  1. Registered Cloner[T] — if reg contains a Cloner for type T, it is used.
//     This is the fastest path and gives you full control over the clone logic.
//
//  2. core.SelfClonable[T] — if T implements SelfClonable[T], its Clone()
//     method is called. All existing SelfClonable types work out of the box.
//
//  3. Reflection engine — the engine recursively clones the value, consulting
//     registered field-level cloners (via registry.RegisterField) before
//     falling through to default reflection for each struct field.
//
// CloneDeep is the entry point for the "default deep copy + selective override"
// workflow introduced in Phase 3. It is the recommended API when you have a
// struct with many fields but only need custom clone logic for a few:
//
//	reg := registry.New()
//	registry.RegisterField[BigStruct, *Address](reg, "HomeAddr", core.NewFuncCloner(cloneAddr))
//	registry.RegisterField[BigStruct, []string](reg, "Tags", core.NewFuncCloner(
//	    func(src []string) ([]string, error) { return append([]string{}, src...), nil },
//	))
//	cloned, err := doppel.CloneDeep(bigStruct, reg)
//
// Pass nil for reg to use pure reflection without any registered cloners.
//
// The engine respects doppel struct tags (see engine package documentation for details).
func CloneDeep[T any](src T, reg *registry.Registry) (T, error) {
	// 1. Registry fast path — type-level cloner.
	if reg != nil {
		if cloner, found := registry.Lookup[T](reg); found {
			return cloner.Clone(src)
		}
	}

	// 2. SelfClonable fallback.
	if selfClonable, ok := any(src).(core.SelfClonable[T]); ok {
		return selfClonable.Clone()
	}

	// 3. Reflection engine — reg may be nil; engine handles nil TypeLookup gracefully.
	var lookup engine.TypeLookup
	if reg != nil {
		lookup = reg // *registry.Registry implements TypeLookup (+ FieldLookup)
	}

	eng := engine.New(lookup)
	clonedVal, err := eng.Clone(reflect.ValueOf(src))
	if err != nil {
		var zero T
		return zero, err
	}

	if !clonedVal.IsValid() {
		var zero T
		return zero, nil
	}

	return clonedVal.Interface().(T), nil
}

// MustCloneDeep is like CloneDeep but panics instead of returning an error.
// Intended for use in tests and program initialization, where a cloning
// failure is always a programming error rather than a recoverable condition.
func MustCloneDeep[T any](src T, reg *registry.Registry) T {
	result, err := CloneDeep(src, reg)
	if err != nil {
		panic("doppel.MustCloneDeep: " + err.Error())
	}
	return result
}

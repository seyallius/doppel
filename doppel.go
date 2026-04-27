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
//   - doppel.CloneWithRegistry — dispatches via registry → SelfClonable → error.
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
package doppel

import (
	"github.com/seyallius/doppel/core"
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
//  3. registry.ErrNoCloner — returned when neither strategy is available.
//     This is an explicit signal to either register a Cloner or implement
//     SelfClonable on the type. (Reflection fallback arrives in Phase 4.)
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

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
//   - core.Cloner[T]    — the extension interface.
//   - core.SelfClonable[T] — optional interface for self-cloning types.
//   - manual.CloneSlice / manual.CloneSliceOf   — generic slice deep copy.
//   - manual.CloneMap   / manual.CloneMapOf     — generic map deep copy.
//   - manual.ClonePointer / manual.ClonePointerOf — generic pointer deep copy.
//   - manual.Identity   / manual.IdentityValue  — no-op helpers for primitives.
//   - doppel.Clone      — dispatches to src.Clone() for SelfClonable types.
//   - doppel.CloneWith  — dispatches to an external Cloner[T].
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

import "github.com/seyallius/doppel/core"

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

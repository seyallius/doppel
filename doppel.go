// Package doppel provides safe, explicit deep cloning of Go data structures.
//
// "Your data's doppelgänger — deep copies without side effects."
//
// # Design philosophy
//
// doppel is built around a performance-first, explicit-over-magic mindset:
//
//  1. Manual cloning is the ONLY path — no reflection, no magic, maximum speed.
//  2. Everything is composable — CloneSlice, CloneMap, and ClonePointer are
//     generic helpers you wire together inside your type's own Clone() method.
//  3. Everything is explicit — every clone path is visible and auditable.
//
// # Quick start
//
//	// 1. Implement Clone() on your type.
//	func (u *User) Clone() (*User, error) {
//	    if u == nil { return nil, nil }
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
// # Future: Code generator
//
// A future code generator will read doppel struct tags and emit Clone()
// implementations automatically. See core/tags.go for the tag contract.
package doppel

import "github.com/seyallius/doppel/core"

// Clone produces a deep copy of src by calling src.Clone().
// src must satisfy core.SelfClonable[T]; the compiler enforces this.
//
// This is a zero-overhead dispatch — Clone is a direct call to the type's
// own Clone() method with no reflection, no registry lookup, no priority chain.
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

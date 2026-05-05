// Package core. cloner - defines the foundational cloning interfaces (Cloner[T], SelfClonable[T])
// and the FuncCloner adapter for registering custom clone logic without reflection.
//
// This file establishes the extension contract for doppel:
//   - Cloner[T]: external clone logic for types you don't own or need context for.
//   - SelfClonable[T]: opt-in interface for types that manage their own deep-copy logic.
//   - FuncCloner[T]: lightweight adapter to use plain functions as Cloner[T] implementations.
//
// TODO: Generator integration — a future code generator will emit SelfClonable[T]
// implementations automatically from struct definitions and doppel tags.
//
//go:generate go run github.com/seyallius/doppel/cmd/doppelgen
package core

// -------------------------------------------- Types --------------------------------------------

// Cloner is the central extension interface for doppel.
// Any value that can produce a deep copy of a T satisfies Cloner[T].
// It is the contract used throughout the library for external clone logic.
//
// Implementations must:
//   - Never return a value that shares any mutable memory with src.
//   - Return a non-nil error only when cloning cannot complete safely.
//   - Be safe for concurrent calls (stateless, or internally synchronized).
type Cloner[T any] interface {
	Clone(src T) (T, error)
}

// FuncCloner adapts a plain function to the Cloner[T] interface.
// It lets callers register clone logic inline without a named struct type:
//
//	cloner := core.NewFuncCloner(func(src MyType) (MyType, error) { ... })
type FuncCloner[T any] struct {
	cloneFn func(T) (T, error)
}

// SelfClonable is an optional interface a type can implement so that
// doppel.Clone can dispatch directly to it without an external Cloner.
//
// Choose SelfClonable when the type owns all the state it needs to copy.
// When cloning requires external context (e.g., re-fetching a lazy field),
// implement Cloner[T] separately and keep the type itself unaware of cloning.
//
// Example:
//
//	func (u *User) Clone() (*User, error) {
//	    tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
//	    if err != nil {
//	        return nil, core.WrapError("User.Tags", err)
//	    }
//	    return &User{ID: u.ID, Name: u.Name, Tags: tags}, nil
//	}
type SelfClonable[T any] interface {
	Clone() (T, error)
}

// -------------------------------------------- Constructor(s) --------------------------------------------

// NewFuncCloner wraps cloneFn as a Cloner[T].
func NewFuncCloner[T any](cloneFn func(T) (T, error)) *FuncCloner[T] {
	return &FuncCloner[T]{cloneFn: cloneFn}
}

// -------------------------------------------- Public API --------------------------------------------

// Clone delegates to the wrapped function.
func (fc *FuncCloner[T]) Clone(src T) (T, error) { return fc.cloneFn(src) }

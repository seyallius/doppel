// Package core. cloner - Defines the foundational interfaces and error types for doppel.
// It has zero dependencies on any other doppel sub-package and no use of reflect.
package core

// -------------------------------------------- Types --------------------------------------------

// Cloner is the central extension interface for doppel.
// Any value that can produce a deep copy of a T satisfies Cloner[T].
// It is the contract used throughout the registry (Phase 2) and the
// field-level customization layer (Phase 3).
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
// When cloning requires external context (e.g. re-fetching a lazy field),
// implement Cloner[T] separately and keep the type itself unaware of cloning.
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

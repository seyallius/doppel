// Package core. errors - provides contextual error types (CloneError) and helpers (WrapError)
// that annotate cloning failures with field-path information for easier debugging.
//
// Key exports:
//   - CloneError: carries Context (field path) and Cause (underlying error) for precise diagnostics.
//   - WrapError: convenience constructor that builds CloneError with a human-readable context string.
//   - ErrNilSource: sentinel error reserved for future strict-nil modes (not used by default).
//   - ErrNoCloner: sentinel error returned by `CloneWithRegistry` when neither a registered `Cloner[T]` nor a `SelfClonable[T]` implementation is found.
//
// All errors implement CloneError.Unwrap() for compatibility with errors.Is/errors.As,
// enabling robust error handling in nested clone operations without losing the root cause.
package core

import (
	"errors"
	"fmt"
)

// -------------------------------------------- Types, Variables & Constants --------------------------------------------

var (
	// ErrNilSource is returned by ClonePointer when the source pointer is nil
	// and the caller has opted in to strict nil-rejection mode.
	// By default, ClonePointer propagates nil as nil (no error); this sentinel
	// is reserved for future optional strict mode.
	ErrNilSource = errors.New("doppel: clone source is nil")

	// ErrNoCloner is returned by doppel.CloneWithRegistry when neither a registered
	// Cloner[T] nor a core.SelfClonable[T] implementation is found for type T.
	// Use errors.Is to check for this sentinel in calling code.
	ErrNoCloner = errors.New("doppel/registry: no cloner registered and type does not implement SelfClonable")
)

// CloneError carries contextual path information about a cloning failure,
// making it straightforward to identify which field or index triggered the error.
type CloneError struct {
	Context string // Context is a human-readable path to the failing field, e.g. "User.Address".
	Cause   error  // Cause is the underlying error that triggered the failure.
}

// -------------------------------------------- Public API --------------------------------------------

// Error returns a descriptive error string including the context path.
func (e *CloneError) Error() string {
	return fmt.Sprintf("doppel: error cloning %s: %v", e.Context, e.Cause)
}

// Unwrap exposes the underlying cause for errors.Is / errors.As inspection.
func (e *CloneError) Unwrap() error { return e.Cause }

// WrapError creates a CloneError that annotates cause with a context path.
// Use this inside manual Clone() implementations to produce meaningful errors:
//
//	addr, err := manual.ClonePointer(u.Address, cloneAddress)
//	if err != nil {
//	    return User{}, core.WrapError("User.Address", err)
//	}
func WrapError(context string, cause error) error {
	return &CloneError{Context: context, Cause: cause}
}

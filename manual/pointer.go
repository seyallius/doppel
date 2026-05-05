// Package manual. pointer - provides ClonePointer[T] — a generic helper for creating
// independent deep copies of pointer values, allocating fresh memory for the clone.
//
// Key behaviors:
//   - Nil-safety: nil src → (nil, nil) without invoking the clone function.
//   - Independence: cloned pointer points to newly allocated memory; original and clone never share addresses.
//   - Error propagation: ClonePointer wraps cloneVal errors with core.WrapError for contextual debugging.
//
// Choose ClonePointer when cloning the pointed-to value can fail (e.g., nested struct with validation).
// For primitive types, pass manual.Identity[T] as the cloneVal function.
//
// Typical usage inside a struct's Clone() method:
//
//	addr, err := manual.ClonePointer(u.Address, cloneAddress)
//	if err != nil {
//	    return nil, core.WrapError("User.Address", err)
//	}
package manual

import "github.com/seyallius/doppel/core"

// ClonePointer creates an independent deep copy of the value that src points to.
//
// cloneVal is called on *src to produce the cloned payload; the result is
// placed into a freshly allocated *T so the original and clone never share
// the same pointer address.
//
// Nil-safety contract:
//   - If src is nil, (nil, nil) is returned — nil is preserved without error.
func ClonePointer[T any](src *T, cloneVal func(T) (T, error)) (*T, error) {
	if src == nil {
		return nil, nil
	}

	cloned, err := cloneVal(*src)
	if err != nil {
		return nil, core.WrapError("pointer", err)
	}

	dst := new(T)
	*dst = cloned

	return dst, nil
}

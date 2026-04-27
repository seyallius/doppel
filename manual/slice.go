// Package manual. slice - provides CloneSlice[T] and CloneSliceOf[T] — generic helpers
// for creating independent deep copies of slices with fallible or infallible element cloners.
//
// Key behaviors:
//   - Nil-safety: nil src → (nil, nil); empty non-nil src → fresh empty slice (preserves nil/empty distinction).
//   - Independence: cloned slice has its own backing array; mutations to src never affect the clone.
//   - Error context: on failure, CloneSlice returns an error annotated with the failing index.
//
// Choose CloneSlice when element cloning can fail (e.g., nested structs with validation).
// Choose CloneSliceOf for primitive types or infallible copy logic using IdentityValue[T].
//
// Performance note: These helpers allocate exactly one new slice header + backing array —
// no reflection overhead, no hidden allocations beyond what the element cloner requires.
package manual

import "fmt"

// CloneSlice creates an independent deep copy of src.
//
// cloneElem is called once per element to produce its copy. For primitive
// element types, pass manual.Identity[T]. For struct element types, pass
// a function that calls the struct's own Clone method.
//
// Nil-safety contract:
//   - A nil src returns (nil, nil) — nil is preserved as nil.
//   - An empty (non-nil) src returns a fresh empty slice, not nil.
//
// On error, CloneSlice returns nil and a contextual error that identifies
// the offending index.
func CloneSlice[T any](src []T, cloneElem func(T) (T, error)) ([]T, error) {
	if src == nil {
		return nil, nil
	}

	dst := make([]T, len(src))

	for idx, elem := range src {
		cloned, err := cloneElem(elem)
		if err != nil {
			return nil, fmt.Errorf("doppel: CloneSlice index [%d]: %w", idx, err)
		}
		dst[idx] = cloned
	}

	return dst, nil
}

// CloneSliceOf creates an independent deep copy of src using a simple
// (infallible) element cloner. It is the convenience sibling of CloneSlice
// for cases where cloning the element type cannot fail — typically primitive
// types when paired with manual.IdentityValue:
//
//	tags := manual.CloneSliceOf(u.Tags, manual.IdentityValue[string])
//
// A nil src returns nil; an empty src returns a fresh empty slice.
func CloneSliceOf[T any](src []T, cloneElem func(T) T) []T {
	if src == nil {
		return nil
	}

	dst := make([]T, len(src))

	for idx, elem := range src {
		dst[idx] = cloneElem(elem)
	}

	return dst
}

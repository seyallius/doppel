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

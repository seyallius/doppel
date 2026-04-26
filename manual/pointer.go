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
//
// Typical usage inside a struct's Clone() method:
//
//	addr, err := manual.ClonePointer(u.Address, cloneAddress)
//	if err != nil {
//	    return nil, core.WrapError("User.Address", err)
//	}
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

// ClonePointerOf creates an independent deep copy of the value that src points to
// using a simple (infallible) value cloner. Use this when the pointed-to value's
// clone function cannot fail — typically when cloneVal is IdentityValue[T] or a
// handwritten no-error copy function:
//
//	label, err := manual.ClonePointerOf(u.Label, manual.IdentityValue[string])
//
// If src is nil, nil is returned.
func ClonePointerOf[T any](src *T, cloneVal func(T) T) *T {
	if src == nil {
		return nil
	}

	dst := new(T)
	*dst = cloneVal(*src)

	return dst
}

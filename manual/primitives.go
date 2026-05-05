// Package manual. primitives - provides Identity[T] and IdentityValue[T] — no-op helpers
// that express "this primitive type needs no deep copy" explicitly in clone pipelines.
//
// Why these exist:
//   - For primitive Go types (bool, int, string, float, etc.), assignment IS a deep copy.
//   - These helpers make that intent explicit rather than relying on implicit behavior.
//   - Identity[T] returns (T, error) for use with fallible helpers (CloneSlice, CloneMap, ClonePointer).
//   - IdentityValue[T] returns T for use when callers want an infallible function signature.
//
// Usage example:
//
//	tags, err := manual.CloneSlice(user.Tags, manual.Identity[string])
//	scores, err := manual.CloneMap(user.Scores, manual.Identity[int])
//
//	// compose helpers inside a type's own Clone() method:
//	func (u *User) Clone() (*User, error) {
//	    tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
//	    ...
//	}
package manual

// Identity returns src unchanged together with a nil error.
//
// For all primitive Go types (bool, integer types, float types, complex types,
// string, uintptr) a direct copy IS a complete deep copy, because these types
// carry no pointers and contain no heap allocations. Identity is therefore the
// correct element cloner to pass to CloneSlice / CloneMap for primitive element
// types, e.g.:
//
//	tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
//	ids,  err := manual.CloneSlice(u.IDs,  manual.Identity[int])
func Identity[T any](src T) (T, error) {
	return src, nil
}

// IdentityValue returns src unchanged without an error return value.
// This is useful when callers need a func(T) T signature — for example,
// in custom clone functions that cannot fail:
//
//	func cloneValue(src string) string {
//	    return manual.IdentityValue(src)
//	}
func IdentityValue[T any](src T) T {
	return src
}

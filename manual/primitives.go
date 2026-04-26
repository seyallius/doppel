// Package manual provides generic helper functions for manual (reflection-free)
// deep copying of slices, maps, and pointers.
//
// Usage pattern — compose helpers inside a type's own Clone() method:
//
//	func (u *User) Clone() (*User, error) {
//	    tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
//	    ...
//	}
//
// No reflect package is imported anywhere in this package.
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
// Use this with CloneSliceOf and CloneMapOf when you want a terser call-site
// for slices / maps of primitive types and the clone function cannot fail:
//
//	tags := manual.CloneSliceOf(u.Tags, manual.IdentityValue[string])
func IdentityValue[T any](src T) T {
	return src
}

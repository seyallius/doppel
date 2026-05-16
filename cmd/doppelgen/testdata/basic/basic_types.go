// Package testdata provides fixture types for doppelgen parser and emitter tests.
package basic

// BasicUser is a struct with a mix of primitive, slice, map, skip, shallow, and empty fields.
type BasicUser struct {
	ID     int64             `doppel:"deep"`
	Name   string            `doppel:"deep"`
	Tags   []string          `doppel:"deep"`
	Scores map[string]int    `doppel:"deep"`
	Secret string            `doppel:"-"`
	Config map[string]string `doppel:"shallow"`
	Cache  []string          `doppel:"empty"`
}

// Address is a simple all-primitive struct for testing basic Clone() generation.
type Address struct {
	Street string `doppel:"deep"`
	City   string `doppel:"deep"`
	State  string `doppel:"deep"`
	Zip    string `doppel:"deep"`
}

// NestedUser contains pointer and slice fields that reference other struct types.
type NestedUser struct {
	ID      int64             `doppel:"deep"`
	Name    string            `doppel:"deep"`
	Address *Address          `doppel:"deep"`
	Items   []Address         `doppel:"deep"`
	Labels  map[string]string `doppel:"deep"`
}

// PointerPrimitives tests deep clone behaviour for pointer-to-primitive fields.
type PointerPrimitives struct {
	Name    *string `doppel:"deep"`
	Age     *int    `doppel:"deep"`
	Active  *bool   `doppel:"deep"`
	Secret  *string `doppel:"-"`
	Shallow *string `doppel:"shallow"`
	EmptyP  *string `doppel:"empty"`
}

// EmptyFields tests the "empty" tag directive across different type categories.
type EmptyFields struct {
	Name   string         `doppel:"empty"`
	Tags   []string       `doppel:"empty"`
	Scores map[int]string `doppel:"empty"`
	Addr   *Address       `doppel:"empty"`
}

// CloneTagUser uses the "clone" tag directive for convention-based clone function calls.
type CloneTagUser struct {
	ID      int64    `doppel:"deep"`
	Profile *Profile `doppel:"clone"`
}

// Profile is a supporting type referenced by CloneTagUser.
type Profile struct {
	Bio string `doppel:"deep"`
}

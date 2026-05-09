package testdata

// BasicUser has a mix of primitive and collection fields.
type BasicUser struct {
	ID     int64
	Name   string
	Tags   []string          `doppel:"deep"`
	Scores map[string]int    `doppel:"deep"`
	Secret string            `doppel:"-"`
	Config map[string]string `doppel:"shallow"`
	Cache  []string          `doppel:"empty"`
}

// Address is a simple struct with all primitive fields.
type Address struct {
	Street string
	City   string
	State  string
	Zip    string
}

// NestedUser has pointer and slice-of-struct fields.
type NestedUser struct {
	ID      int64
	Name    string
	Address *Address          `doppel:"deep"`
	Items   []Address         `doppel:"deep"`
	Labels  map[string]string `doppel:"deep"`
}

// ShallowAll copies everything by shallow assignment.
type ShallowAll struct {
	Name   string            `doppel:"shallow"`
	Items  []string          `doppel:"shallow"`
	Config map[string]string `doppel:"shallow"`
}

// EmptyFields demonstrates the empty tag.
type EmptyFields struct {
	Name   string         `doppel:"empty"` // primitive, ignored
	Tags   []string       `doppel:"empty"`
	Scores map[int]string `doppel:"empty"`
	Addr   *Address       `doppel:"empty"`
}

// NoTags struct has no doppel tags at all.
type NoTags struct {
	Name string
	Age  int
}

// unexportedStruct should be skipped (unexported).
type unexportedStruct struct {
	Value string `doppel:"deep"`
}

// CloneTagUser uses the clone tag for custom cloner.
type CloneTagUser struct {
	ID      int64
	Profile *Profile `doppel:"clone"`
}

// Profile is referenced by CloneTagUser.
type Profile struct {
	Handle string
	Bio    string
}

// ExistingClone already has a Clone() method and should be skipped.
type ExistingClone struct {
	Value int
}

func (e *ExistingClone) Clone() (*ExistingClone, error) {
	return &ExistingClone{Value: e.Value}, nil
}

// SkipGenStruct has a skip-gen comment and should be skipped.
// doppel:skip-gen
type SkipGenStruct struct {
	Data string `doppel:"deep"`
}

// PointerPrimitives demonstrates pointer-to-primitive types.
type PointerPrimitives struct {
	Name    *string `doppel:"deep"`
	Age     *int    `doppel:"deep"`
	Active  *bool   `doppel:"deep"`
	Secret  *string `doppel:"-"`
	Shallow *string `doppel:"shallow"`
	EmptyP  *string `doppel:"empty"`
}

// SliceOfPointers has a slice of pointer elements.
type SliceOfPointers struct {
	Addresses []*Address `doppel:"deep"`
}

// Package testdata provides fixture types for doppelgen parser and emitter tests.
package basic

// ExistingClone is a struct that already defines its own Clone() method.
// The generator should detect this and skip generation.
type ExistingClone struct {
	Value string `doppel:"deep"`
}

// Clone is a hand-written Clone method — doppelgen must not overwrite it.
func (e *ExistingClone) Clone() (*ExistingClone, error) {
	return &ExistingClone{Value: e.Value}, nil
}

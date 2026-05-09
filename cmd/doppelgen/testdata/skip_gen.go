// Package testdata — this file tests the doppel:skip-gen opt-out comment.
package testdata

// SkipGenStruct is annotated with doppel:skip-gen and should be excluded from generation.
//
// doppel:skip-gen
type SkipGenStruct struct {
	Name string `doppel:"deep"`
}

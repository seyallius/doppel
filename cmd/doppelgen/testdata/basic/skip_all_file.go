// Package testdata — this entire file is opted out of generation via doppel:skip-all.
// All struct types in this file should be skipped.
//
// doppel:skip-all
package basic

// SkippedInFile should be skipped because the file has doppel:skip-all.
type SkippedInFile struct {
	Name string `doppel:"deep"`
}

// AlsoSkippedInFile should also be skipped for the same reason.
type AlsoSkippedInFile struct {
	Value int `doppel:"deep"`
}

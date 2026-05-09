// doppel:skip-all
package testdata

// SkippedInFile should not be generated.
type SkippedInFile struct {
	Name  string `doppel:"deep"`
	Value int    `doppel:"deep"`
}

// AlsoSkippedInFile should not be generated either.
type AlsoSkippedInFile struct {
	Data []string `doppel:"deep"`
}

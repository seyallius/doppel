// Package testdata — unexported struct fixture.
package testdata

// unexportedStruct is not exported and its fields should not appear in generation.
type unexportedStruct struct {
	name  string
	Value int
}

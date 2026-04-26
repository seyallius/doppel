// Package doppel_test. doppel_bench_test - Exercises the full manual cloning stack
// via realistic domain types that demonstrate how CloneSlice, CloneMap, and ClonePointer
// compose inside a struct's own Clone() method.
package doppel_test

import (
	"testing"

	"github.com/seyallius/doppel"
	"github.com/seyallius/doppel/core"
)

// ---------------------------------------------------------------------------
// Benchmarks — manual deep copy vs shallow copy baseline
// ---------------------------------------------------------------------------

func BenchmarkManualClone_Address(b *testing.B) {
	// Address has only primitive fields: clone is a plain struct literal.
	src := *newAddress()
	cloner := core.NewFuncCloner(cloneAddress)
	b.ResetTimer()
	for b.Loop() {
		_, _ = doppel.CloneWith(src, cloner)
	}
}

func BenchmarkShallowCopy_Address(b *testing.B) {
	src := *newAddress()
	b.ResetTimer()
	for b.Loop() {
		dst := src
		_ = dst
	}
}

func BenchmarkManualClone_User(b *testing.B) {
	src := newUser()
	b.ResetTimer()
	for b.Loop() {
		_, _ = doppel.Clone(src)
	}
}

func BenchmarkShallowCopy_User(b *testing.B) {
	src := newUser()
	b.ResetTimer()
	for b.Loop() {
		dst := *src // shallow value copy
		_ = dst
	}
}

func BenchmarkManualClone_Order(b *testing.B) {
	src := newOrder(newUser())
	b.ResetTimer()
	for b.Loop() {
		_, _ = doppel.Clone(src)
	}
}

func BenchmarkShallowCopy_Order(b *testing.B) {
	src := newOrder(newUser())
	b.ResetTimer()
	for b.Loop() {
		dst := *src
		_ = dst
	}
}

func BenchmarkManualClone_UserLargeSlice(b *testing.B) {
	src := newUser()
	src.Tags = make([]string, 1000)
	for idx := range src.Tags {
		src.Tags[idx] = "tag"
	}
	b.ResetTimer()
	for b.Loop() {
		_, _ = doppel.Clone(src)
	}
}

func BenchmarkManualClone_UserLargeMap(b *testing.B) {
	src := newUser()
	src.Scores = make(map[string]int, 500)
	for idx := range 500 {
		src.Scores[key(idx)] = idx
	}
	b.ResetTimer()
	for b.Loop() {
		_, _ = doppel.Clone(src)
	}
}

func key(idx int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	return "key_" + string(alphabet[idx%26])
}

package doppel_test

import (
	"testing"

	"github.com/seyallius/doppel"
)

// --- Manual deep copy benchmarks vs shallow copy baseline --------------------

func BenchmarkManualClone_Address(b *testing.B) {
	cloner := newUser // just to avoid unused import; actual benchmark uses User
	_ = cloner
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = doppel.Clone(newUser())
	}
}

func BenchmarkShallowCopy_Address(b *testing.B) {
	src := *newAddress()
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		dst := src
		_ = dst
	}
}

func BenchmarkManualClone_User(b *testing.B) {
	src := newUser()
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = doppel.Clone(src)
	}
}

func BenchmarkShallowCopy_User(b *testing.B) {
	src := newUser()
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		dst := *src
		_ = dst
	}
}

func BenchmarkManualClone_Order(b *testing.B) {
	src := newOrder(newUser())
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = doppel.Clone(src)
	}
}

func BenchmarkShallowCopy_Order(b *testing.B) {
	src := newOrder(newUser())
	b.ResetTimer()
	b.ReportAllocs()
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
	b.ReportAllocs()
	for b.Loop() {
		_, _ = doppel.Clone(src)
	}
}

func BenchmarkManualClone_UserLargeMap(b *testing.B) {
	src := newUser()
	src.Scores = make(map[string]int, 500)
	for idx := 0; idx < 500; idx++ {
		src.Scores[benchKey(idx)] = idx
	}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = doppel.Clone(src)
	}
}

// -------------------------------------------- Internal Helpers --------------------------------------------

func benchKey(idx int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	return "key_" + string(alphabet[idx%26])
}

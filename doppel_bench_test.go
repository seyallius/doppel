package doppel_test

import (
	"testing"

	"github.com/seyallius/doppel"
	"github.com/seyallius/doppel/core"
	"github.com/seyallius/doppel/registry"
)

// --- Phase 1 — Manual deep copy vs shallow copy baseline --------------------

func BenchmarkManualClone_Address(b *testing.B) {
	// Address has only primitive fields: clone is a plain struct literal.
	src := *newAddress()
	cloner := core.NewFuncCloner(cloneAddress)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = doppel.CloneWith(src, cloner)
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
		dst := *src // shallow value copy
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
	for idx := range 500 {
		src.Scores[benchKey(idx)] = idx
	}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = doppel.Clone(src)
	}
}

// --- Phase 2 — Registry cloner vs SelfClonable fallback vs shallow baseline --------------------

// BenchmarkRegistry_RegisteredCloner measures the hot path: type is in the
// registry, registry lookup + clone dispatched directly to the Cloner[T].
func BenchmarkRegistry_RegisteredCloner_User(b *testing.B) {
	reg := registry.New()
	registry.Register(reg, core.NewFuncCloner(func(src User) (User, error) {
		return *newUser(), nil // fast fixed-allocation stand-in
	}))

	src := newUser()
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = doppel.CloneWithRegistry(*src, reg)
	}
}

// BenchmarkRegistry_SelfClonableFallback measures the cold path: type is NOT in
// the registry, so CloneWithRegistry falls through to SelfClonable.Clone().
func BenchmarkRegistry_SelfClonableFallback_User(b *testing.B) {
	emptyReg := registry.New() // nothing registered
	src := newUser()
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = doppel.CloneWithRegistry(src, emptyReg)
	}
}

// BenchmarkRegistry_DirectClone is the baseline: CloneWithRegistry vs
// the equivalent direct doppel.Clone call.
func BenchmarkRegistry_DirectClone_User(b *testing.B) {
	src := newUser()
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = doppel.Clone(src)
	}
}

// BenchmarkRegistry_LookupOverhead isolates the registry map lookup cost
// using a no-op clone function so we measure only the dispatch overhead.
func BenchmarkRegistry_LookupOverhead(b *testing.B) {
	reg := registry.New()
	registry.Register(reg, core.NewFuncCloner(func(src Address) (Address, error) {
		return src, nil // identity: measures dispatch cost only
	}))

	src := *newAddress()
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = doppel.CloneWithRegistry(src, reg)
	}
}

// BenchmarkRegistry_ShallowBaseline is the floor for the above benchmarks.
func BenchmarkRegistry_ShallowBaseline(b *testing.B) {
	src := *newAddress()
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		dst := src
		_ = dst
	}
}

// -------------------------------------------- Internal Helpers --------------------------------------------

func benchKey(idx int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	return "key_" + string(alphabet[idx%26])
}

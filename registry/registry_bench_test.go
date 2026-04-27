package registry_test

import (
	"testing"

	"github.com/seyallius/doppel/core"
	"github.com/seyallius/doppel/registry"
)

func BenchmarkRegistry_Register(b *testing.B) {
	cloner := core.NewFuncCloner(cloneTypeA)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		reg := registry.New()
		registry.Register(reg, cloner)
	}
}

func BenchmarkRegistry_Lookup_Hit(b *testing.B) {
	reg := registry.New()
	registry.Register(reg, core.NewFuncCloner(cloneTypeA))
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = registry.Lookup[TypeA](reg)
	}
}

func BenchmarkRegistry_Lookup_Miss(b *testing.B) {
	reg := registry.New() // empty
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = registry.Lookup[TypeA](reg)
	}
}

func BenchmarkRegistry_Has(b *testing.B) {
	reg := registry.New()
	registry.Register(reg, core.NewFuncCloner(cloneTypeA))
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = registry.Has[TypeA](reg)
	}
}

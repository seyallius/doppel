package doppel_test

import (
	"testing"

	"github.com/seyallius/doppel"
	"github.com/seyallius/doppel/core"
	"github.com/seyallius/doppel/registry"
)

// --- Phase 3 — CloneDeep benchmarks --------------------

func BenchmarkCloneDeep_PureReflection_DeepUser(b *testing.B) {
	src := deepUser{
		ID:      1,
		Name:    "bench",
		Active:  true,
		Address: &deepAddress{Street: "1 Main", City: "Metro", State: "ST", Zip: "00000"},
		Tags:    []string{"a", "b", "c"},
		Scores:  map[string]int{"x": 1, "y": 2},
	}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = doppel.CloneDeep(src, nil)
	}
}

func BenchmarkCloneDeep_WithFieldCloners_DeepUser(b *testing.B) {
	reg := registry.New()
	registry.RegisterField[deepUser, *deepAddress](reg, "Address", core.NewFuncCloner(
		func(src *deepAddress) (*deepAddress, error) {
			return &deepAddress{Street: src.Street, City: src.City, State: src.State, Zip: src.Zip}, nil
		},
	))
	registry.RegisterField[deepUser, []string](reg, "Tags", core.NewFuncCloner(
		func(src []string) ([]string, error) {
			return append([]string{}, src...), nil
		},
	))

	src := deepUser{
		ID:      1,
		Name:    "bench",
		Active:  true,
		Address: &deepAddress{Street: "1 Main", City: "Metro", State: "ST", Zip: "00000"},
		Tags:    []string{"a", "b", "c"},
		Scores:  map[string]int{"x": 1, "y": 2},
	}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = doppel.CloneDeep(src, reg)
	}
}

func BenchmarkCloneDeep_WithTypeCloner_DeepUser(b *testing.B) {
	reg := registry.New()
	registry.Register(reg, core.NewFuncCloner(func(src deepUser) (deepUser, error) {
		tags := make([]string, len(src.Tags))
		copy(tags, src.Tags)
		scores := make(map[string]int, len(src.Scores))
		for k, v := range src.Scores {
			scores[k] = v
		}
		var addr *deepAddress
		if src.Address != nil {
			a := *src.Address
			addr = &a
		}
		return deepUser{ID: src.ID, Name: src.Name, Active: src.Active, Address: addr, Tags: tags, Scores: scores}, nil
	}))

	src := deepUser{
		ID:      1,
		Name:    "bench",
		Active:  true,
		Address: &deepAddress{Street: "1 Main", City: "Metro", State: "ST", Zip: "00000"},
		Tags:    []string{"a", "b", "c"},
		Scores:  map[string]int{"x": 1, "y": 2},
	}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = doppel.CloneDeep(src, reg)
	}
}

func BenchmarkCloneDeep_SelfClonable(b *testing.B) {
	src := &selfClonableUser{
		ID:   1,
		Name: "bench",
		Tags: []string{"a", "b", "c"},
	}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = doppel.CloneDeep(src, nil)
	}
}

func BenchmarkCloneDeep_BigModel(b *testing.B) {
	src := bigModel{
		ID:       42,
		Name:     "BigBench",
		Active:   true,
		Priority: 5,
		Weight:   3.14,
		Metadata: map[string]string{"env": "prod", "region": "us-east"},
		Address:  &deepAddress{Street: "1 Main", City: "Metro", State: "ST", Zip: "00000"},
		Tags:     []string{"critical", "monitored"},
	}

	b.Run("PureReflection", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = doppel.CloneDeep(src, nil)
		}
	})

	b.Run("WithFieldCloners", func(b *testing.B) {
		reg := registry.New()
		registry.RegisterField[bigModel, *deepAddress](reg, "Address", core.NewFuncCloner(
			func(src *deepAddress) (*deepAddress, error) {
				return &deepAddress{Street: src.Street, City: src.City, State: src.State, Zip: src.Zip}, nil
			},
		))
		registry.RegisterField[bigModel, []string](reg, "Tags", core.NewFuncCloner(
			func(src []string) ([]string, error) { return append([]string{}, src...), nil },
		))

		b.ResetTimer()
		b.ReportAllocs()
		for b.Loop() {
			_, _ = doppel.CloneDeep(src, reg)
		}
	})

	b.Run("WithManualClone", func(b *testing.B) {
		// Baseline: hand-written manual clone for comparison.
		b.ResetTimer()
		b.ReportAllocs()
		for b.Loop() {
			tags := make([]string, len(src.Tags))
			copy(tags, src.Tags)
			metadata := make(map[string]string, len(src.Metadata))
			for k, v := range src.Metadata {
				metadata[k] = v
			}
			var addr *deepAddress
			if src.Address != nil {
				a := *src.Address
				addr = &a
			}
			_ = bigModel{
				ID: src.ID, Name: src.Name, Active: src.Active,
				Priority: src.Priority, Weight: src.Weight,
				Metadata: metadata, Address: addr, Tags: tags,
			}
		}
	})
}

func BenchmarkCloneDeep_ShallowBaseline(b *testing.B) {
	src := deepUser{
		ID:      1,
		Name:    "bench",
		Address: &deepAddress{City: "Metro"},
		Tags:    []string{"a", "b"},
	}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		dst := src
		_ = dst
	}
}

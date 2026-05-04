package engine_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/seyallius/doppel/core"
	"github.com/seyallius/doppel/engine"
	"github.com/seyallius/doppel/registry"
)

// --- Fixture types for field-level cloner tests --------------------

// fieldHost is a struct that will host field-level cloners.
type fieldHost struct {
	Name    string
	Value   int
	Nested  *fieldNested
	Tags    []string
	Scores  map[string]int
	Skipped string // unexported — should be skipped by engine
}

// fieldNested is a nested struct used to test pointer field cloners.
type fieldNested struct {
	Label string
	Count int
}

// withCloneTag uses the doppel:"clone" tag to require a field cloner.
type withCloneTag struct {
	Name   string
	Nested *fieldNested `doppel:"clone"`
}

// withDeepTag uses the doppel:"deep" tag for explicit deep copy.
type withDeepTag struct {
	Name  string
	Tags  []string `doppel:"deep"`
	Value int
}

// withReadonlyTag uses the doppel:"readonly" tag.
type withReadonlyTag struct {
	Name   string
	Config map[string]string `doppel:"readonly"`
}

// bigStruct simulates a struct with many fields where only a few need custom cloning.
type bigStruct struct {
	ID       int64
	Label    string
	Active   bool
	Priority int
	Weight   float64
	Address  *fieldNested // needs custom cloning
	Tags     []string     // needs custom cloning
	Scores   map[string]int
	// ... imagine 190 more primitive fields here
}

func cloneFieldNested(src *fieldNested) (*fieldNested, error) {
	if src == nil {
		return nil, nil
	}
	return &fieldNested{Label: src.Label + "_field", Count: src.Count}, nil
}

// --- Engine field cloner auto-discovery tests --------------------

func TestEngine_FieldCloner_AutoDiscovery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		src       fieldHost
		setupReg  func(*registry.Registry)
		wantLabel string // expected Label on cloned Nested field
	}{
		{
			name: "field_cloner_used_for_registered_field",
			src: fieldHost{
				Name:   "test",
				Value:  42,
				Nested: &fieldNested{Label: "inner", Count: 7},
			},
			setupReg: func(r *registry.Registry) {
				registry.RegisterField[fieldHost, *fieldNested](r, "Nested", core.NewFuncCloner(cloneFieldNested))
			},
			wantLabel: "inner_field",
		},
		{
			name: "no_field_cloner_uses_default_reflection",
			src: fieldHost{
				Name:   "test",
				Value:  42,
				Nested: &fieldNested{Label: "inner", Count: 7},
			},
			setupReg:  func(_ *registry.Registry) {},
			wantLabel: "inner", // default reflection copies as-is
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reg := registry.New()
			tc.setupReg(reg)

			eng := engine.New(reg)
			clonedVal, err := eng.Clone(reflect.ValueOf(tc.src))
			requireNoError(t, err)

			cloned := clonedVal.Interface().(fieldHost)
			if cloned.Name != tc.src.Name {
				t.Errorf("Name: got %q, want %q", cloned.Name, tc.src.Name)
			}
			if cloned.Value != tc.src.Value {
				t.Errorf("Value: got %d, want %d", cloned.Value, tc.src.Value)
			}
			if cloned.Nested == nil {
				t.Fatal("Nested is nil in clone")
			}
			if cloned.Nested.Label != tc.wantLabel {
				t.Errorf("Nested.Label: got %q, want %q", cloned.Nested.Label, tc.wantLabel)
			}
		})
	}
}

func TestEngine_FieldCloner_PrimitiveFieldsUntouched(t *testing.T) {
	t.Parallel()

	// Register field cloner for only the "Nested" field; all other fields
	// should be cloned by default reflection.
	reg := registry.New()
	registry.RegisterField[fieldHost, *fieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))

	src := fieldHost{
		Name:   "primitive_test",
		Value:  99,
		Nested: &fieldNested{Label: "nested", Count: 3},
		Tags:   []string{"a", "b"},
		Scores: map[string]int{"x": 10},
	}

	eng := engine.New(reg)
	clonedVal, err := eng.Clone(reflect.ValueOf(src))
	requireNoError(t, err)

	cloned := clonedVal.Interface().(fieldHost)

	// Primitives should be copied correctly.
	if cloned.Name != src.Name {
		t.Errorf("Name: got %q, want %q", cloned.Name, src.Name)
	}
	if cloned.Value != src.Value {
		t.Errorf("Value: got %d, want %d", cloned.Value, src.Value)
	}

	// Field cloner should have been used for Nested.
	if cloned.Nested.Label != "nested_field" {
		t.Errorf("Nested.Label: got %q, want %q", cloned.Nested.Label, "nested_field")
	}

	// Tags and Scores should be cloned via default reflection.
	if len(cloned.Tags) != 2 || cloned.Tags[0] != "a" {
		t.Errorf("Tags: got %v", cloned.Tags)
	}
	if cloned.Scores["x"] != 10 {
		t.Errorf("Scores[x]: got %d, want 10", cloned.Scores["x"])
	}
}

func TestEngine_FieldCloner_Independence(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	registry.RegisterField[fieldHost, *fieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))

	src := fieldHost{
		Name:   "indep_test",
		Nested: &fieldNested{Label: "orig", Count: 1},
		Tags:   []string{"tag1"},
	}

	eng := engine.New(reg)
	clonedVal, err := eng.Clone(reflect.ValueOf(src))
	requireNoError(t, err)

	cloned := clonedVal.Interface().(fieldHost)

	// Mutating original should not affect clone.
	src.Nested.Label = "mutated"
	if cloned.Nested.Label == "mutated" {
		t.Error("clone shares memory with original Nested")
	}

	src.Tags[0] = "mutated_tag"
	if cloned.Tags[0] == "mutated_tag" {
		t.Error("clone shares memory with original Tags")
	}
}

func TestEngine_FieldCloner_NilPointerField(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	registry.RegisterField[fieldHost, *fieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))

	src := fieldHost{
		Name:   "nil_nested",
		Nested: nil,
	}

	eng := engine.New(reg)
	clonedVal, err := eng.Clone(reflect.ValueOf(src))
	requireNoError(t, err)

	cloned := clonedVal.Interface().(fieldHost)
	if cloned.Nested != nil {
		t.Error("nil Nested should remain nil in clone")
	}
}

func TestEngine_FieldCloner_ErrorPropagation(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("field clone failed")
	reg := registry.New()
	registry.RegisterField[fieldHost, *fieldNested](reg, "Nested", core.NewFuncCloner(
		func(src *fieldNested) (*fieldNested, error) {
			return nil, sentinel
		},
	))

	src := fieldHost{
		Name:   "error_test",
		Nested: &fieldNested{Label: "fail"},
	}

	eng := engine.New(reg)
	_, err := eng.Clone(reflect.ValueOf(src))
	if err == nil {
		t.Fatal("expected error from field cloner, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("errors.Is failed: got %v", err)
	}
}

// --- doppel:"clone" tag tests --------------------

func TestEngine_CloneTag_RequiresFieldCloner(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		setupReg func(*registry.Registry)
		wantErr  bool
	}{
		{
			name: "with_registered_field_cloner_succeeds",
			setupReg: func(r *registry.Registry) {
				registry.RegisterField[withCloneTag, *fieldNested](r, "Nested", core.NewFuncCloner(cloneFieldNested))
			},
			wantErr: false,
		},
		{
			name:     "without_registered_field_cloner_errors",
			setupReg: func(_ *registry.Registry) {},
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reg := registry.New()
			tc.setupReg(reg)

			src := withCloneTag{
				Name:   "tagged",
				Nested: &fieldNested{Label: "inner", Count: 5},
			}

			eng := engine.New(reg)
			_, err := eng.Clone(reflect.ValueOf(src))

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error for doppel:\"clone\" without field cloner, got nil")
				}
				return
			}

			requireNoError(t, err)
		})
	}
}

func TestEngine_CloneTag_WithoutFieldLookup_Errors(t *testing.T) {
	t.Parallel()

	// When no FieldLookup is available at all (nil lookup), doppel:"clone" should error.
	src := withCloneTag{
		Name:   "no_lookup",
		Nested: &fieldNested{Label: "inner"},
	}

	eng := engine.New(nil)
	_, err := eng.Clone(reflect.ValueOf(src))
	if err == nil {
		t.Fatal("expected error for doppel:\"clone\" without FieldLookup, got nil")
	}
}

// --- doppel:"deep" tag tests --------------------

func TestEngine_DeepTag_ExplicitDeepCopy(t *testing.T) {
	t.Parallel()

	src := withDeepTag{
		Name:  "deep_test",
		Tags:  []string{"a", "b", "c"},
		Value: 42,
	}

	eng := engine.New(nil)
	clonedVal, err := eng.Clone(reflect.ValueOf(src))
	requireNoError(t, err)

	cloned := clonedVal.Interface().(withDeepTag)

	// Verify deep copy.
	if cloned.Name != src.Name {
		t.Errorf("Name: got %q, want %q", cloned.Name, src.Name)
	}
	if len(cloned.Tags) != 3 {
		t.Errorf("Tags len: got %d, want 3", len(cloned.Tags))
	}

	// Verify independence.
	src.Tags[0] = "mutated"
	if cloned.Tags[0] == "mutated" {
		t.Error("doppel:\"deep\" field shares memory with original")
	}
}

// --- doppel:"readonly" tag tests --------------------

func TestEngine_ReadonlyTag_SharesBacking(t *testing.T) {
	t.Parallel()

	src := withReadonlyTag{
		Name:   "readonly_test",
		Config: map[string]string{"key": "value"},
	}

	eng := engine.New(nil)
	clonedVal, err := eng.Clone(reflect.ValueOf(src))
	requireNoError(t, err)

	cloned := clonedVal.Interface().(withReadonlyTag)

	// Readonly should share the same map (shallow copy).
	src.Config["key"] = "mutated"
	if cloned.Config["key"] != "mutated" {
		t.Error("doppel:\"readonly\" field should share backing with original")
	}
}

// --- BigStruct scenario — the core Phase 3 use case --------------------

func TestEngine_BigStruct_SelectiveOverride(t *testing.T) {
	t.Parallel()

	// This is the primary Phase 3 use case: a struct with many fields
	// where only a few need custom clone logic. Instead of writing a
	// 200-field Clone() method, register field cloners for just the
	// fields that need custom handling.
	reg := registry.New()

	// Custom clone for the Address pointer field
	registry.RegisterField[bigStruct, *fieldNested](reg, "Address", core.NewFuncCloner(cloneFieldNested))

	// Custom clone for the Tags slice field — ensure independence
	registry.RegisterField[bigStruct, []string](reg, "Tags", core.NewFuncCloner(
		func(src []string) ([]string, error) {
			return append([]string{}, src...), nil
		},
	))

	src := bigStruct{
		ID:       1,
		Label:    "big_struct",
		Active:   true,
		Priority: 5,
		Weight:   3.14,
		Address:  &fieldNested{Label: "addr", Count: 10},
		Tags:     []string{"important", "urgent"},
		Scores:   map[string]int{"quality": 95},
	}

	eng := engine.New(reg)
	clonedVal, err := eng.Clone(reflect.ValueOf(src))
	requireNoError(t, err)

	cloned := clonedVal.Interface().(bigStruct)

	// Primitives cloned by default reflection.
	if cloned.ID != src.ID {
		t.Errorf("ID: got %d, want %d", cloned.ID, src.ID)
	}
	if cloned.Label != src.Label {
		t.Errorf("Label: got %q, want %q", cloned.Label, src.Label)
	}
	if cloned.Active != src.Active {
		t.Errorf("Active: got %v, want %v", cloned.Active, src.Active)
	}
	if cloned.Priority != src.Priority {
		t.Errorf("Priority: got %d, want %d", cloned.Priority, src.Priority)
	}
	if cloned.Weight != src.Weight {
		t.Errorf("Weight: got %f, want %f", cloned.Weight, src.Weight)
	}

	// Address field used custom cloner.
	if cloned.Address == nil {
		t.Fatal("Address is nil in clone")
	}
	if cloned.Address.Label != "addr_field" {
		t.Errorf("Address.Label: got %q, want %q", cloned.Address.Label, "addr_field")
	}

	// Tags field used custom cloner.
	if len(cloned.Tags) != 2 {
		t.Fatalf("Tags len: got %d, want 2", len(cloned.Tags))
	}

	// Scores cloned by default reflection.
	if cloned.Scores["quality"] != 95 {
		t.Errorf("Scores[quality]: got %d, want 95", cloned.Scores["quality"])
	}

	// Independence checks.
	src.Address.Label = "mutated"
	if cloned.Address.Label == "mutated" {
		t.Error("Address not independent from original")
	}

	src.Tags[0] = "mutated_tag"
	if cloned.Tags[0] == "mutated_tag" {
		t.Error("Tags not independent from original")
	}

	src.Scores["quality"] = 0
	if cloned.Scores["quality"] == 0 {
		t.Error("Scores not independent from original")
	}
}

// --- Field cloner + type cloner priority --------------------

func TestEngine_FieldCloner_PriorityOverTypeCloner(t *testing.T) {
	t.Parallel()

	// Register both a type-level cloner and a field-level cloner for *fieldNested.
	// The field-level cloner should take priority for the specific field.
	reg := registry.New()

	// Type-level cloner — appends "_type"
	registry.Register(reg, core.NewFuncCloner(func(src *fieldNested) (*fieldNested, error) {
		return &fieldNested{Label: src.Label + "_type", Count: src.Count}, nil
	}))

	// Field-level cloner — appends "_field"
	registry.RegisterField[fieldHost, *fieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))

	src := fieldHost{
		Name:   "priority_test",
		Nested: &fieldNested{Label: "inner", Count: 1},
	}

	eng := engine.New(reg)
	clonedVal, err := eng.Clone(reflect.ValueOf(src))
	requireNoError(t, err)

	cloned := clonedVal.Interface().(fieldHost)

	// Field cloner should win over type cloner for the specific field.
	if cloned.Nested.Label != "inner_field" {
		t.Errorf("Field cloner should take priority: got %q, want %q",
			cloned.Nested.Label, "inner_field")
	}
}

func TestEngine_FieldCloner_TypeClonerUsedForOtherFields(t *testing.T) {
	t.Parallel()

	// Register a type cloner for *fieldNested and a field cloner only for
	// fieldHost.Nested. Other structs with *fieldNested fields should
	// use the type cloner.
	reg := registry.New()

	// Type-level cloner
	registry.Register(reg, core.NewFuncCloner(func(src *fieldNested) (*fieldNested, error) {
		return &fieldNested{Label: src.Label + "_type", Count: src.Count}, nil
	}))

	// Field-level cloner only for fieldHost.Nested
	registry.RegisterField[fieldHost, *fieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))

	// Struct without field-level registration
	type otherOwner struct {
		Item *fieldNested
	}

	src := otherOwner{Item: &fieldNested{Label: "other", Count: 2}}

	eng := engine.New(reg)
	clonedVal, err := eng.Clone(reflect.ValueOf(src))
	requireNoError(t, err)

	cloned := clonedVal.Interface().(otherOwner)

	// Should use the type cloner, not any field cloner.
	if cloned.Item.Label != "other_type" {
		t.Errorf("Type cloner should be used for non-registered fields: got %q, want %q",
			cloned.Item.Label, "other_type")
	}
}

// --- Multiple field cloners on same struct --------------------

func TestEngine_MultipleFieldCloners(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	registry.RegisterField[fieldHost, *fieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))
	registry.RegisterField[fieldHost, []string](reg, "Tags", core.NewFuncCloner(
		func(src []string) ([]string, error) {
			result := make([]string, len(src))
			for i, s := range src {
				result[i] = s + "_cloned"
			}
			return result, nil
		},
	))

	src := fieldHost{
		Name:   "multi_field",
		Nested: &fieldNested{Label: "inner", Count: 3},
		Tags:   []string{"a", "b"},
	}

	eng := engine.New(reg)
	clonedVal, err := eng.Clone(reflect.ValueOf(src))
	requireNoError(t, err)

	cloned := clonedVal.Interface().(fieldHost)

	// Nested should use field cloner.
	if cloned.Nested.Label != "inner_field" {
		t.Errorf("Nested.Label: got %q, want %q", cloned.Nested.Label, "inner_field")
	}

	// Tags should use field cloner.
	if len(cloned.Tags) != 2 {
		t.Fatalf("Tags len: got %d, want 2", len(cloned.Tags))
	}
	if cloned.Tags[0] != "a_cloned" {
		t.Errorf("Tags[0]: got %q, want %q", cloned.Tags[0], "a_cloned")
	}
	if cloned.Tags[1] != "b_cloned" {
		t.Errorf("Tags[1]: got %q, want %q", cloned.Tags[1], "b_cloned")
	}
}

// --- DeregisterField stops field cloner from being used --------------------

func TestEngine_DeregisterField_FallsBackToReflection(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	registry.RegisterField[fieldHost, *fieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))

	src := fieldHost{
		Name:   "dereg_test",
		Nested: &fieldNested{Label: "inner", Count: 1},
	}

	// Before deregister: field cloner is used.
	eng := engine.New(reg)
	clonedVal1, err := eng.Clone(reflect.ValueOf(src))
	requireNoError(t, err)
	cloned1 := clonedVal1.Interface().(fieldHost)
	if cloned1.Nested.Label != "inner_field" {
		t.Errorf("before deregister: got %q, want %q", cloned1.Nested.Label, "inner_field")
	}

	// After deregister: default reflection is used.
	registry.DeregisterField[fieldHost](reg, "Nested")
	eng2 := engine.New(reg)
	clonedVal2, err := eng2.Clone(reflect.ValueOf(src))
	requireNoError(t, err)
	cloned2 := clonedVal2.Interface().(fieldHost)
	if cloned2.Nested.Label != "inner" {
		t.Errorf("after deregister: got %q, want %q (default reflection)", cloned2.Nested.Label, "inner")
	}
}

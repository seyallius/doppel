package registry_test

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/seyallius/doppel/core"
	"github.com/seyallius/doppel/registry"
)

// --- Fixture types for field-level cloner tests --------------------

type (
	// FieldHost is a struct that will host field-level cloners.
	FieldHost struct {
		Name    string
		Value   int
		Nested  *FieldNested
		Tags    []string
		Scores  map[string]int
		Skipped string
	}

	// FieldNested is a nested struct used to test pointer field cloners.
	FieldNested struct {
		Label string
		Count int
	}

	// AnotherHost has the same field name "Nested" but a different struct type.
	AnotherHost struct {
		Nested *FieldNested
		Extra  string
	}
)

func cloneFieldNested(src *FieldNested) (*FieldNested, error) {
	if src == nil {
		return nil, nil
	}
	return &FieldNested{Label: src.Label + "_cloned", Count: src.Count}, nil
}

// --- RegisterField --------------------

func TestRegisterField(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "registered_field_cloner_is_discoverable_via_has_field",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))
				if !registry.HasField[FieldHost](reg, "Nested") {
					t.Error("HasField[FieldHost](\"Nested\") returned false after RegisterField")
				}
			},
		},
		{
			name: "field_len_increments_on_new_field",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))
				if reg.FieldLen() != 1 {
					t.Errorf("FieldLen after one RegisterField: got %d, want 1", reg.FieldLen())
				}
				registry.RegisterField[FieldHost, []string](reg, "Tags", core.NewFuncCloner(
					func(src []string) ([]string, error) { return append([]string{}, src...), nil },
				))
				if reg.FieldLen() != 2 {
					t.Errorf("FieldLen after two RegisterFields: got %d, want 2", reg.FieldLen())
				}
			},
		},
		{
			name: "registering_same_field_twice_does_not_grow_field_len",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))
				if reg.FieldLen() != 1 {
					t.Errorf("FieldLen after double-register: got %d, want 1", reg.FieldLen())
				}
			},
		},
		{
			name: "same_field_name_different_struct_types_are_independent",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))
				registry.RegisterField[AnotherHost, *FieldNested](reg, "Nested", core.NewFuncCloner(
					func(src *FieldNested) (*FieldNested, error) {
						return &FieldNested{Label: "another_" + src.Label, Count: src.Count}, nil
					},
				))

				if !registry.HasField[FieldHost](reg, "Nested") {
					t.Error("FieldHost.Nested should be registered")
				}
				if !registry.HasField[AnotherHost](reg, "Nested") {
					t.Error("AnotherHost.Nested should be registered")
				}
				if reg.FieldLen() != 2 {
					t.Errorf("FieldLen: got %d, want 2", reg.FieldLen())
				}
			},
		},
		{
			name: "register_field_with_pointer_struct_type_works",
			check: func(t *testing.T) {
				reg := registry.New()
				// RegisterField[*FieldHost, ...] should resolve to the same key as RegisterField[FieldHost, ...]
				registry.RegisterField[*FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))
				if !registry.HasField[FieldHost](reg, "Nested") {
					t.Error("RegisterField with *FieldHost should resolve to FieldHost")
				}
			},
		},
		{
			name: "type_and_field_cloners_coexist",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.Register(reg, core.NewFuncCloner(cloneTypeA))
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))

				if reg.Len() != 1 {
					t.Errorf("Len: got %d, want 1", reg.Len())
				}
				if reg.FieldLen() != 1 {
					t.Errorf("FieldLen: got %d, want 1", reg.FieldLen())
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.check(t)
		})
	}
}

func TestRegisterField_Panics(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		register  func()
		wantPanic string
	}{
		{
			name: "non_existent_field_panics",
			register: func() {
				reg := registry.New()
				registry.RegisterField[FieldHost, int](reg, "NonExistent", core.NewFuncCloner(
					func(src int) (int, error) { return src, nil },
				))
			},
			wantPanic: "has no field named",
		},
		{
			name: "unexported_field_panics",
			register: func() {
				reg := registry.New()
				registry.RegisterField[FieldHost, string](reg, "Skipped", core.NewFuncCloner(
					func(src string) (string, error) { return src, nil },
				))
			},
			wantPanic: "unexported",
		},
		{
			name: "non_struct_type_panics",
			register: func() {
				reg := registry.New()
				registry.RegisterField[int, int](reg, "invalid", core.NewFuncCloner(
					func(src int) (int, error) { return src, nil },
				))
			},
			wantPanic: "must be a struct type",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				rec := recover()
				if rec == nil {
					t.Fatal("expected panic, got none")
				}
				msg := rec.(string)
				if !strings.Contains(msg, tc.wantPanic) {
					t.Errorf("panic message: got %q, want substring %q", msg, tc.wantPanic)
				}
			}()

			tc.register()
		})
	}
}

// --- LookupField --------------------

func TestLookupField(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "lookup_returns_false_for_unregistered_field",
			check: func(t *testing.T) {
				reg := registry.New()
				cloner, found := registry.LookupField[FieldHost, *FieldNested](reg, "Nested")
				if found {
					t.Error("LookupField returned true for unregistered field")
				}
				if cloner != nil {
					t.Error("LookupField returned non-nil cloner for unregistered field")
				}
			},
		},
		{
			name: "lookup_returns_true_for_registered_field",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))
				cloner, found := registry.LookupField[FieldHost, *FieldNested](reg, "Nested")
				if !found {
					t.Fatal("LookupField returned false for registered field")
				}
				if cloner == nil {
					t.Error("LookupField returned nil cloner for registered field")
				}
			},
		},
		{
			name: "looked_up_field_cloner_produces_correct_result",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))

				cloner, found := registry.LookupField[FieldHost, *FieldNested](reg, "Nested")
				if !found {
					t.Fatal("cloner not found")
				}

				src := &FieldNested{Label: "test", Count: 42}
				got, err := cloner.Clone(src)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.Label != "test_cloned" {
					t.Errorf("cloner result: got Label=%q, want %q", got.Label, "test_cloned")
				}
			},
		},
		{
			name: "lookup_does_not_find_different_field_name",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))

				_, found := registry.LookupField[FieldHost, *FieldNested](reg, "Tags")
				if found {
					t.Error("LookupField[FieldHost](\"Tags\") returned true when only Nested is registered")
				}
			},
		},
		{
			name: "second_register_replaces_first",
			check: func(t *testing.T) {
				reg := registry.New()

				// First: multiplies Count by 10.
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))
				// Second: multiplies Count by 100.
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(
					func(src *FieldNested) (*FieldNested, error) {
						return &FieldNested{Label: src.Label, Count: src.Count * 100}, nil
					},
				))

				cloner, _ := registry.LookupField[FieldHost, *FieldNested](reg, "Nested")
				got, _ := cloner.Clone(&FieldNested{Count: 3})
				if got.Count != 300 {
					t.Errorf("expected second field cloner to be active (got %d, want 300)", got.Count)
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.check(t)
		})
	}
}

// --- DeregisterField --------------------

func TestDeregisterField(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "deregister_removes_registered_field_cloner",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))
				removed := registry.DeregisterField[FieldHost](reg, "Nested")

				if !removed {
					t.Error("DeregisterField should return true for existing registration")
				}
				if registry.HasField[FieldHost](reg, "Nested") {
					t.Error("HasField returned true after DeregisterField")
				}
				if reg.FieldLen() != 0 {
					t.Errorf("FieldLen after DeregisterField: got %d, want 0", reg.FieldLen())
				}
			},
		},
		{
			name: "deregister_unregistered_field_returns_false",
			check: func(t *testing.T) {
				reg := registry.New()
				removed := registry.DeregisterField[FieldHost](reg, "Nested")

				if removed {
					t.Error("DeregisterField should return false for non-existent registration")
				}
			},
		},
		{
			name: "deregister_only_removes_target_field",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))
				registry.RegisterField[FieldHost, []string](reg, "Tags", core.NewFuncCloner(
					func(src []string) ([]string, error) { return append([]string{}, src...), nil },
				))
				registry.DeregisterField[FieldHost](reg, "Nested")

				if registry.HasField[FieldHost](reg, "Nested") {
					t.Error("Nested should be removed")
				}
				if !registry.HasField[FieldHost](reg, "Tags") {
					t.Error("Tags should still be registered")
				}
			},
		},
		{
			name: "can_re_register_after_deregister",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))
				registry.DeregisterField[FieldHost](reg, "Nested")
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(
					func(src *FieldNested) (*FieldNested, error) {
						return &FieldNested{Label: src.Label + "_rereg", Count: src.Count}, nil
					},
				))

				cloner, found := registry.LookupField[FieldHost, *FieldNested](reg, "Nested")
				if !found {
					t.Fatal("re-registered field cloner not found")
				}
				got, _ := cloner.Clone(&FieldNested{Label: "test"})
				if got.Label != "test_rereg" {
					t.Errorf("re-registered cloner: got %q, want %q", got.Label, "test_rereg")
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.check(t)
		})
	}
}

// --- HasField --------------------

func TestHasField(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		setup      func(*registry.Registry)
		wantNested bool
		wantTags   bool
		wantExtra  bool
	}{
		{
			name:       "empty_registry_has_no_field_cloners",
			setup:      func(_ *registry.Registry) {},
			wantNested: false,
			wantTags:   false,
			wantExtra:  false,
		},
		{
			name: "has_only_registered_fields",
			setup: func(r *registry.Registry) {
				registry.RegisterField[FieldHost, *FieldNested](r, "Nested", core.NewFuncCloner(cloneFieldNested))
			},
			wantNested: true,
			wantTags:   false,
			wantExtra:  false,
		},
		{
			name: "has_multiple_registered_fields",
			setup: func(r *registry.Registry) {
				registry.RegisterField[FieldHost, *FieldNested](r, "Nested", core.NewFuncCloner(cloneFieldNested))
				registry.RegisterField[FieldHost, []string](r, "Tags", core.NewFuncCloner(
					func(src []string) ([]string, error) { return append([]string{}, src...), nil },
				))
			},
			wantNested: true,
			wantTags:   true,
			wantExtra:  false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reg := registry.New()
			tc.setup(reg)

			if registry.HasField[FieldHost](reg, "Nested") != tc.wantNested {
				t.Errorf("HasField[Nested]: got %v, want %v", registry.HasField[FieldHost](reg, "Nested"), tc.wantNested)
			}
			if registry.HasField[FieldHost](reg, "Tags") != tc.wantTags {
				t.Errorf("HasField[Tags]: got %v, want %v", registry.HasField[FieldHost](reg, "Tags"), tc.wantTags)
			}
			if registry.HasField[FieldHost](reg, "Extra") != tc.wantExtra {
				t.Errorf("HasField[Extra]: got %v, want %v", registry.HasField[FieldHost](reg, "Extra"), tc.wantExtra)
			}
		})
	}
}

// --- FieldLen --------------------

func TestFieldLen(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		actions func(*registry.Registry)
		wantLen int
	}{
		{
			name:    "empty_registry",
			actions: func(_ *registry.Registry) {},
			wantLen: 0,
		},
		{
			name: "one_field_registration",
			actions: func(r *registry.Registry) {
				registry.RegisterField[FieldHost, *FieldNested](r, "Nested", core.NewFuncCloner(cloneFieldNested))
			},
			wantLen: 1,
		},
		{
			name: "two_distinct_fields",
			actions: func(r *registry.Registry) {
				registry.RegisterField[FieldHost, *FieldNested](r, "Nested", core.NewFuncCloner(cloneFieldNested))
				registry.RegisterField[FieldHost, []string](r, "Tags", core.NewFuncCloner(
					func(src []string) ([]string, error) { return append([]string{}, src...), nil },
				))
			},
			wantLen: 2,
		},
		{
			name: "same_field_on_different_struct_types",
			actions: func(r *registry.Registry) {
				registry.RegisterField[FieldHost, *FieldNested](r, "Nested", core.NewFuncCloner(cloneFieldNested))
				registry.RegisterField[AnotherHost, *FieldNested](r, "Nested", core.NewFuncCloner(cloneFieldNested))
			},
			wantLen: 2,
		},
		{
			name: "duplicate_field_registration_does_not_grow",
			actions: func(r *registry.Registry) {
				registry.RegisterField[FieldHost, *FieldNested](r, "Nested", core.NewFuncCloner(cloneFieldNested))
				registry.RegisterField[FieldHost, *FieldNested](r, "Nested", core.NewFuncCloner(cloneFieldNested))
			},
			wantLen: 1,
		},
		{
			name: "register_then_deregister",
			actions: func(r *registry.Registry) {
				registry.RegisterField[FieldHost, *FieldNested](r, "Nested", core.NewFuncCloner(cloneFieldNested))
				registry.RegisterField[FieldHost, []string](r, "Tags", core.NewFuncCloner(
					func(src []string) ([]string, error) { return append([]string{}, src...), nil },
				))
				registry.DeregisterField[FieldHost](r, "Nested")
			},
			wantLen: 1,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reg := registry.New()
			tc.actions(reg)

			if got := reg.FieldLen(); got != tc.wantLen {
				t.Errorf("FieldLen: got %d, want %d", got, tc.wantLen)
			}
		})
	}
}

// --- Error propagation from field cloner --------------------

func TestFieldCloner_ErrorPropagation(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("field clone failure")

	reg := registry.New()
	registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(
		func(src *FieldNested) (*FieldNested, error) {
			return nil, sentinel
		},
	))

	cloner, found := registry.LookupField[FieldHost, *FieldNested](reg, "Nested")
	if !found {
		t.Fatal("field cloner not found")
	}

	_, err := cloner.Clone(&FieldNested{Label: "test"})
	if err == nil {
		t.Fatal("expected error from field cloner, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("errors.Is failed: got %v, want %v", err, sentinel)
	}
}

// --- LookupAnyField (reflect-level bridge) --------------------

func TestLookupAnyField(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "returns_false_for_unregistered_field",
			check: func(t *testing.T) {
				reg := registry.New()
				_, found := reg.LookupAnyField(reflect.TypeOf(FieldHost{}), "Nested")
				if found {
					t.Error("LookupAnyField should return false for unregistered field")
				}
			},
		},
		{
			name: "returns_true_for_registered_field",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))
				_, found := reg.LookupAnyField(reflect.TypeOf(FieldHost{}), "Nested")
				if !found {
					t.Error("LookupAnyField should return true for registered field")
				}
			},
		},
		{
			name: "returned_function_produces_correct_result",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))

				cloneFn, found := reg.LookupAnyField(reflect.TypeOf(FieldHost{}), "Nested")
				if !found {
					t.Fatal("LookupAnyField returned false")
				}

				src := reflect.ValueOf(&FieldNested{Label: "test", Count: 7})
				result, err := cloneFn(src)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				cloned := result.Interface().(*FieldNested)
				if cloned.Label != "test_cloned" {
					t.Errorf("got Label=%q, want %q", cloned.Label, "test_cloned")
				}
				if cloned.Count != 7 {
					t.Errorf("got Count=%d, want 7", cloned.Count)
				}
			},
		},
		{
			name: "different_struct_types_with_same_field_name_are_independent",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))
				registry.RegisterField[AnotherHost, *FieldNested](reg, "Nested", core.NewFuncCloner(
					func(src *FieldNested) (*FieldNested, error) {
						return &FieldNested{Label: "another_" + src.Label, Count: src.Count}, nil
					},
				))

				// FieldHost's "Nested" should use the first cloner
				fn1, found1 := reg.LookupAnyField(reflect.TypeOf(FieldHost{}), "Nested")
				if !found1 {
					t.Fatal("FieldHost.Nested not found")
				}
				result1, _ := fn1(reflect.ValueOf(&FieldNested{Label: "x"}))
				if result1.Interface().(*FieldNested).Label != "x_cloned" {
					t.Error("FieldHost.Nested used wrong cloner")
				}

				// AnotherHost's "Nested" should use the second cloner
				fn2, found2 := reg.LookupAnyField(reflect.TypeOf(AnotherHost{}), "Nested")
				if !found2 {
					t.Fatal("AnotherHost.Nested not found")
				}
				result2, _ := fn2(reflect.ValueOf(&FieldNested{Label: "x"}))
				if result2.Interface().(*FieldNested).Label != "another_x" {
					t.Error("AnotherHost.Nested used wrong cloner")
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.check(t)
		})
	}
}

// --- Thread safety for field cloners --------------------

func TestFieldCloners_ThreadSafety(t *testing.T) {
	t.Parallel()

	reg := registry.New()

	const goroutineCount = 100
	var waitGroup sync.WaitGroup

	for workerIdx := 0; workerIdx < goroutineCount; workerIdx++ {
		waitGroup.Add(1)
		workerIdx := workerIdx

		go func() {
			defer waitGroup.Done()

			switch workerIdx % 6 {
			case 0:
				registry.RegisterField[FieldHost, *FieldNested](reg, "Nested", core.NewFuncCloner(cloneFieldNested))
			case 1:
				registry.LookupField[FieldHost, *FieldNested](reg, "Nested")
			case 2:
				registry.HasField[FieldHost](reg, "Nested")
			case 3:
				registry.DeregisterField[FieldHost](reg, "Nested")
			case 4:
				_ = reg.FieldLen()
			case 5:
				reg.LookupAnyField(reflect.TypeOf(FieldHost{}), "Nested")
			}
		}()
	}

	waitGroup.Wait()
	// No assertions needed: a data race would cause a test failure via -race.
}

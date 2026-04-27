package registry_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/seyallius/doppel/core"
	"github.com/seyallius/doppel/registry"
)

// --- Fixture types — small, focused types that exercise specific lookup paths. --------------------

type (
	TypeA struct{ Value int }
	TypeB struct{ Label string }
	TypeC struct{ Active bool }
)

func cloneTypeA(src TypeA) (TypeA, error) {
	return TypeA{Value: src.Value * 10}, nil
}

func cloneTypeB(src TypeB) (TypeB, error) {
	return TypeB{Label: src.Label + "_cloned"}, nil
}

// -------------------------------------------- Tests --------------------------------------------

func TestNew(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		check func(t *testing.T, reg *registry.Registry)
	}{
		{
			name: "returns_non_nil_registry",
			check: func(t *testing.T, reg *registry.Registry) {
				if reg == nil {
					t.Error("New() returned nil")
				}
			},
		},
		{
			name: "starts_with_zero_cloners",
			check: func(t *testing.T, reg *registry.Registry) {
				if reg.Len() != 0 {
					t.Errorf("new registry Len: got %d, want 0", reg.Len())
				}
			},
		},
		{
			name: "each_call_returns_independent_registry",
			check: func(t *testing.T, _ *registry.Registry) {
				regA := registry.New()
				regB := registry.New()
				registry.Register(regA, core.NewFuncCloner(cloneTypeA))

				if registry.Has[TypeA](regB) {
					t.Error("registering in regA leaked into regB")
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.check(t, registry.New())
		})
	}
}

func TestRegister(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "registered_cloner_is_discoverable_via_has",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.Register(reg, core.NewFuncCloner(cloneTypeA))
				if !registry.Has[TypeA](reg) {
					t.Error("Has[TypeA] returned false after Register")
				}
			},
		},
		{
			name: "len_increments_on_new_type",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.Register(reg, core.NewFuncCloner(cloneTypeA))
				if reg.Len() != 1 {
					t.Errorf("Len after one Register: got %d, want 1", reg.Len())
				}
				registry.Register(reg, core.NewFuncCloner(cloneTypeB))
				if reg.Len() != 2 {
					t.Errorf("Len after two Registers: got %d, want 2", reg.Len())
				}
			},
		},
		{
			name: "registering_same_type_twice_does_not_grow_len",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.Register(reg, core.NewFuncCloner(cloneTypeA))
				registry.Register(reg, core.NewFuncCloner(cloneTypeA))
				if reg.Len() != 1 {
					t.Errorf("Len after double-register: got %d, want 1", reg.Len())
				}
			},
		},
		{
			name: "different_types_stored_independently",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.Register(reg, core.NewFuncCloner(cloneTypeA))
				registry.Register(reg, core.NewFuncCloner(cloneTypeB))

				if !registry.Has[TypeA](reg) {
					t.Error("TypeA not found after registration")
				}
				if !registry.Has[TypeB](reg) {
					t.Error("TypeB not found after registration")
				}
				if registry.Has[TypeC](reg) {
					t.Error("TypeC found but was never registered")
				}
			},
		},
		{
			name: "pointer_and_value_type_are_distinct_keys",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.Register(reg, core.NewFuncCloner(cloneTypeA))                       // TypeA (value)
				registry.Register(reg, core.NewFuncCloner(func(src *TypeA) (*TypeA, error) { // *TypeA (pointer)
					return &TypeA{Value: src.Value}, nil
				}))
				// Both are stored — they have different reflect.Type keys.
				if reg.Len() != 2 {
					t.Errorf("expected 2 entries for TypeA and *TypeA, got %d", reg.Len())
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

func TestLookup(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "lookup_returns_false_for_unregistered_type",
			check: func(t *testing.T) {
				reg := registry.New()
				cloner, found := registry.Lookup[TypeA](reg)
				if found {
					t.Error("Lookup found a cloner for unregistered type")
				}
				if cloner != nil {
					t.Error("Lookup returned non-nil cloner for unregistered type")
				}
			},
		},
		{
			name: "lookup_returns_true_for_registered_type",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.Register(reg, core.NewFuncCloner(cloneTypeA))
				cloner, found := registry.Lookup[TypeA](reg)
				if !found {
					t.Fatal("Lookup returned false for registered type")
				}
				if cloner == nil {
					t.Error("Lookup returned nil cloner for registered type")
				}
			},
		},
		{
			name: "looked_up_cloner_produces_correct_result",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.Register(reg, core.NewFuncCloner(cloneTypeA))

				cloner, found := registry.Lookup[TypeA](reg)
				if !found {
					t.Fatal("cloner not found")
				}

				src := TypeA{Value: 5}
				got, err := cloner.Clone(src)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.Value != 50 { // cloneTypeA multiplies by 10
					t.Errorf("cloner result: got %d, want 50", got.Value)
				}
			},
		},
		{
			name: "lookup_does_not_find_different_type",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.Register(reg, core.NewFuncCloner(cloneTypeA))

				_, found := registry.Lookup[TypeB](reg)
				if found {
					t.Error("Lookup[TypeB] returned true when only TypeA is registered")
				}
			},
		},
		{
			name: "second_register_replaces_first",
			check: func(t *testing.T) {
				reg := registry.New()

				// First registration: multiplies by 10.
				registry.Register(reg, core.NewFuncCloner(cloneTypeA))
				// Second registration: multiplies by 100.
				registry.Register(reg, core.NewFuncCloner(func(src TypeA) (TypeA, error) {
					return TypeA{Value: src.Value * 100}, nil
				}))

				cloner, _ := registry.Lookup[TypeA](reg)
				got, _ := cloner.Clone(TypeA{Value: 3})
				if got.Value != 300 { // second cloner (×100) should have replaced first (×10)
					t.Errorf("expected second cloner to be active (got %d, want 300)", got.Value)
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

func TestDeregister(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "deregister_removes_registered_cloner",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.Register(reg, core.NewFuncCloner(cloneTypeA))
				registry.Deregister[TypeA](reg)

				if registry.Has[TypeA](reg) {
					t.Error("Has[TypeA] returned true after Deregister")
				}
				if reg.Len() != 0 {
					t.Errorf("Len after Deregister: got %d, want 0", reg.Len())
				}
			},
		},
		{
			name: "deregister_unregistered_type_is_noop",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.Register(reg, core.NewFuncCloner(cloneTypeA))

				// Deregister a type that was never registered — must not panic or corrupt state.
				registry.Deregister[TypeB](reg)

				if !registry.Has[TypeA](reg) {
					t.Error("Deregistering TypeB removed TypeA")
				}
				if reg.Len() != 1 {
					t.Errorf("Len after noop Deregister: got %d, want 1", reg.Len())
				}
			},
		},
		{
			name: "deregister_only_removes_target_type",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.Register(reg, core.NewFuncCloner(cloneTypeA))
				registry.Register(reg, core.NewFuncCloner(cloneTypeB))
				registry.Deregister[TypeA](reg)

				if registry.Has[TypeA](reg) {
					t.Error("TypeA still present after Deregister")
				}
				if !registry.Has[TypeB](reg) {
					t.Error("TypeB was removed unexpectedly")
				}
			},
		},
		{
			name: "can_re_register_after_deregister",
			check: func(t *testing.T) {
				reg := registry.New()
				registry.Register(reg, core.NewFuncCloner(cloneTypeA))
				registry.Deregister[TypeA](reg)
				registry.Register(reg, core.NewFuncCloner(func(src TypeA) (TypeA, error) {
					return TypeA{Value: src.Value + 1}, nil
				}))

				cloner, found := registry.Lookup[TypeA](reg)
				if !found {
					t.Fatal("re-registered cloner not found")
				}
				got, _ := cloner.Clone(TypeA{Value: 9})
				if got.Value != 10 {
					t.Errorf("re-registered cloner result: got %d, want 10", got.Value)
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

func TestHas(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		setup     func(*registry.Registry)
		wantTypeA bool
		wantTypeB bool
		wantTypeC bool
	}{
		{
			name:      "empty_registry_has_nothing",
			setup:     func(_ *registry.Registry) {},
			wantTypeA: false,
			wantTypeB: false,
			wantTypeC: false,
		},
		{
			name:      "has_only_registered_types",
			setup:     func(r *registry.Registry) { registry.Register(r, core.NewFuncCloner(cloneTypeA)) },
			wantTypeA: true,
			wantTypeB: false,
			wantTypeC: false,
		},
		{
			name: "has_multiple_registered_types",
			setup: func(r *registry.Registry) {
				registry.Register(r, core.NewFuncCloner(cloneTypeA))
				registry.Register(r, core.NewFuncCloner(cloneTypeB))
			},
			wantTypeA: true,
			wantTypeB: true,
			wantTypeC: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reg := registry.New()
			tc.setup(reg)

			if registry.Has[TypeA](reg) != tc.wantTypeA {
				t.Errorf("Has[TypeA]: got %v, want %v", registry.Has[TypeA](reg), tc.wantTypeA)
			}
			if registry.Has[TypeB](reg) != tc.wantTypeB {
				t.Errorf("Has[TypeB]: got %v, want %v", registry.Has[TypeB](reg), tc.wantTypeB)
			}
			if registry.Has[TypeC](reg) != tc.wantTypeC {
				t.Errorf("Has[TypeC]: got %v, want %v", registry.Has[TypeC](reg), tc.wantTypeC)
			}
		})
	}
}

func TestLen(t *testing.T) {
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
			name:    "one_registration",
			actions: func(r *registry.Registry) { registry.Register(r, core.NewFuncCloner(cloneTypeA)) },
			wantLen: 1,
		},
		{
			name: "two_distinct_registrations",
			actions: func(r *registry.Registry) {
				registry.Register(r, core.NewFuncCloner(cloneTypeA))
				registry.Register(r, core.NewFuncCloner(cloneTypeB))
			},
			wantLen: 2,
		},
		{
			name: "duplicate_registration_does_not_grow",
			actions: func(r *registry.Registry) {
				registry.Register(r, core.NewFuncCloner(cloneTypeA))
				registry.Register(r, core.NewFuncCloner(cloneTypeA)) // same type, second time
			},
			wantLen: 1,
		},
		{
			name: "register_then_deregister",
			actions: func(r *registry.Registry) {
				registry.Register(r, core.NewFuncCloner(cloneTypeA))
				registry.Register(r, core.NewFuncCloner(cloneTypeB))
				registry.Deregister[TypeA](r)
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

			if got := reg.Len(); got != tc.wantLen {
				t.Errorf("Len: got %d, want %d", got, tc.wantLen)
			}
		})
	}
}

func TestRegistry_ErrorFromCloner(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("intentional clone failure")

	testCases := []struct {
		name      string
		cloneFn   func(TypeA) (TypeA, error)
		expectErr bool
		wantErr   error
	}{
		{
			name:      "successful_cloner_returns_no_error",
			cloneFn:   cloneTypeA,
			expectErr: false,
		},
		{
			name: "failing_cloner_error_propagates",
			cloneFn: func(TypeA) (TypeA, error) {
				return TypeA{}, sentinel
			},
			expectErr: true,
			wantErr:   sentinel,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reg := registry.New()
			registry.Register(reg, core.NewFuncCloner(tc.cloneFn))

			cloner, found := registry.Lookup[TypeA](reg)
			if !found {
				t.Fatal("cloner not found after registration")
			}

			_, err := cloner.Clone(TypeA{Value: 1})

			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("errors.Is failed: got %v, want %v", err, tc.wantErr)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestRegistry_ThreadSafety(t *testing.T) {
	t.Parallel()

	// Concurrently Register, Lookup, Has, Deregister, and Len against the
	// same registry. The race detector will surface any unsynchronised access.
	reg := registry.New()

	const goroutineCount = 100
	var waitGroup sync.WaitGroup

	for workerIdx := 0; workerIdx < goroutineCount; workerIdx++ {
		waitGroup.Add(1)
		workerIdx := workerIdx

		go func() {
			defer waitGroup.Done()

			switch workerIdx % 5 {
			case 0:
				registry.Register(reg, core.NewFuncCloner(cloneTypeA))
			case 1:
				registry.Lookup[TypeA](reg)
			case 2:
				registry.Has[TypeA](reg)
			case 3:
				registry.Deregister[TypeA](reg)
			case 4:
				_ = reg.Len()
			}
		}()
	}

	waitGroup.Wait()
	// No assertions needed: a data race would cause a test failure via -race.
}

// -------------------------------------------- Internal Helpers --------------------------------------------

// formatType is used only in error messages; kept to avoid direct reflect import in tests.
func formatType[T any]() string {
	return fmt.Sprintf("%T", *new(T))
}

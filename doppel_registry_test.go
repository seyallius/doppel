// Package doppel_test — Phase 2 integration tests for CloneWithRegistry.
//
// These tests verify the full lookup chain:
//
//	Registered Cloner[T]  →  core.SelfClonable[T]  →  registry.ErrNoCloner
//
// Domain types (User, Order, Address) are defined in doppel_test.go and
// shared across all files in the doppel_test package.
package doppel_test

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/seyallius/doppel"
	"github.com/seyallius/doppel/core"
	"github.com/seyallius/doppel/manual"
	"github.com/seyallius/doppel/registry"
)

// --- Priority 1 — Registered Cloner[T] is used when present --------------------

func TestCloneWithRegistry_UsesRegisteredCloner(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		setupReg func(*registry.Registry)
		src      Address
		wantCity string
	}{
		{
			name: "registered_cloner_is_invoked",
			setupReg: func(r *registry.Registry) {
				registry.Register(r, core.NewFuncCloner(func(src Address) (Address, error) {
					return Address{
						Street: src.Street,
						City:   src.City + "_reg",
						State:  src.State,
						Zip:    src.Zip,
					}, nil
				}))
			},
			src:      Address{Street: "1 Main", City: "Testville", State: "TX", Zip: "00001"},
			wantCity: "Testville_reg",
		},
		{
			name: "registered_cloner_can_transform_value",
			setupReg: func(r *registry.Registry) {
				registry.Register(r, core.NewFuncCloner(func(src Address) (Address, error) {
					return Address{City: "ALWAYS_THIS"}, nil
				}))
			},
			src:      Address{City: "ignored"},
			wantCity: "ALWAYS_THIS",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reg := registry.New()
			tc.setupReg(reg)

			got, err := doppel.CloneWithRegistry(tc.src, reg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.City != tc.wantCity {
				t.Errorf("City: got %q, want %q", got.City, tc.wantCity)
			}
		})
	}
}

// TestCloneWithRegistry_RegisteredClonerTakesPriorityOverSelfClonable verifies
// that a registered Cloner[T] wins even when T also implements SelfClonable[T].
func TestCloneWithRegistry_RegisteredClonerTakesPriorityOverSelfClonable(t *testing.T) {
	t.Parallel()

	// User implements SelfClonable[*User] via its Clone() method.
	// We register a custom Cloner[*User] that produces a sentinel value;
	// if the registry is consulted first, we see the sentinel.
	reg := registry.New()
	registry.Register(reg, core.NewFuncCloner(func(src *User) (*User, error) {
		return &User{Name: "from_registry"}, nil
	}))

	original := newUser() // Name = "Alice"
	cloned, err := doppel.CloneWithRegistry(original, reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cloned.Name != "from_registry" {
		t.Errorf("expected registry cloner to win; got Name=%q, want %q",
			cloned.Name, "from_registry")
	}
}

// --- Priority 2 — SelfClonable fallback when type is not in registry --------------------

func TestCloneWithRegistry_FallsBackToSelfClonable(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		src  *User
	}{
		{
			name: "fully_populated_user",
			src:  newUser(),
		},
		{
			name: "user_with_nil_address",
			src: &User{
				ID:   99,
				Name: "NilAddr",
				Contact: ContactInfo{
					Email:   "nil@example.com",
					Address: nil,
				},
			},
		},
		{
			name: "user_with_empty_collections",
			src: &User{
				ID:     100,
				Name:   "Empty",
				Tags:   []string{},
				Scores: map[string]int{},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			emptyReg := registry.New() // *User not registered

			cloned, err := doppel.CloneWithRegistry(tc.src, emptyReg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(cloned, tc.src) {
				t.Fatalf("SelfClonable fallback result mismatch:\ngot  %+v\nwant %+v", cloned, tc.src)
			}
		})
	}
}

// TestCloneWithRegistry_SelfClonableFallback_Independence verifies that the
// fallback path (SelfClonable) still produces an independent deep copy.
func TestCloneWithRegistry_SelfClonableFallback_Independence(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		mutate func(*User)
		check  func(t *testing.T, cloned *User)
	}{
		{
			name:   "mutate_tags_does_not_affect_clone",
			mutate: func(u *User) { u.Tags[0] = "mutated" },
			check: func(t *testing.T, cloned *User) {
				if cloned.Tags[0] == "mutated" {
					t.Error("SelfClonable fallback clone shares Tags backing array")
				}
			},
		},
		{
			name:   "mutate_nested_address_does_not_affect_clone",
			mutate: func(u *User) { u.Contact.Address.City = "Mutated City" },
			check: func(t *testing.T, cloned *User) {
				if cloned.Contact.Address.City == "Mutated City" {
					t.Error("SelfClonable fallback clone shares Address pointer")
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			original := newUser()
			emptyReg := registry.New()

			cloned, err := doppel.CloneWithRegistry(original, emptyReg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tc.mutate(original)
			tc.check(t, cloned)
		})
	}
}

// --- Priority 3 — ErrNoCloner returned when neither strategy is available --------------------

func TestCloneWithRegistry_ReturnsErrNoClonerWhenNeitherStrategyExists(t *testing.T) {
	t.Parallel()

	// Address does not implement SelfClonable[Address].
	// If it is not in the registry either, ErrNoCloner must be returned.
	testCases := []struct {
		name string
		src  Address
	}{
		{name: "zero_value", src: Address{}},
		{name: "populated", src: Address{Street: "1 St", City: "Nowhere"}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			emptyReg := registry.New()
			_, err := doppel.CloneWithRegistry(tc.src, emptyReg)

			if err == nil {
				t.Fatal("expected ErrNoCloner, got nil")
			}
			if !errors.Is(err, core.ErrNoCloner) {
				t.Errorf("errors.Is(err, ErrNoCloner) failed: got %v", err)
			}
		})
	}
}

// --- Registry override — registering replaces previous behavior --------------------

func TestCloneWithRegistry_ReRegisterOverridesBehaviour(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		firstClone  func(Address) (Address, error)
		secondClone func(Address) (Address, error)
		src         Address
		wantFirst   string
		wantSecond  string
	}{
		{
			name:        "second_registration_replaces_first",
			firstClone:  func(src Address) (Address, error) { return Address{City: "first"}, nil },
			secondClone: func(src Address) (Address, error) { return Address{City: "second"}, nil },
			src:         Address{},
			wantFirst:   "first",
			wantSecond:  "second",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reg := registry.New()

			registry.Register(reg, core.NewFuncCloner(tc.firstClone))
			got1, err := doppel.CloneWithRegistry(tc.src, reg)
			if err != nil {
				t.Fatalf("first clone error: %v", err)
			}
			if got1.City != tc.wantFirst {
				t.Errorf("after first registration: got City=%q, want %q", got1.City, tc.wantFirst)
			}

			registry.Register(reg, core.NewFuncCloner(tc.secondClone))
			got2, err := doppel.CloneWithRegistry(tc.src, reg)
			if err != nil {
				t.Fatalf("second clone error: %v", err)
			}
			if got2.City != tc.wantSecond {
				t.Errorf("after second registration: got City=%q, want %q", got2.City, tc.wantSecond)
			}
		})
	}
}

// --- Deregister — falling through to SelfClonable after removal

func TestCloneWithRegistry_DeregisterFallsThroughToSelfClonable(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	// Register a *User cloner that always returns a sentinel.
	registry.Register(reg, core.NewFuncCloner(func(src *User) (*User, error) {
		return &User{Name: "sentinel"}, nil
	}))

	// Before deregister: sentinel is returned.
	got, err := doppel.CloneWithRegistry(newUser(), reg)
	if err != nil {
		t.Fatalf("unexpected error before deregister: %v", err)
	}
	if got.Name != "sentinel" {
		t.Fatalf("expected sentinel before deregister, got %q", got.Name)
	}

	// After deregister: SelfClonable.Clone() is called instead.
	registry.Deregister[*User](reg)

	original := newUser()
	got2, err := doppel.CloneWithRegistry(original, reg)
	if err != nil {
		t.Fatalf("unexpected error after deregister: %v", err)
	}
	if !reflect.DeepEqual(got2, original) {
		t.Error("SelfClonable result after deregister does not match original")
	}
	if got2.Name == "sentinel" {
		t.Error("registry cloner still invoked after Deregister")
	}
}

// --- Conditional cloning via registered Cloner --------------------

func TestCloneWithRegistry_ConditionalClone(t *testing.T) {
	t.Parallel()

	// Register a Cloner[*User] that deep-copies only active users;
	// inactive users are replaced with a zero-value placeholder.
	reg := registry.New()
	registry.Register(reg, core.NewFuncCloner(func(src *User) (*User, error) {
		if !src.Active {
			return &User{ID: src.ID, Name: "[inactive]"}, nil
		}
		return src.Clone()
	}))

	testCases := []struct {
		name     string
		src      *User
		wantName string
	}{
		{
			name:     "active_user_is_fully_cloned",
			src:      &User{ID: 1, Name: "Alice", Active: true, Tags: []string{"admin"}, Scores: map[string]int{}},
			wantName: "Alice",
		},
		{
			name:     "inactive_user_receives_placeholder",
			src:      &User{ID: 2, Name: "Bob", Active: false},
			wantName: "[inactive]",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cloned, err := doppel.CloneWithRegistry(tc.src, reg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cloned.Name != tc.wantName {
				t.Errorf("Name: got %q, want %q", cloned.Name, tc.wantName)
			}
		})
	}
}

// --- Multiple types registered — each dispatches to its own Cloner --------------------

func TestCloneWithRegistry_MultipleTypesRegisteredIndependently(t *testing.T) {
	t.Parallel()

	reg := registry.New()

	// Two completely different clone functions for two different types.
	registry.Register(reg, core.NewFuncCloner(func(src Address) (Address, error) {
		return Address{City: "addr_" + src.City}, nil
	}))
	registry.Register(reg, core.NewFuncCloner(func(src Score) (Score, error) {
		return Score{Label: src.Label + "_score", Value: src.Value * 2}, nil
	}))

	testCases := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "address_uses_address_cloner",
			check: func(t *testing.T) {
				src := Address{City: "Springfield"}
				got, err := doppel.CloneWithRegistry(src, reg)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.City != "addr_Springfield" {
					t.Errorf("City: got %q, want %q", got.City, "addr_Springfield")
				}
			},
		},
		{
			name: "score_uses_score_cloner",
			check: func(t *testing.T) {
				src := Score{Label: "math", Value: 50}
				got, err := doppel.CloneWithRegistry(src, reg)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.Label != "math_score" || got.Value != 100 {
					t.Errorf("Score: got {%q, %v}, want {%q, %v}", got.Label, got.Value, "math_score", 100.0)
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

// --- Error propagation from registered Cloner --------------------

func TestCloneWithRegistry_ErrorFromRegisteredClonerPropagates(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("cloner failed")

	testCases := []struct {
		name    string
		cloneFn func(Address) (Address, error)
		wantErr error
	}{
		{
			name:    "error_returned_from_registered_cloner",
			cloneFn: func(Address) (Address, error) { return Address{}, sentinel },
			wantErr: sentinel,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reg := registry.New()
			registry.Register(reg, core.NewFuncCloner(tc.cloneFn))

			_, err := doppel.CloneWithRegistry(Address{}, reg)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("errors.Is failed: got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

// --- Nil values --------------------

func TestCloneWithRegistry_NilPointerType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		setupReg func(*registry.Registry)
	}{
		{
			name: "via_registered_cloner",
			setupReg: func(r *registry.Registry) {
				registry.Register(r, core.NewFuncCloner(func(src *User) (*User, error) {
					if src == nil {
						return nil, nil
					}
					return src.Clone()
				}))
			},
		},
		{
			name:     "via_self_clonable_fallback",
			setupReg: func(_ *registry.Registry) {}, // empty reg → falls through to Clone()
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reg := registry.New()
			tc.setupReg(reg)

			var src *User // nil
			cloned, err := doppel.CloneWithRegistry(src, reg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cloned != nil {
				t.Errorf("expected nil clone for nil input, got %+v", cloned)
			}
		})
	}
}

// --- Concurrency — CloneWithRegistry is safe for concurrent use --------------------

func TestCloneWithRegistry_Concurrency(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	registry.Register(reg, core.NewFuncCloner(func(src Address) (Address, error) {
		// A simple non-trivial clone to exercise the Cloner under load.
		opts, err := manual.CloneMap(map[string]string{"city": src.City}, manual.Identity[string])
		if err != nil {
			return Address{}, err
		}
		return Address{Street: src.Street, City: opts["city"], State: src.State, Zip: src.Zip}, nil
	}))

	original := *newAddress()

	const goroutineCount = 100
	errChannel := make(chan error, goroutineCount)
	var waitGroup sync.WaitGroup

	for range goroutineCount {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			cloned, cloneErr := doppel.CloneWithRegistry(original, reg)
			if cloneErr != nil {
				errChannel <- cloneErr
				return
			}
			if !reflect.DeepEqual(cloned, original) {
				errChannel <- errors.New("concurrent clone value mismatch")
			}
		}()
	}

	waitGroup.Wait()
	close(errChannel)

	for err := range errChannel {
		t.Error(err)
	}
}

// --- Composite: registry + manual helpers compose naturally --------------------

func TestCloneWithRegistry_ComposesWithManualHelpers(t *testing.T) {
	t.Parallel()

	// Register a Cloner[Order] that uses manual helpers internally.
	// This demonstrates the intended composition: registry controls dispatch,
	// manual helpers do the actual field-by-field work.
	reg := registry.New()
	registry.Register(reg, core.NewFuncCloner(func(src Order) (Order, error) {
		metadata, err := manual.CloneMap(src.Metadata, manual.Identity[string])
		if err != nil {
			return Order{}, core.WrapError("Order.Metadata", err)
		}
		items, err := manual.CloneSlice(src.Items, func(s Score) (Score, error) {
			return Score{Label: s.Label, Value: s.Value}, nil
		})
		if err != nil {
			return Order{}, core.WrapError("Order.Items", err)
		}
		return Order{
			ID:       src.ID,
			Customer: src.Customer, // shallow for this test — Customer is separately registered
			Items:    items,
			Metadata: metadata,
		}, nil
	}))

	original := Order{
		ID:       "ORD-999",
		Items:    []Score{{Label: "bolt", Value: 1.99}, {Label: "nut", Value: 0.49}},
		Metadata: map[string]string{"channel": "direct"},
	}

	cloned, err := doppel.CloneWithRegistry(original, reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(cloned, original) {
		t.Fatalf("composite clone mismatch:\ngot  %+v\nwant %+v", cloned, original)
	}

	// Verify independence of cloned collections.
	original.Items[0].Value = 999
	if cloned.Items[0].Value == 999 {
		t.Error("cloned Items shares memory with original")
	}
	original.Metadata["channel"] = "mutated"
	if cloned.Metadata["channel"] == "mutated" {
		t.Error("cloned Metadata shares memory with original")
	}
}

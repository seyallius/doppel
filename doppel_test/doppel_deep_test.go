// Package doppel_test — Phase 3 integration tests for CloneDeep.
//
// These tests verify the full priority chain including reflection fallback:
//
//	Registered Cloner[T]  →  SelfClonable[T]  →  Engine (with field cloners)
package doppel_test

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/seyallius/doppel"
	"github.com/seyallius/doppel/core"
	"github.com/seyallius/doppel/registry"
)

// --- Fixture types for CloneDeep tests --------------------

// deepAddress has only primitive fields — no SelfClonable implementation.
type deepAddress struct {
	Street string
	City   string
	State  string
	Zip    string
}

// deepNested is a nested struct with a pointer field.
type deepNested struct {
	Label string
	Count int
}

// deepUser is a realistic aggregate that does NOT implement SelfClonable.
// This is the primary Phase 3 use case: CloneDeep uses reflection + field cloners.
type deepUser struct {
	ID      int64
	Name    string
	Active  bool
	Address *deepAddress
	Tags    []string
	Scores  map[string]int
}

// selfClonableUser implements SelfClonable — used to test priority:
// type cloner > self clonable > engine.
type selfClonableUser struct {
	ID   int64
	Name string
	Tags []string
}

func (u *selfClonableUser) Clone() (*selfClonableUser, error) {
	if u == nil {
		return nil, nil
	}
	tags := make([]string, len(u.Tags))
	copy(tags, u.Tags)
	return &selfClonableUser{ID: u.ID, Name: u.Name + "_self", Tags: tags}, nil
}

// bigModel simulates a struct with many fields where only a few need custom cloning.
type bigModel struct {
	ID       int64
	Name     string
	Active   bool
	Priority int
	Weight   float64
	Metadata map[string]string
	Address  *deepAddress
	Tags     []string
	// ... imagine 190 more primitive fields here
}

// withCloneTagStruct uses doppel:"clone" on the Address field.
type withCloneTagStruct struct {
	Name    string
	Address *deepAddress `doppel:"clone"`
}

// --- Priority 1 — Registered type Cloner[T] wins over everything --------------------

func TestCloneDeep_RegisteredTypeClonerWins(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	registry.Register(reg, core.NewFuncCloner(func(src deepUser) (deepUser, error) {
		return deepUser{Name: "from_type_cloner"}, nil
	}))

	src := deepUser{ID: 1, Name: "Alice"}
	cloned, err := doppel.CloneDeep(src, reg)
	requireNoError(t, err)

	if cloned.Name != "from_type_cloner" {
		t.Errorf("type cloner should win: got %q, want %q", cloned.Name, "from_type_cloner")
	}
}

// --- Priority 2 — SelfClonable fallback when no type cloner --------------------

func TestCloneDeep_SelfClonableFallback(t *testing.T) {
	t.Parallel()

	emptyReg := registry.New()
	src := &selfClonableUser{ID: 1, Name: "Alice", Tags: []string{"admin"}}

	cloned, err := doppel.CloneDeep(src, emptyReg)
	requireNoError(t, err)

	if cloned.Name != "Alice_self" {
		t.Errorf("SelfClonable fallback: got %q, want %q", cloned.Name, "Alice_self")
	}
}

// --- Priority 3 — Engine reflection fallback --------------------

func TestCloneDeep_ReflectionFallback(t *testing.T) {
	t.Parallel()

	// deepUser has no SelfClonable and no type cloner — engine is used.
	emptyReg := registry.New()
	src := deepUser{
		ID:      1,
		Name:    "Alice",
		Active:  true,
		Address: &deepAddress{Street: "123 Main", City: "Springfield"},
		Tags:    []string{"admin", "editor"},
		Scores:  map[string]int{"math": 95},
	}

	cloned, err := doppel.CloneDeep(src, emptyReg)
	requireNoError(t, err)

	if cloned.Name != "Alice" {
		t.Errorf("Name: got %q, want %q", cloned.Name, "Alice")
	}
	if cloned.Address == nil {
		t.Fatal("Address is nil")
	}
	if cloned.Address.City != "Springfield" {
		t.Errorf("Address.City: got %q, want %q", cloned.Address.City, "Springfield")
	}
	if len(cloned.Tags) != 2 {
		t.Errorf("Tags len: got %d, want 2", len(cloned.Tags))
	}
	if cloned.Scores["math"] != 95 {
		t.Errorf("Scores[math]: got %d, want 95", cloned.Scores["math"])
	}
}

func TestCloneDeep_ReflectionFallback_Independence(t *testing.T) {
	t.Parallel()

	emptyReg := registry.New()
	src := deepUser{
		ID:      1,
		Name:    "Alice",
		Address: &deepAddress{Street: "123 Main"},
		Tags:    []string{"admin"},
		Scores:  map[string]int{"math": 95},
	}

	cloned, err := doppel.CloneDeep(src, emptyReg)
	requireNoError(t, err)

	// Mutate original — clone must be independent.
	src.Address.Street = "mutated"
	if cloned.Address.Street == "mutated" {
		t.Error("Address not independent")
	}

	src.Tags[0] = "mutated_tag"
	if cloned.Tags[0] == "mutated_tag" {
		t.Error("Tags not independent")
	}

	src.Scores["math"] = 0
	if cloned.Scores["math"] == 0 {
		t.Error("Scores not independent")
	}
}

// --- CloneDeep with nil registry — pure reflection --------------------

func TestCloneDeep_NilRegistry_UsesPureReflection(t *testing.T) {
	t.Parallel()

	src := deepUser{
		ID:   1,
		Name: "NoRegistry",
		Tags: []string{"a", "b"},
	}

	cloned, err := doppel.CloneDeep(src, nil)
	requireNoError(t, err)

	if cloned.Name != "NoRegistry" {
		t.Errorf("Name: got %q, want %q", cloned.Name, "NoRegistry")
	}
	if len(cloned.Tags) != 2 {
		t.Errorf("Tags len: got %d, want 2", len(cloned.Tags))
	}
}

// --- CloneDeep with field cloners — the core Phase 3 use case --------------------

func TestCloneDeep_FieldCloners_SelectiveOverride(t *testing.T) {
	t.Parallel()

	// This is the primary Phase 3 use case: a struct with many fields
	// where only a few need custom clone logic.
	reg := registry.New()

	// Custom clone for Address — transforms the city name.
	registry.RegisterField[deepUser, *deepAddress](reg, "Address", core.NewFuncCloner(
		func(src *deepAddress) (*deepAddress, error) {
			if src == nil {
				return nil, nil
			}
			return &deepAddress{
				Street: src.Street,
				City:   src.City + "_cloned",
				State:  src.State,
				Zip:    src.Zip,
			}, nil
		},
	))

	// Custom clone for Tags — ensures independent copy with explicit control.
	registry.RegisterField[deepUser, []string](reg, "Tags", core.NewFuncCloner(
		func(src []string) ([]string, error) {
			result := make([]string, len(src))
			for i, s := range src {
				result[i] = s + "_cloned"
			}
			return result, nil
		},
	))

	src := deepUser{
		ID:      1,
		Name:    "Alice",
		Active:  true,
		Address: &deepAddress{Street: "123 Main", City: "Springfield", State: "IL", Zip: "62701"},
		Tags:    []string{"admin", "editor"},
		Scores:  map[string]int{"math": 95, "english": 88},
	}

	cloned, err := doppel.CloneDeep(src, reg)
	requireNoError(t, err)

	// Primitives: default reflection.
	if cloned.ID != src.ID {
		t.Errorf("ID: got %d, want %d", cloned.ID, src.ID)
	}
	if cloned.Name != src.Name {
		t.Errorf("Name: got %q, want %q", cloned.Name, src.Name)
	}
	if cloned.Active != src.Active {
		t.Errorf("Active: got %v, want %v", cloned.Active, src.Active)
	}

	// Address: field cloner applied.
	if cloned.Address == nil {
		t.Fatal("Address is nil")
	}
	if cloned.Address.City != "Springfield_cloned" {
		t.Errorf("Address.City: got %q, want %q", cloned.Address.City, "Springfield_cloned")
	}
	if cloned.Address.Street != "123 Main" {
		t.Errorf("Address.Street: got %q, want %q", cloned.Address.Street, "123 Main")
	}

	// Tags: field cloner applied.
	if len(cloned.Tags) != 2 {
		t.Fatalf("Tags len: got %d, want 2", len(cloned.Tags))
	}
	if cloned.Tags[0] != "admin_cloned" {
		t.Errorf("Tags[0]: got %q, want %q", cloned.Tags[0], "admin_cloned")
	}
	if cloned.Tags[1] != "editor_cloned" {
		t.Errorf("Tags[1]: got %q, want %q", cloned.Tags[1], "editor_cloned")
	}

	// Scores: default reflection.
	if cloned.Scores["math"] != 95 {
		t.Errorf("Scores[math]: got %d, want 95", cloned.Scores["math"])
	}
}

func TestCloneDeep_BigModel_SelectiveOverride(t *testing.T) {
	t.Parallel()

	// Simulates the "200-field struct, only 2 need custom cloning" scenario.
	reg := registry.New()

	registry.RegisterField[bigModel, *deepAddress](reg, "Address", core.NewFuncCloner(
		func(src *deepAddress) (*deepAddress, error) {
			if src == nil {
				return nil, nil
			}
			return &deepAddress{
				Street: src.Street,
				City:   src.City + "_deep",
				State:  src.State,
				Zip:    src.Zip,
			}, nil
		},
	))

	registry.RegisterField[bigModel, []string](reg, "Tags", core.NewFuncCloner(
		func(src []string) ([]string, error) {
			return append([]string{}, src...), nil
		},
	))

	src := bigModel{
		ID:       42,
		Name:     "BigModel",
		Active:   true,
		Priority: 5,
		Weight:   3.14,
		Metadata: map[string]string{"env": "prod"},
		Address:  &deepAddress{Street: "1 Main", City: "Metro"},
		Tags:     []string{"critical", "monitored"},
	}

	cloned, err := doppel.CloneDeep(src, reg)
	requireNoError(t, err)

	// All primitive fields cloned by default reflection.
	if cloned.ID != 42 {
		t.Errorf("ID: got %d, want 42", cloned.ID)
	}
	if cloned.Name != "BigModel" {
		t.Errorf("Name: got %q, want %q", cloned.Name, "BigModel")
	}
	if !cloned.Active {
		t.Error("Active: got false, want true")
	}
	if cloned.Priority != 5 {
		t.Errorf("Priority: got %d, want 5", cloned.Priority)
	}
	if cloned.Weight != 3.14 {
		t.Errorf("Weight: got %f, want 3.14", cloned.Weight)
	}

	// Address: field cloner applied.
	if cloned.Address.City != "Metro_deep" {
		t.Errorf("Address.City: got %q, want %q", cloned.Address.City, "Metro_deep")
	}

	// Tags: field cloner applied (independent copy).
	if len(cloned.Tags) != 2 {
		t.Errorf("Tags len: got %d, want 2", len(cloned.Tags))
	}
	src.Tags[0] = "mutated"
	if cloned.Tags[0] == "mutated" {
		t.Error("Tags should be independent from original")
	}

	// Metadata: default reflection (independent).
	if cloned.Metadata["env"] != "prod" {
		t.Errorf("Metadata[env]: got %q, want %q", cloned.Metadata["env"], "prod")
	}
	src.Metadata["env"] = "mutated"
	if cloned.Metadata["env"] == "mutated" {
		t.Error("Metadata should be independent from original")
	}
}

// --- Type cloner > Field cloner > Engine priority --------------------

func TestCloneDeep_TypeClonerWinsOverFieldCloner(t *testing.T) {
	t.Parallel()

	reg := registry.New()

	// Type cloner for deepUser — should win over everything.
	registry.Register(reg, core.NewFuncCloner(func(src deepUser) (deepUser, error) {
		return deepUser{Name: "type_cloner_wins"}, nil
	}))

	// Field cloner for deepUser.Address — should NOT be reached.
	registry.RegisterField[deepUser, *deepAddress](reg, "Address", core.NewFuncCloner(
		func(src *deepAddress) (*deepAddress, error) {
			return &deepAddress{City: "field_cloner_used"}, nil
		},
	))

	src := deepUser{ID: 1, Name: "Alice", Address: &deepAddress{City: "Original"}}
	cloned, err := doppel.CloneDeep(src, reg)
	requireNoError(t, err)

	if cloned.Name != "type_cloner_wins" {
		t.Errorf("type cloner should win: got %q", cloned.Name)
	}
	// Address should not be "field_cloner_used" because type cloner handled the whole struct.
}

func TestCloneDeep_SelfClonableWinsOverFieldCloner(t *testing.T) {
	t.Parallel()

	reg := registry.New()

	// Field cloner for selfClonableUser.Tags — should NOT be reached
	// because SelfClonable takes priority.
	registry.RegisterField[selfClonableUser, []string](reg, "Tags", core.NewFuncCloner(
		func(src []string) ([]string, error) {
			return []string{"field_cloner_used"}, nil
		},
	))

	src := &selfClonableUser{ID: 1, Name: "Alice", Tags: []string{"admin"}}
	cloned, err := doppel.CloneDeep(src, reg)
	requireNoError(t, err)

	// SelfClonable should have been used.
	if cloned.Name != "Alice_self" {
		t.Errorf("SelfClonable should win: got %q, want %q", cloned.Name, "Alice_self")
	}
	// Tags should NOT be ["field_cloner_used"].
	if len(cloned.Tags) == 1 && cloned.Tags[0] == "field_cloner_used" {
		t.Error("Field cloner should not be used when SelfClonable handles the type")
	}
}

// --- CloneDeep with doppel:"clone" tag --------------------

func TestCloneDeep_CloneTag(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	registry.RegisterField[withCloneTagStruct, *deepAddress](reg, "Address", core.NewFuncCloner(
		func(src *deepAddress) (*deepAddress, error) {
			return &deepAddress{City: src.City + "_tagged"}, nil
		},
	))

	src := withCloneTagStruct{
		Name:    "tagged",
		Address: &deepAddress{City: "Metro"},
	}

	cloned, err := doppel.CloneDeep(src, reg)
	requireNoError(t, err)

	if cloned.Address.City != "Metro_tagged" {
		t.Errorf("Address.City: got %q, want %q", cloned.Address.City, "Metro_tagged")
	}
}

func TestCloneDeep_CloneTag_WithoutFieldCloner(t *testing.T) {
	t.Parallel()

	emptyReg := registry.New()
	src := withCloneTagStruct{
		Name:    "no_cloner",
		Address: &deepAddress{City: "Metro"},
	}

	_, err := doppel.CloneDeep(src, emptyReg)
	if err == nil {
		t.Fatal("expected error for doppel:\"clone\" without field cloner, got nil")
	}
}

// --- MustCloneDeep --------------------

func TestMustCloneDeep(t *testing.T) {
	t.Parallel()

	t.Run("returns_correct_clone", func(t *testing.T) {
		t.Parallel()
		src := deepUser{ID: 1, Name: "must_clone"}
		cloned := doppel.MustCloneDeep(src, nil)
		if cloned.Name != "must_clone" {
			t.Errorf("got %q, want %q", cloned.Name, "must_clone")
		}
	})

	t.Run("panics_on_error", func(t *testing.T) {
		t.Parallel()
		defer func() {
			rec := recover()
			if rec == nil {
				t.Error("expected panic from MustCloneDeep on error, got none")
			}
		}()

		reg := registry.New()
		registry.RegisterField[deepUser, *deepAddress](reg, "Address", core.NewFuncCloner(
			func(src *deepAddress) (*deepAddress, error) {
				return nil, errors.New("deliberate failure")
			},
		))

		src := deepUser{Address: &deepAddress{City: "fail"}}
		_ = doppel.MustCloneDeep(src, reg)
	})
}

// --- Nil values --------------------

func TestCloneDeep_NilPointer(t *testing.T) {
	t.Parallel()

	var src *deepUser // nil
	cloned, err := doppel.CloneDeep(src, nil)
	requireNoError(t, err)
	if cloned != nil {
		t.Errorf("expected nil clone for nil input, got %+v", cloned)
	}
}

func TestCloneDeep_NilSlice(t *testing.T) {
	t.Parallel()

	src := deepUser{ID: 1, Tags: nil}
	cloned, err := doppel.CloneDeep(src, nil)
	requireNoError(t, err)
	if cloned.Tags != nil {
		t.Error("nil slice should be preserved as nil")
	}
}

func TestCloneDeep_EmptySlice(t *testing.T) {
	t.Parallel()

	src := deepUser{ID: 1, Tags: []string{}}
	cloned, err := doppel.CloneDeep(src, nil)
	requireNoError(t, err)
	if cloned.Tags == nil {
		t.Error("empty slice should be preserved as non-nil empty slice")
	}
	if len(cloned.Tags) != 0 {
		t.Errorf("empty slice length: got %d, want 0", len(cloned.Tags))
	}
}

// --- Error propagation --------------------

func TestCloneDeep_ErrorFromFieldCloner(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("field failure")
	reg := registry.New()
	registry.RegisterField[deepUser, *deepAddress](reg, "Address", core.NewFuncCloner(
		func(src *deepAddress) (*deepAddress, error) {
			return nil, sentinel
		},
	))

	src := deepUser{Address: &deepAddress{City: "fail"}}
	_, err := doppel.CloneDeep(src, reg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("errors.Is failed: got %v", err)
	}
}

// --- Concurrency --------------------

func TestCloneDeep_Concurrency(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	registry.RegisterField[deepUser, *deepAddress](reg, "Address", core.NewFuncCloner(
		func(src *deepAddress) (*deepAddress, error) {
			return &deepAddress{Street: src.Street, City: src.City, State: src.State, Zip: src.Zip}, nil
		},
	))

	src := deepUser{
		ID:      1,
		Name:    "concurrent",
		Address: &deepAddress{City: "Metro"},
		Tags:    []string{"a", "b"},
	}

	const goroutineCount = 50
	errCh := make(chan error, goroutineCount)
	var wg sync.WaitGroup

	for range goroutineCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cloned, cloneErr := doppel.CloneDeep(src, reg)
			if cloneErr != nil {
				errCh <- cloneErr
				return
			}
			if cloned.Name != "concurrent" {
				errCh <- errors.New("concurrent clone value mismatch")
			}
		}()
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Error(err)
	}
}

// --- Comparison: CloneWithRegistry vs CloneDeep --------------------

func TestCloneDeep_ReturnsResultWhenCloneWithRegistryWouldError(t *testing.T) {
	t.Parallel()

	emptyReg := registry.New()
	src := deepUser{ID: 1, Name: "Comparison"}

	// CloneWithRegistry would return ErrNoCloner because deepUser
	// has no type cloner and doesn't implement SelfClonable.
	_, err := doppel.CloneWithRegistry(src, emptyReg)
	if err == nil {
		t.Error("CloneWithRegistry should return error for non-SelfClonable type without registry")
	}
	if !errors.Is(err, core.ErrNoCloner) {
		t.Errorf("expected ErrNoCloner, got %v", err)
	}

	// CloneDeep should succeed using the reflection engine.
	cloned, err := doppel.CloneDeep(src, emptyReg)
	requireNoError(t, err)
	if cloned.Name != "Comparison" {
		t.Errorf("CloneDeep: got %q, want %q", cloned.Name, "Comparison")
	}
}

// --- Deep equality check --------------------

func TestCloneDeep_DeepEquality(t *testing.T) {
	t.Parallel()

	src := deepUser{
		ID:      1,
		Name:    "DeepEqual",
		Active:  true,
		Address: &deepAddress{Street: "1 St", City: "Town", State: "ST", Zip: "00000"},
		Tags:    []string{"a", "b", "c"},
		Scores:  map[string]int{"x": 1, "y": 2, "z": 3},
	}

	cloned, err := doppel.CloneDeep(src, nil)
	requireNoError(t, err)

	if !reflect.DeepEqual(cloned, src) {
		t.Fatalf("clone not equal to original:\ngot  %+v\nwant %+v", cloned, src)
	}
}

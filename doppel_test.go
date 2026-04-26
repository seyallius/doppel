// Package doppel_test. doppel_test - Exercises the full manual cloning stack
// via realistic domain types that demonstrate how CloneSlice, CloneMap, and ClonePointer
// compose inside a struct's own Clone() method.
package doppel_test

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/seyallius/doppel"
	"github.com/seyallius/doppel/core"
	"github.com/seyallius/doppel/manual"
)

// ---------------------------------------------------------------------------
// Domain types & their manual Clone() implementations
// ---------------------------------------------------------------------------

// Address holds a physical location. All fields are primitives, so its
// clone is a plain struct literal — no helper functions needed.
type Address struct {
	Street string
	City   string
	State  string
	Zip    string
}

// cloneAddress is a stand-alone clone function for Address.
// It is intentionally not a method so we can pass it as a func(Address)(Address,error)
// to ClonePointer.
func cloneAddress(src Address) (Address, error) {
	return Address{
		Street: src.Street,
		City:   src.City,
		State:  src.State,
		Zip:    src.Zip,
	}, nil
}

// ContactInfo embeds an Address via pointer, illustrating pointer deep copy.
type ContactInfo struct {
	Email   string
	Phone   string
	Address *Address
}

// cloneContactInfo is the clone function for ContactInfo.
func cloneContactInfo(src ContactInfo) (ContactInfo, error) {
	clonedAddress, err := manual.ClonePointer(src.Address, cloneAddress)
	if err != nil {
		return ContactInfo{}, core.WrapError("ContactInfo.Address", err)
	}
	return ContactInfo{
		Email:   src.Email,
		Phone:   src.Phone,
		Address: clonedAddress,
	}, nil
}

// Score represents a labelled numeric score.
type Score struct {
	Label string
	Value float64
}

// cloneScore clones a Score (struct with only value fields).
func cloneScore(src Score) (Score, error) {
	return Score{Label: src.Label, Value: src.Value}, nil
}

// User is a realistic aggregate with nested structs, a pointer field,
// string slices, and numeric maps.
type User struct {
	ID      int64
	Name    string
	Active  bool
	Contact ContactInfo
	Tags    []string
	Scores  map[string]int
	Aliases []string
}

// Clone implements core.SelfClonable[*User].
// It composes the manual helpers to produce a complete deep copy.
func (u *User) Clone() (*User, error) {
	if u == nil {
		return nil, nil
	}

	contact, err := cloneContactInfo(u.Contact)
	if err != nil {
		return nil, core.WrapError("User.Contact", err)
	}

	tags, err := manual.CloneSlice(u.Tags, manual.Identity[string])
	if err != nil {
		return nil, core.WrapError("User.Tags", err)
	}

	scores, err := manual.CloneMap(u.Scores, manual.Identity[int])
	if err != nil {
		return nil, core.WrapError("User.Scores", err)
	}

	aliases, err := manual.CloneSlice(u.Aliases, manual.Identity[string])
	if err != nil {
		return nil, core.WrapError("User.Aliases", err)
	}

	return &User{
		ID:      u.ID,
		Name:    u.Name,
		Active:  u.Active,
		Contact: contact,
		Tags:    tags,
		Scores:  scores,
		Aliases: aliases,
	}, nil
}

// Order contains a slice of Score structs, demonstrating CloneSlice with a
// non-trivial element cloner.
type Order struct {
	ID       string
	Customer *User
	Items    []Score
	Metadata map[string]string
}

// Clone implements core.SelfClonable[*Order].
func (o *Order) Clone() (*Order, error) {
	if o == nil {
		return nil, nil
	}

	customer, err := manual.ClonePointer(o.Customer, func(u User) (User, error) {
		cloned, err := u.Clone() // *User.Clone() — pointer receiver
		if err != nil || cloned == nil {
			return User{}, err
		}
		return *cloned, nil
	})
	if err != nil {
		return nil, core.WrapError("Order.Customer", err)
	}

	items, err := manual.CloneSlice(o.Items, cloneScore)
	if err != nil {
		return nil, core.WrapError("Order.Items", err)
	}

	metadata, err := manual.CloneMap(o.Metadata, manual.Identity[string])
	if err != nil {
		return nil, core.WrapError("Order.Metadata", err)
	}

	return &Order{
		ID:       o.ID,
		Customer: customer,
		Items:    items,
		Metadata: metadata,
	}, nil
}

// ---------------------------------------------------------------------------
// Factories
// ---------------------------------------------------------------------------

func newAddress() *Address {
	return &Address{Street: "123 Main St", City: "Springfield", State: "IL", Zip: "62701"}
}

func newUser() *User {
	return &User{
		ID:     1,
		Name:   "Alice",
		Active: true,
		Contact: ContactInfo{
			Email:   "alice@example.com",
			Phone:   "555-0100",
			Address: newAddress(),
		},
		Tags:    []string{"admin", "editor", "viewer"},
		Scores:  map[string]int{"math": 95, "english": 88, "science": 91},
		Aliases: []string{"al", "ally"},
	}
}

func newOrder(customer *User) *Order {
	return &Order{
		ID:       "ORD-001",
		Customer: customer,
		Items:    []Score{{Label: "widget", Value: 9.99}, {Label: "gadget", Value: 24.99}},
		Metadata: map[string]string{"source": "web", "promo": "SAVE10"},
	}
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestClone_User — nested struct with pointer, slice, and map fields
// ---------------------------------------------------------------------------

func TestClone_User(t *testing.T) {
	t.Parallel()

	original := newUser()
	cloned, err := doppel.Clone(original)
	requireNoError(t, err)

	if !reflect.DeepEqual(cloned, original) {
		t.Fatalf("clone not equal to original:\ngot  %+v\nwant %+v", cloned, original)
	}
}

func TestClone_User_Pointer_Independence(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		mutate func(u *User)
		check  func(t *testing.T, cloned *User, original *User)
	}{
		{
			name:   "mutate_name_does_not_affect_clone",
			mutate: func(u *User) { u.Name = "mutated" },
			check: func(t *testing.T, cloned *User, _ *User) {
				if cloned.Name == "mutated" {
					t.Error("clone Name was affected by original mutation")
				}
			},
		},
		{
			name:   "mutate_nested_address_does_not_affect_clone",
			mutate: func(u *User) { u.Contact.Address.Street = "999 Evil Ave" },
			check: func(t *testing.T, cloned *User, _ *User) {
				if cloned.Contact.Address.Street == "999 Evil Ave" {
					t.Error("clone Address.Street was affected by original mutation")
				}
			},
		},
		{
			name:   "mutate_tag_slice_does_not_affect_clone",
			mutate: func(u *User) { u.Tags[0] = "mutated_tag" },
			check: func(t *testing.T, cloned *User, _ *User) {
				if cloned.Tags[0] == "mutated_tag" {
					t.Error("clone Tags[0] was affected by original mutation")
				}
			},
		},
		{
			name:   "mutate_scores_map_does_not_affect_clone",
			mutate: func(u *User) { u.Scores["math"] = 0 },
			check: func(t *testing.T, cloned *User, _ *User) {
				if cloned.Scores["math"] == 0 {
					t.Error("clone Scores[math] was affected by original mutation")
				}
			},
		},
		{
			name:   "replace_address_pointer_does_not_affect_clone",
			mutate: func(u *User) { u.Contact.Address = &Address{Street: "replaced"} },
			check: func(t *testing.T, cloned *User, _ *User) {
				if cloned.Contact.Address.Street == "replaced" {
					t.Error("clone Address was affected by pointer replacement in original")
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			original := newUser()
			cloned, err := doppel.Clone(original)
			requireNoError(t, err)

			tc.mutate(original)
			tc.check(t, cloned, original)
		})
	}
}

// ---------------------------------------------------------------------------
// TestClone_User_NilFields
// ---------------------------------------------------------------------------

func TestClone_User_NilFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		original *User
	}{
		{
			name: "nil_address_pointer",
			original: &User{
				ID:   2,
				Name: "Bob",
				Contact: ContactInfo{
					Email:   "bob@example.com",
					Address: nil, // explicit nil
				},
			},
		},
		{
			name: "nil_tags_slice",
			original: &User{
				ID:   3,
				Name: "Carol",
				Tags: nil, // nil slice preserved
			},
		},
		{
			name: "nil_scores_map",
			original: &User{
				ID:     4,
				Name:   "Dave",
				Scores: nil, // nil map preserved
			},
		},
		{
			name: "empty_tags_and_scores",
			original: &User{
				ID:     5,
				Name:   "Eve",
				Tags:   []string{},
				Scores: map[string]int{},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cloned, err := doppel.Clone(tc.original)
			requireNoError(t, err)

			if !reflect.DeepEqual(cloned, tc.original) {
				t.Fatalf("clone mismatch:\ngot  %+v\nwant %+v", cloned, tc.original)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestClone_Order — slice of structs + nested pointer
// ---------------------------------------------------------------------------

func TestClone_Order(t *testing.T) {
	t.Parallel()

	original := newOrder(newUser())
	cloned, err := doppel.Clone(original)
	requireNoError(t, err)

	if !reflect.DeepEqual(cloned, original) {
		t.Fatalf("Order clone mismatch")
	}
}

func TestClone_Order_Independence(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		mutate func(o *Order)
		check  func(t *testing.T, cloned *Order)
	}{
		{
			name:   "mutate_order_id",
			mutate: func(o *Order) { o.ID = "MUTATED" },
			check: func(t *testing.T, cloned *Order) {
				if cloned.ID == "MUTATED" {
					t.Error("clone OrderID was mutated")
				}
			},
		},
		{
			name:   "mutate_item_value",
			mutate: func(o *Order) { o.Items[0].Value = 0.01 },
			check: func(t *testing.T, cloned *Order) {
				if cloned.Items[0].Value == 0.01 {
					t.Error("clone Items[0].Value was mutated")
				}
			},
		},
		{
			name:   "mutate_customer_name",
			mutate: func(o *Order) { o.Customer.Name = "MutatedCustomer" },
			check: func(t *testing.T, cloned *Order) {
				if cloned.Customer.Name == "MutatedCustomer" {
					t.Error("clone Customer.Name was mutated")
				}
			},
		},
		{
			name:   "mutate_metadata",
			mutate: func(o *Order) { o.Metadata["source"] = "hacked" },
			check: func(t *testing.T, cloned *Order) {
				if cloned.Metadata["source"] == "hacked" {
					t.Error("clone Metadata was mutated")
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			original := newOrder(newUser())
			cloned, err := doppel.Clone(original)
			requireNoError(t, err)

			tc.mutate(original)
			tc.check(t, cloned)
		})
	}
}

// ---------------------------------------------------------------------------
// TestCloneWith — external Cloner[T] (non-SelfClonable type)
// ---------------------------------------------------------------------------

func TestCloneWith(t *testing.T) {
	t.Parallel()

	// Address does not implement SelfClonable; use CloneWith + FuncCloner.
	addressCloner := core.NewFuncCloner(cloneAddress)

	testCases := []struct {
		name     string
		original Address
	}{
		{
			name:     "zero_value",
			original: Address{},
		},
		{
			name:     "fully_populated",
			original: Address{Street: "1 Infinite Loop", City: "Cupertino", State: "CA", Zip: "95014"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cloned, err := doppel.CloneWith(tc.original, addressCloner)
			requireNoError(t, err)

			if cloned != tc.original {
				t.Fatalf("clone mismatch:\ngot  %+v\nwant %+v", cloned, tc.original)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestMustClone
// ---------------------------------------------------------------------------

func TestMustClone(t *testing.T) {
	t.Parallel()

	t.Run("returns_correct_clone", func(t *testing.T) {
		t.Parallel()
		original := newUser()
		cloned := doppel.MustClone(original)
		if !reflect.DeepEqual(cloned, original) {
			t.Fatal("MustClone result not equal to original")
		}
	})

	t.Run("panics_on_error", func(t *testing.T) {
		t.Parallel()

		failingUser := &failingClonable{}
		defer func() {
			rec := recover()
			if rec == nil {
				t.Error("expected panic from MustClone on error, got none")
			}
		}()
		_ = doppel.MustClone(failingUser)
	})
}

// failingClonable is a test-only type whose Clone() always returns an error.
type failingClonable struct{}

func (f *failingClonable) Clone() (*failingClonable, error) {
	return nil, errors.New("intentional failure")
}

// ---------------------------------------------------------------------------
// TestMustCloneWith
// ---------------------------------------------------------------------------

func TestMustCloneWith(t *testing.T) {
	t.Parallel()

	t.Run("returns_correct_clone", func(t *testing.T) {
		t.Parallel()
		cloner := core.NewFuncCloner(cloneAddress)
		addr := Address{Street: "42 Answer St", City: "Meaning", State: "OF", Zip: "LIFE"}
		cloned := doppel.MustCloneWith(addr, cloner)
		if cloned != addr {
			t.Fatalf("MustCloneWith result mismatch")
		}
	})

	t.Run("panics_on_error", func(t *testing.T) {
		t.Parallel()

		errCloner := core.NewFuncCloner(func(a Address) (Address, error) {
			return Address{}, errors.New("deliberate")
		})
		defer func() {
			if recover() == nil {
				t.Error("expected panic, got none")
			}
		}()
		_ = doppel.MustCloneWith(Address{}, errCloner)
	})
}

// ---------------------------------------------------------------------------
// TestFuncCloner
// ---------------------------------------------------------------------------

func TestFuncCloner(t *testing.T) {
	t.Parallel()

	invocations := 0
	cloner := core.NewFuncCloner(func(v int) (int, error) {
		invocations++
		return v * 10, nil
	})

	got, err := cloner.Clone(7)
	requireNoError(t, err)

	if got != 70 {
		t.Errorf("FuncCloner result: got %d, want 70", got)
	}
	if invocations != 1 {
		t.Errorf("FuncCloner invocation count: got %d, want 1", invocations)
	}
}

// ---------------------------------------------------------------------------
// TestConcurrency — Clone must be safe for concurrent use
// ---------------------------------------------------------------------------

func TestConcurrency(t *testing.T) {
	t.Parallel()

	original := newOrder(newUser())

	const goroutineCount = 50
	errChannel := make(chan error, goroutineCount)
	var waitGroup sync.WaitGroup

	for range goroutineCount {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			cloned, cloneErr := doppel.Clone(original)
			if cloneErr != nil {
				errChannel <- cloneErr
				return
			}
			if !reflect.DeepEqual(cloned, original) {
				errChannel <- errors.New("concurrent clone produced wrong value")
			}
		}()
	}

	waitGroup.Wait()
	close(errChannel)

	for err := range errChannel {
		t.Error(err)
	}
}

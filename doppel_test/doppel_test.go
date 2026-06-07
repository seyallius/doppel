package doppel_test

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/seyallius/doppel"
	"github.com/seyallius/doppel/core"
	"github.com/stretchr/testify/require"
)

// --- TestClone_User — nested struct with pointer, slice, and map fields --------------------

func TestClone_User(t *testing.T) {
	t.Parallel()

	original := newUser()
	cloned, err := doppel.Clone(original)
	require.NoError(t, err)

	if !reflect.DeepEqual(cloned, original) {
		t.Fatalf("clone not equal to original:\ngot  %+v\nwant %+v", cloned, original)
	}
}

func TestClone_User_Pointer_Independence(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		mutate func(u *User)
		check  func(t *testing.T, cloned *User)
	}{
		{
			name:   "mutate_name_does_not_affect_clone",
			mutate: func(u *User) { u.Name = "mutated" },
			check: func(t *testing.T, cloned *User) {
				if cloned.Name == "mutated" {
					t.Error("clone Name was affected by original mutation")
				}
			},
		},
		{
			name:   "mutate_nested_address_does_not_affect_clone",
			mutate: func(u *User) { u.Contact.Address.Street = "999 Evil Ave" },
			check: func(t *testing.T, cloned *User) {
				if cloned.Contact.Address.Street == "999 Evil Ave" {
					t.Error("clone Address.Street was affected by original mutation")
				}
			},
		},
		{
			name:   "mutate_tag_slice_does_not_affect_clone",
			mutate: func(u *User) { u.Tags[0] = "mutated_tag" },
			check: func(t *testing.T, cloned *User) {
				if cloned.Tags[0] == "mutated_tag" {
					t.Error("clone Tags[0] was affected by original mutation")
				}
			},
		},
		{
			name:   "mutate_scores_map_does_not_affect_clone",
			mutate: func(u *User) { u.Scores["math"] = 0 },
			check: func(t *testing.T, cloned *User) {
				if cloned.Scores["math"] == 0 {
					t.Error("clone Scores[math] was affected by original mutation")
				}
			},
		},
		{
			name:   "replace_address_pointer_does_not_affect_clone",
			mutate: func(u *User) { u.Contact.Address = &Address{Street: "replaced"} },
			check: func(t *testing.T, cloned *User) {
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
			require.NoError(t, err)

			tc.mutate(original)
			tc.check(t, cloned)
		})
	}
}

// --- TestClone_User_NilFields --------------------

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
					Address: nil,
				},
			},
		},
		{
			name: "nil_tags_slice",
			original: &User{
				ID:   3,
				Name: "Carol",
				Tags: nil,
			},
		},
		{
			name: "nil_scores_map",
			original: &User{
				ID:     4,
				Name:   "Dave",
				Scores: nil,
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
			require.NoError(t, err)

			if !reflect.DeepEqual(cloned, tc.original) {
				t.Fatalf("clone mismatch:\ngot  %+v\nwant %+v", cloned, tc.original)
			}
		})
	}
}

// --- TestClone_Order — slice of structs + nested pointer --------------------

func TestClone_Order(t *testing.T) {
	t.Parallel()

	original := newOrder(newUser())
	cloned, err := doppel.Clone(original)
	require.NoError(t, err)

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
			require.NoError(t, err)

			tc.mutate(original)
			tc.check(t, cloned)
		})
	}
}

// --- TestMustClone --------------------

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

// --- TestFuncCloner --------------------

func TestFuncCloner(t *testing.T) {
	t.Parallel()

	invocations := 0
	cloner := core.NewFuncCloner(func(v int) (int, error) {
		invocations++
		return v * 10, nil
	})

	got, err := cloner.Clone(7)
	require.NoError(t, err)

	if got != 70 {
		t.Errorf("FuncCloner result: got %d, want 70", got)
	}
	if invocations != 1 {
		t.Errorf("FuncCloner invocation count: got %d, want 1", invocations)
	}
}

// --- TestConcurrency — Clone must be safe for concurrent use --------------------

func TestConcurrency(t *testing.T) {
	t.Parallel()

	original := newOrder(newUser())

	const goroutineCount = 50
	errChannel := make(chan error, goroutineCount)
	var waitGroup sync.WaitGroup

	for i := 0; i < goroutineCount; i++ {
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

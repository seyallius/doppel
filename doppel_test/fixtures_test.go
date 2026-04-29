package doppel_test

import (
	"errors"

	"github.com/seyallius/doppel/core"
	"github.com/seyallius/doppel/manual"
)

// -------------------------------------------- Internal Helpers --------------------------------------------

// Address holds a physical location. All fields are primitives, so its
// clone is a plain struct literal — no helper functions needed.
type Address struct {
	Street string
	City   string
	State  string
	Zip    string
}

// ContactInfo embeds an Address via pointer, illustrating pointer deep copy.
type ContactInfo struct {
	Email   string
	Phone   string
	Address *Address
}

// Score represents a labelled numeric score.
type Score struct {
	Label string
	Value float64
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

// Order contains a slice of Score structs, demonstrating CloneSlice with a
// non-trivial element cloner.
type Order struct {
	ID       string
	Customer *User
	Items    []Score
	Metadata map[string]string
}

// failingClonable is a test-only type whose Clone() always returns an error.
type failingClonable struct{}

// -------------------------------------------- Constructor(s) --------------------------------------------

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

// -------------------------------------------- Internal Helpers --------------------------------------------

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

func (f *failingClonable) Clone() (*failingClonable, error) {
	return nil, errors.New("intentional failure")
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

// cloneScore clones a Score (struct with only value fields).
func cloneScore(src Score) (Score, error) {
	return Score{Label: src.Label, Value: src.Value}, nil
}

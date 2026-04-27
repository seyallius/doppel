// Package registry. registry - implements a thread-safe, type-keyed store of Cloner[T] values.
//
// Reflection is used in this package exclusively for type identification — as the
// key in the internal map. It is never used to read field values, traverse struct
// layouts, or perform any kind of dynamic copying. All actual cloning is delegated
// to the core.Cloner[T] the caller registered.
//
// Lookup priority inside doppel.CloneWithRegistry:
//
//  1. Registered core.Cloner[T] — fastest; skips even the SelfClonable interface check.
//  2. core.SelfClonable[T] — used when no registration exists for T.
//
// Typical setup (once at startup, safe to share across goroutines thereafter):
//
//	reg := registry.New()
//	registry.Register(reg, core.NewFuncCloner(func(src MyType) (MyType, error) {
//	    return MyType{Field: src.Field}, nil
//	}))
//
//	cloned, err := doppel.CloneWithRegistry(value, reg)
package registry

import (
	"reflect"
	"sync"

	"github.com/seyallius/doppel/core"
)

// -------------------------------------------- Types, Variables & Constants --------------------------------------------

// Registry is a thread-safe, type-keyed store that maps reflect.Type keys to
// core.Cloner[T] values. The reflect package is used solely to derive a stable
// map key from T — never for value inspection or dynamic field access.
type Registry struct {
	mu          sync.RWMutex
	typeCloners map[reflect.Type]any // values are always core.Cloner[T] for the keyed T
}

// -------------------------------------------- Constructor(s) --------------------------------------------

// New creates and returns an empty, ready-to-use Registry.
// The returned Registry is safe for concurrent use immediately.
func New() *Registry {
	return &Registry{
		typeCloners: make(map[reflect.Type]any),
	}
}

// -------------------------------------------- Public API --------------------------------------------

// Register stores cloner as the Cloner[T] for type T.
// If a cloner is already registered for T it is silently replaced,
// making Register safe to call multiple times during initialization.
// Safe for concurrent use.
func Register[T any](r *Registry, cloner core.Cloner[T]) {
	key := typeKeyFor[T]()

	r.mu.Lock()
	defer r.mu.Unlock()

	r.typeCloners[key] = cloner
}

// Lookup retrieves the registered Cloner[T] for type T.
// Returns (cloner, true) when found, (nil, false) when not registered.
// Safe for concurrent use.
func Lookup[T any](r *Registry) (core.Cloner[T], bool) {
	key := typeKeyFor[T]()

	r.mu.RLock()
	defer r.mu.RUnlock()

	raw, found := r.typeCloners[key]
	if !found {
		return nil, false
	}

	// The type assertion here is always safe: Register[T] is the only writer,
	// and it always stores a core.Cloner[T] under the key for T.
	cloner, ok := raw.(core.Cloner[T])
	return cloner, ok
}

// Deregister removes the registered Cloner for type T, if any.
// Calling Deregister on a type that has no registration is a no-op.
// Safe for concurrent use.
func Deregister[T any](r *Registry) {
	key := typeKeyFor[T]()

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.typeCloners, key)
}

// Has reports whether a Cloner is registered for type T.
// Safe for concurrent use.
func Has[T any](r *Registry) bool {
	key := typeKeyFor[T]()

	r.mu.RLock()
	defer r.mu.RUnlock()

	_, found := r.typeCloners[key]
	return found
}

// Len returns the total number of registered cloners.
// Safe for concurrent use.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.typeCloners)
}

// -------------------------------------------- Internal Helpers --------------------------------------------

// typeKeyFor returns the canonical reflect.Type for T.
//
// We derive the key as reflect.TypeOf((*T)(nil)).Elem() rather than
// reflect.TypeOf(someValue) for two reasons:
//  1. It works correctly when T is an interface type (TypeOf on a nil
//     interface value yields nil, not the interface's reflect.Type).
//  2. It does not require a non-nil value, so callers like Lookup and
//     Deregister can obtain the key without any concrete instance.
//
// Example: When T is io.Reader
//   - (*io.Reader)(nil) creates a typed nil pointer to io.Reader
//   - reflect.TypeOf returns the *io.Reader type (pointers preserve type info even when nil)
//   - .Elem() dereferences to get the io.Reader interface type
//     Without this trick, reflect.TypeOf(var zero io.Reader) would return nil
//     because zero is a nil interface value with no dynamic type information.
func typeKeyFor[T any]() reflect.Type {
	typedNilPtr := (*T)(nil)                      // Create a typed nil pointer to T - this preserves type info even when nil
	ptrReflectType := reflect.TypeOf(typedNilPtr) // Get the reflect.Type of the pointer (e.g., *io.Reader, *int)
	typeOfT := ptrReflectType.Elem()              // Dereference to get the actual T type (e.g., io.Reader, int)

	return typeOfT
}

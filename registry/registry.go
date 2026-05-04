// Package registry. registry - implements a thread-safe, type-keyed store of Cloner[T] values,
// with additional support for field-level cloners that override per-field clone behavior
// within struct types.
//
// Reflection is used in this package exclusively for two purposes:
//  1. Deriving a stable map key from T (typeKeyFor).
//  2. Wrapping a stored Cloner[T] as a reflect-level function (LookupAny / LookupAnyField),
//     so the reflection engine in engine/ can consult registered cloners during graph traversal
//     without knowing T at compile time.
//
// Actual cloning is always delegated to the core.Cloner[T] the caller registered —
// no field traversal, no struct layout inspection, no dynamic value copying happens here.
//
// # Type-level cloners (Phase 2)
//
// Lookup priority inside doppel.CloneWithRegistry (Phase 2+4):
//
//  1. Registered Cloner[T] — fastest; skips SelfClonable and reflection entirely.
//  2. core.SelfClonable[T] — fallback when no registration exists for T.
//  3. Reflection engine    — final fallback introduced in Phase 4 (replaces ErrNoCloner).
//
// # Field-level cloners (Phase 3)
//
// Field-level cloners provide fine-grained control over individual struct fields.
// When the reflection engine clones a struct, it checks for registered field cloners
// before falling through to default reflection-based cloning for each field.
//
// This enables a "default deep copy + selective override" workflow:
//
//	reg := registry.New()
//
//	// Override clone behavior for specific fields only
//	registry.RegisterField[User, *Address](reg, "HomeAddr", core.NewFuncCloner(cloneAddr))
//	registry.RegisterField[User, []string](reg, "Tags", core.NewFuncCloner(func(src []string) ([]string, error) {
//	    return append([]string{}, src...), nil
//	}))
//
//	// CloneDeep uses reflection for all fields EXCEPT the ones with registered field cloners
//	cloned, err := doppel.CloneDeep(user, reg)
//
// The per-field priority chain inside the engine is:
//
//  1. Struct tag directive (doppel:"-", doppel:"shallow", doppel:"clone", etc.)
//  2. Registered field Cloner  (auto-discovered by field name)
//  3. Registered type Cloner   (for the field's type)
//  4. SelfClonable on the field value
//  5. Reflection fallback
//
// Typical setup (once at startup, safe to share across goroutines thereafter):
//
//	reg := registry.New()
//	registry.Register(reg, core.NewFuncCloner(func(src MyType) (MyType, error) {
//	    return MyType{Field: src.Field}, nil
//	}))
//	registry.RegisterField[BigStruct, *Nested](reg, "Nested", core.NewFuncCloner(cloneNested))
//
//	cloned, err := doppel.CloneDeep(value, reg)
package registry

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/seyallius/doppel/core"
)

// -------------------------------------------- Types, Variables & Constants --------------------------------------------

// fieldKey uniquely identifies a struct field for the field-level cloner map.
// The structType is the reflect.Type of the enclosing struct (value type, not pointer),
// and fieldName is the Go field name as declared in the struct definition.
type fieldKey struct {
	structType reflect.Type
	fieldName  string
}

// Registry is a thread-safe, type-keyed store that maps reflect.Type keys to
// core.Cloner[T] values, and (structType, fieldName) pairs to field-level cloners.
// The reflect package is used solely to derive stable map keys — never for value
// inspection or dynamic field access.
type Registry struct {
	mu           sync.RWMutex
	typeCloners  map[reflect.Type]any // values are always core.Cloner[T] for the keyed T
	fieldCloners map[fieldKey]any     // values are always core.Cloner[F] for the field's type F
}

// -------------------------------------------- Constructor(s) --------------------------------------------

// New creates and returns an empty, ready-to-use Registry.
// The returned Registry is safe for concurrent use immediately.
func New() *Registry {
	return &Registry{
		typeCloners:  make(map[reflect.Type]any),
		fieldCloners: make(map[fieldKey]any),
	}
}

// -------------------------------------------- Public API — Type-Level Cloners --------------------------------------------

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

// Len returns the total number of registered type-level cloners.
// Safe for concurrent use.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.typeCloners)
}

// LookupAny returns a reflect-level clone function for the given reflect.Type.
//
// It is the bridge between the type-safe generic registry and the reflection
// engine in engine/, which operates at reflect.Value level without knowing T
// at compile time. The engine calls this method to check whether a registered
// Cloner[T] exists for a type it encounters during graph traversal.
//
// The returned function accepts a reflect.Value of the registered type and
// returns a reflect.Value of the same type (or an error). The internal type
// assertion is always safe because Register[T] guarantees the stored value is
// core.Cloner[T] under the key for T.
//
// Returns (nil, false) when no Cloner is registered for t.
// Safe for concurrent use.
func (r *Registry) LookupAny(t reflect.Type) (func(reflect.Value) (reflect.Value, error), bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	raw, found := r.typeCloners[t]
	if !found {
		return nil, false
	}

	// Wrap the stored core.Cloner[T] (whose concrete type is unknown here) as a
	// reflect-level function by calling its Clone method through reflect.Value.
	clonerVal := reflect.ValueOf(raw)
	cloneMethod := clonerVal.MethodByName("Clone")

	return func(src reflect.Value) (reflect.Value, error) {
		results := cloneMethod.Call([]reflect.Value{src})
		errResult := results[1]
		if !errResult.IsNil() {
			return reflect.Value{}, errResult.Interface().(error)
		}
		return results[0], nil
	}, true
}

// -------------------------------------------- Public API — Field-Level Cloners --------------------------------------------

// RegisterField registers a Cloner for a specific field of struct type T.
// The cloner is invoked when the reflection engine encounters this field
// during a deep copy operation, overriding the default reflection-based cloning.
//
// T must be a struct type (or pointer to a struct), and fieldName must name an
// exported field of T whose type is compatible with F.
//
// If a field cloner is already registered for the same struct type and field name,
// it is silently replaced (consistent with Register behavior).
//
// Example:
//
//	registry.RegisterField[User, *Address](reg, "HomeAddr", core.NewFuncCloner(cloneAddr))
//	registry.RegisterField[User, []string](reg, "Tags", core.NewFuncCloner(
//	    func(src []string) ([]string, error) { return append([]string{}, src...), nil },
//	))
//
// Safe for concurrent use.
func RegisterField[T any, F any](r *Registry, fieldName string, cloner core.Cloner[F]) {
	structType := structTypeFor[T]()

	if structType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("doppel/registry: RegisterField: T must be a struct type, got %s", structType.Kind()))
	}

	field, found := structType.FieldByName(fieldName)
	if !found {
		panic(fmt.Sprintf("doppel/registry: RegisterField: struct %s has no field named %q", structType.Name(), fieldName))
	}

	if !field.IsExported() {
		panic(fmt.Sprintf("doppel/registry: RegisterField: field %s.%s is unexported; field cloners can only be registered for exported fields", structType.Name(), fieldName))
	}

	key := fieldKey{structType: structType, fieldName: fieldName}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.fieldCloners[key] = cloner
}

// LookupField retrieves the registered field-level Cloner[F] for a specific field
// of struct type T. Returns (cloner, true) when found, (nil, false) when not registered.
//
// Example:
//
//	cloner, found := registry.LookupField[User, *Address](reg, "HomeAddr")
//	if found {
//	    cloned, err := cloner.Clone(user.HomeAddr)
//	}
//
// Safe for concurrent use.
func LookupField[T any, F any](r *Registry, fieldName string) (core.Cloner[F], bool) {
	structType := structTypeFor[T]()
	key := fieldKey{structType: structType, fieldName: fieldName}

	r.mu.RLock()
	defer r.mu.RUnlock()

	raw, found := r.fieldCloners[key]
	if !found {
		return nil, false
	}

	cloner, ok := raw.(core.Cloner[F])
	return cloner, ok
}

// HasField reports whether a field-level Cloner is registered for the given
// struct type T and field name.
//
// Example:
//
//	if registry.HasField[User](reg, "HomeAddr") { ... }
//
// Safe for concurrent use.
func HasField[T any](r *Registry, fieldName string) bool {
	structType := structTypeFor[T]()
	key := fieldKey{structType: structType, fieldName: fieldName}

	r.mu.RLock()
	defer r.mu.RUnlock()

	_, found := r.fieldCloners[key]
	return found
}

// DeregisterField removes the registered field-level Cloner for the given
// struct type T and field name. Returns true if a cloner was removed,
// false if no cloner was registered for that field.
//
// Example:
//
//	removed := registry.DeregisterField[User](reg, "HomeAddr")
//
// Safe for concurrent use.
func DeregisterField[T any](r *Registry, fieldName string) bool {
	structType := structTypeFor[T]()
	key := fieldKey{structType: structType, fieldName: fieldName}

	r.mu.Lock()
	defer r.mu.Unlock()

	_, existed := r.fieldCloners[key]
	delete(r.fieldCloners, key)
	return existed
}

// FieldLen returns the total number of registered field-level cloners.
// Safe for concurrent use.
func (r *Registry) FieldLen() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.fieldCloners)
}

// LookupAnyField returns a reflect-level clone function for the given struct type
// and field name. It is the field-level counterpart of LookupAny, enabling the
// reflection engine to discover and invoke field-level cloners without knowing
// the concrete generic types at compile time.
//
// The returned function accepts a reflect.Value of the field's type and returns
// a reflect.Value of the same type (or an error).
//
// Returns (nil, false) when no field Cloner is registered for the given
// struct type and field name.
// Safe for concurrent use.
func (r *Registry) LookupAnyField(structType reflect.Type, fieldName string) (func(reflect.Value) (reflect.Value, error), bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := fieldKey{structType: structType, fieldName: fieldName}
	raw, found := r.fieldCloners[key]
	if !found {
		return nil, false
	}

	clonerVal := reflect.ValueOf(raw)
	cloneMethod := clonerVal.MethodByName("Clone")

	return func(src reflect.Value) (reflect.Value, error) {
		results := cloneMethod.Call([]reflect.Value{src})
		errResult := results[1]
		if !errResult.IsNil() {
			return reflect.Value{}, errResult.Interface().(error)
		}
		return results[0], nil
	}, true
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

// structTypeFor returns the reflect.Type for T, dereferencing any pointer
// to reach the underlying struct type. This is used by field-level cloner
// functions so that both RegisterField[User, ...] and RegisterField[*User, ...]
// resolve to the same struct type key.
//
// Panics if the resolved type is not a struct (field cloners only apply to struct fields).
func structTypeFor[T any]() reflect.Type {
	t := typeKeyFor[T]()
	// Dereference pointer types: RegisterField[*User, ...] and RegisterField[User, ...]
	// should resolve to the same struct type.
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

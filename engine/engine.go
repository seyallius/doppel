// Package engine. engine - Implements the reflection-based deep copy engine for doppel.
//
// This package is the LAST strategy in the priority chain:
//
//	Registered Cloner[T]  →  SelfClonable[T]  →  engine (reflection)
//
// The engine is consulted only when neither a registry cloner nor a SelfClonable
// implementation exists for a given type — it is never the default, always the fallback.
//
// # Reflection scope
//
// Reflection is used here for legitimate deep-copy purposes only:
//   - Reading and writing struct fields.
//   - Allocating new slices, maps, arrays, and pointers.
//   - Detecting whether a value implements the SelfClonable contract (method lookup).
//   - Dispatching to registered Cloner[T]s at every graph node (via TypeLookup).
//
// Unexported struct fields are skipped. To include unexported fields implement
// core.SelfClonable[T] on the type — that remains Phase 1 manual territory.
//
// # Struct field tags
//
// The engine respects the following doppel struct tags:
//
//	doppel:"-"       Skip the field entirely; the clone receives the zero value.
//	doppel:"shallow" Assign without recursing; the clone shares the field's value.
//
// # Cycle safety
//
// Pointer and map addresses are recorded in a per-clone visited map, preventing
// infinite recursion on cyclic graphs. Shared sub-graph structure is also preserved:
// if two pointers in the original point to the same allocation, the clone will too.
// Full cycle-detection semantics are tightened further in Phase 5.
//
// # Concurrency
//
// Engine is safe for concurrent use. All mutable state lives in the per-call
// cloneState value, which is never shared across goroutines.
package engine

import (
	"fmt"
	"reflect"

	"github.com/seyallius/doppel/core"
)

// -------------------------------------------- Types, Variables & Constants --------------------------------------------

// tagKey is the struct tag key consulted by the engine for field directives.
const tagKey = "doppel"

// errorType is the reflect.Type for the built-in error interface, used when
// detecting whether a value's Clone() method satisfies SelfClonable[T].
var errorType = reflect.TypeOf((*error)(nil)).Elem()

// TypeLookup is implemented by *registry.Registry. It allows the engine to
// consult registered Cloner[T]s at every node of the value graph without
// importing the registry package, keeping the dependency graph acyclic.
//
// *registry.Registry satisfies TypeLookup automatically via its LookupAny method —
// no explicit declaration is needed.
type TypeLookup interface {
	// LookupAny returns a reflect-level clone function for the given reflect.Type.
	// The returned function accepts and returns a reflect.Value of that type.
	// Returns (nil, false) when no Cloner is registered for t.
	LookupAny(t reflect.Type) (func(reflect.Value) (reflect.Value, error), bool)
}

// Engine performs reflection-based deep copying of arbitrary Go values.
//
// Construct one with New and call Clone for each top-level copy operation.
// A single Engine instance is safe for concurrent use because all mutable
// state lives in per-call cloneState values.
type Engine struct {
	lookup TypeLookup // nil if no registry was provided
}

// cloneState holds per-clone-call mutable state, keeping Engine itself stateless.
type cloneState struct {
	engine  *Engine
	visited map[uintptr]reflect.Value // ptr/map address → cloned value (cycle + sharing guard)
}

// -------------------------------------------- Constructor(s) --------------------------------------------

// New creates an Engine. Pass a *registry.Registry (or any TypeLookup) to have
// the engine consult registered Cloner[T]s at every node of the value graph.
// Pass nil to rely only on SelfClonable detection and pure reflection.
func New(lookup TypeLookup) *Engine {
	return &Engine{lookup: lookup}
}

// -------------------------------------------- Public API --------------------------------------------

// Clone deep-copies src and returns a reflect.Value of the same type.
// It is the entry point for a single top-level clone operation.
func (e *Engine) Clone(src reflect.Value) (reflect.Value, error) {
	state := &cloneState{
		engine:  e,
		visited: make(map[uintptr]reflect.Value),
	}
	return state.clone(src)
}

// -------------------------------------------- Internal Helpers --------------------------------------------

// clone is the central recursive dispatch. It walks src, consulting the registry
// and SelfClonable before falling back to kind-specific reflection.
func (s *cloneState) clone(src reflect.Value) (reflect.Value, error) {
	if !src.IsValid() {
		return reflect.Value{}, nil
	}

	// 1. Registry lookup — check if a Cloner[T] is registered for this exact type.
	if s.engine.lookup != nil {
		if cloneFn, ok := s.engine.lookup.LookupAny(src.Type()); ok {
			return cloneFn(src)
		}
	}

	// 2. SelfClonable detection — check for a Clone() (T, error) method on the
	//    value or its pointer, matching the core.SelfClonable[T] contract.
	if cloneFn, ok := detectSelfClonable(src); ok {
		return cloneFn()
	}

	// 3. Reflection fallback — kind-specific deep copy.
	return s.cloneByKind(src)
}

// cloneByKind dispatches to the appropriate kind-specific copy function.
func (s *cloneState) cloneByKind(src reflect.Value) (reflect.Value, error) {
	switch src.Kind() {
	case reflect.Ptr:
		return s.clonePointer(src)
	case reflect.Struct:
		return s.cloneStruct(src)
	case reflect.Slice:
		return s.cloneSlice(src)
	case reflect.Map:
		return s.cloneMap(src)
	case reflect.Array:
		return s.cloneArray(src)
	case reflect.Interface:
		return s.cloneInterface(src)

	// Primitive scalar types: assignment is already a complete deep copy because
	// these types carry no pointers and contain no heap allocations.
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		dst := reflect.New(src.Type()).Elem()
		dst.Set(src)
		return dst, nil

	// Chan, Func, and UnsafePointer carry reference semantics that cannot be
	// meaningfully deep-copied. They are shallow-copied (the reference is shared).
	default:
		dst := reflect.New(src.Type()).Elem()
		dst.Set(src)
		return dst, nil
	}
}

// clonePointer deep-copies a pointer value, preserving shared sub-graph structure.
//
// The cloned pointer is registered in s.visited BEFORE recursing so that back-edges
// in cyclic graphs resolve to the already-allocated clone rather than looping.
func (s *cloneState) clonePointer(src reflect.Value) (reflect.Value, error) {
	if src.IsNil() {
		return reflect.Zero(src.Type()), nil
	}

	addr := src.Pointer()
	if alreadyCloned, seen := s.visited[addr]; seen {
		return alreadyCloned, nil
	}

	dst := reflect.New(src.Type().Elem())
	s.visited[addr] = dst // register before recursing to break potential cycles

	cloned, err := s.clone(src.Elem())
	if err != nil {
		return reflect.Value{}, core.WrapError(src.Type().String(), err)
	}
	if cloned.IsValid() {
		dst.Elem().Set(cloned)
	}
	return dst, nil
}

// cloneStruct deep-copies a struct value, field by field.
//
// Unexported fields are skipped — they are inaccessible via reflection without
// unsafe. If unexported fields must be preserved, implement core.SelfClonable[T].
//
// Supported doppel struct tags:
//
//	doppel:"-"       Skip the field; clone receives the zero value.
//	doppel:"shallow" Assign without recursing; clone shares the field's value.
func (s *cloneState) cloneStruct(src reflect.Value) (reflect.Value, error) {
	dst := reflect.New(src.Type()).Elem()
	srcType := src.Type()

	for fieldIdx := 0; fieldIdx < srcType.NumField(); fieldIdx++ {
		fieldMeta := srcType.Field(fieldIdx)
		srcField := src.Field(fieldIdx)
		dstField := dst.Field(fieldIdx)

		// Apply doppel struct tags before anything else.
		switch fieldMeta.Tag.Get(tagKey) {
		case "-":
			continue // skip; dstField stays at zero value
		case "shallow":
			if fieldMeta.IsExported() {
				dstField.Set(srcField)
			}
			continue
		}

		// Unexported fields: skip with no error. Document the limitation clearly.
		if !fieldMeta.IsExported() {
			continue
		}

		cloned, err := s.clone(srcField)
		if err != nil {
			return reflect.Value{}, core.WrapError(
				fmt.Sprintf("%s.%s", srcType.Name(), fieldMeta.Name),
				err,
			)
		}
		if cloned.IsValid() {
			dstField.Set(cloned)
		}
	}
	return dst, nil
}

// cloneSlice deep-copies a slice, preserving the nil-vs-empty distinction.
func (s *cloneState) cloneSlice(src reflect.Value) (reflect.Value, error) {
	if src.IsNil() {
		return reflect.Zero(src.Type()), nil
	}

	dst := reflect.MakeSlice(src.Type(), src.Len(), src.Len())
	for elemIdx := 0; elemIdx < src.Len(); elemIdx++ {
		cloned, err := s.clone(src.Index(elemIdx))
		if err != nil {
			return reflect.Value{}, core.WrapError(fmt.Sprintf("[%d]", elemIdx), err)
		}
		if cloned.IsValid() {
			dst.Index(elemIdx).Set(cloned)
		}
	}
	return dst, nil
}

// cloneMap deep-copies a map, preserving the nil-vs-empty distinction.
// The cloned map is registered in s.visited to handle self-referential maps.
func (s *cloneState) cloneMap(src reflect.Value) (reflect.Value, error) {
	if src.IsNil() {
		return reflect.Zero(src.Type()), nil
	}

	dst := reflect.MakeMapWithSize(src.Type(), src.Len())
	s.visited[src.Pointer()] = dst // register before iterating to break self-referential cycles

	for _, key := range src.MapKeys() {
		clonedKey, err := s.clone(key)
		if err != nil {
			return reflect.Value{}, core.WrapError("map key", err)
		}
		clonedVal, err := s.clone(src.MapIndex(key))
		if err != nil {
			return reflect.Value{}, core.WrapError(fmt.Sprintf("map[%v]", key), err)
		}
		if clonedKey.IsValid() && clonedVal.IsValid() {
			dst.SetMapIndex(clonedKey, clonedVal)
		}
	}
	return dst, nil
}

// cloneArray deep-copies a fixed-length array.
func (s *cloneState) cloneArray(src reflect.Value) (reflect.Value, error) {
	dst := reflect.New(src.Type()).Elem()
	for elemIdx := 0; elemIdx < src.Len(); elemIdx++ {
		cloned, err := s.clone(src.Index(elemIdx))
		if err != nil {
			return reflect.Value{}, core.WrapError(fmt.Sprintf("[%d]", elemIdx), err)
		}
		if cloned.IsValid() {
			dst.Index(elemIdx).Set(cloned)
		}
	}
	return dst, nil
}

// cloneInterface deep-copies the concrete value stored inside an interface.
func (s *cloneState) cloneInterface(src reflect.Value) (reflect.Value, error) {
	if src.IsNil() {
		return reflect.Zero(src.Type()), nil
	}

	cloned, err := s.clone(src.Elem())
	if err != nil {
		return reflect.Value{}, err
	}

	dst := reflect.New(src.Type()).Elem()
	if cloned.IsValid() {
		dst.Set(cloned)
	}
	return dst, nil
}

// detectSelfClonable checks whether val or &val implements the SelfClonable[T]
// contract — i.e. has a method named Clone with signature func() (T, error).
//
// This is the reflect-level equivalent of the core.SelfClonable[T] interface check.
// We cannot use a generic interface assertion here because T is unknown at compile time.
//
// Returns a ready-to-call function and true if found; nil and false otherwise.
func detectSelfClonable(val reflect.Value) (func() (reflect.Value, error), bool) {
	if !val.CanInterface() {
		return nil, false
	}

	// Try the value itself — covers pointer types like *User with Clone() (*User, error).
	if callFn, ok := buildCloneCaller(val); ok {
		return callFn, true
	}

	// Try &val — covers value types with pointer receivers, when addressable.
	if val.Kind() != reflect.Ptr && val.CanAddr() {
		if callFn, ok := buildCloneCaller(val.Addr()); ok {
			return callFn, true
		}
	}

	return nil, false
}

// buildCloneCaller looks for a Clone() (T, error) method on val and returns a
// callable function if the signature matches the SelfClonable[T] contract.
func buildCloneCaller(val reflect.Value) (func() (reflect.Value, error), bool) {
	method := val.MethodByName("Clone")
	if !method.IsValid() {
		return nil, false
	}

	mt := method.Type()

	// Signature must be: func() (T, error) — no inputs, exactly two outputs.
	if mt.NumIn() != 0 || mt.NumOut() != 2 {
		return nil, false
	}

	// Second output must implement the error interface.
	if !mt.Out(1).Implements(errorType) {
		return nil, false
	}

	return func() (reflect.Value, error) {
		results := method.Call(nil)
		errResult := results[1]
		// errResult is of interface kind (error); IsNil() is safe here.
		if !errResult.IsNil() {
			return reflect.Value{}, errResult.Interface().(error)
		}
		return results[0], nil
	}, true
}

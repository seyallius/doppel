package engine_test

import (
	"errors"
	"reflect"
	"sync"
)

// plainStruct has only exported primitive fields — pure reflection path.
type plainStruct struct {
	Name   string
	Value  int
	Score  float64
	Active bool
}

// nestedStruct contains a pointer and a nested value struct.
type nestedStruct struct {
	Meta  plainStruct
	Child *plainStruct
	Count int
}

// sliceStruct exercises slice fields.
type sliceStruct struct {
	Tags    []string
	Numbers []int
	Ptrs    []*plainStruct
}

// mapStruct exercises map fields.
type mapStruct struct {
	Counts  map[string]int
	Records map[int]*plainStruct
}

// withUnexported has a mix of exported and unexported fields.
// The engine skips unexported fields; exported fields are cloned normally.
type withUnexported struct {
	Exported   string
	unexported int      // skipped by engine
	innerSlice []string // skipped by engine
}

// withTags exercises doppel struct tag directives.
type withTags struct {
	Normal  string
	Skipped string   `doppel:"-"`
	Shallow []string `doppel:"shallow"`
	Deep    []string
}

// cyclicNode is used to build self-referential pointer graphs.
type cyclicNode struct {
	ID   int
	Next *cyclicNode
}

// selfClonable implements core.SelfClonable[*selfClonable].
// When the engine encounters this type, it should call Clone() instead of reflecting.
type selfClonable struct {
	Data        string
	cloneCalled bool
}

// Clone implements core.SelfClonable[*selfClonable].
func (s *selfClonable) Clone() (*selfClonable, error) {
	return &selfClonable{Data: s.Data + "_cloned", cloneCalled: true}, nil
}

// stubbedLookup implements engine.TypeLookup for testing registry integration.
type stubbedLookup struct {
	mu       sync.RWMutex
	handlers map[reflect.Type]func(reflect.Value) (reflect.Value, error)
}

func (s *stubbedLookup) LookupAny(t reflect.Type) (func(reflect.Value) (reflect.Value, error), bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fn, ok := s.handlers[t]
	return fn, ok
}

// failingClonable has a Clone() that always returns an error.
type failingClonable struct{}

func (f *failingClonable) Clone() (*failingClonable, error) {
	return nil, errors.New("clone refused")
}

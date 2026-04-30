package engine_test

import (
	"reflect"
	"testing"

	"github.com/seyallius/doppel/engine"
)

func newStubbedLookup() *stubbedLookup {
	return &stubbedLookup{handlers: make(map[reflect.Type]func(reflect.Value) (reflect.Value, error))}
}

func (s *stubbedLookup) register(t reflect.Type, fn func(reflect.Value) (reflect.Value, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[t] = fn
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func cloneVia[T any](t *testing.T, src T, lookup engine.TypeLookup) T {
	t.Helper()
	eng := engine.New(lookup)
	cloned, err := eng.Clone(reflect.ValueOf(src))
	requireNoError(t, err)
	if !cloned.IsValid() {
		var zero T
		return zero
	}
	return cloned.Interface().(T)
}

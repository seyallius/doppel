package engine_test

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/seyallius/doppel/engine"
)

func TestEngine_Primitives(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input any
	}{
		{"bool_true", true},
		{"bool_false", false},
		{"int", 42},
		{"int_negative", -7},
		{"int8", int8(127)},
		{"int16", int16(-300)},
		{"int32", int32(1 << 20)},
		{"int64", int64(1 << 40)},
		{"uint", uint(999)},
		{"float32", float32(3.14)},
		{"float64", 3.14159},
		{"complex128", complex(1.0, 2.0)},
		{"string_empty", ""},
		{"string", "hello"},
		{"string_unicode", "日本語"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			eng := engine.New(nil)
			cloned, err := eng.Clone(reflect.ValueOf(tc.input))
			requireNoError(t, err)
			if !reflect.DeepEqual(cloned.Interface(), tc.input) {
				t.Errorf("Clone(%v) = %v; want %v", tc.input, cloned.Interface(), tc.input)
			}
		})
	}
}

func TestEngine_PlainStruct(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input plainStruct
	}{
		{name: "zero_value", input: plainStruct{}},
		{name: "fully_populated", input: plainStruct{Name: "Alice", Value: 7, Score: 9.5, Active: true}},
		{name: "negative_value", input: plainStruct{Name: "Bob", Value: -42}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cloned := cloneVia(t, tc.input, nil)

			if !reflect.DeepEqual(cloned, tc.input) {
				t.Fatalf("struct clone mismatch:\ngot  %+v\nwant %+v", cloned, tc.input)
			}
		})
	}
}

func TestEngine_NestedStruct(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input nestedStruct
	}{
		{
			name:  "with_nil_pointer",
			input: nestedStruct{Meta: plainStruct{Name: "root"}, Child: nil, Count: 1},
		},
		{
			name: "with_non_nil_pointer",
			input: nestedStruct{
				Meta:  plainStruct{Name: "parent", Value: 10},
				Child: &plainStruct{Name: "child", Value: 20},
				Count: 2,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cloned := cloneVia(t, tc.input, nil)

			if !reflect.DeepEqual(cloned, tc.input) {
				t.Fatalf("nested clone mismatch")
			}

			// Pointer independence.
			if tc.input.Child != nil {
				tc.input.Child.Name = "mutated"
				if cloned.Child.Name == "mutated" {
					t.Error("cloned pointer field shares memory with original")
				}
			}
		})
	}
}

func TestEngine_Slices(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input sliceStruct
	}{
		{name: "nil_slices", input: sliceStruct{}},
		{name: "empty_slices", input: sliceStruct{Tags: []string{}, Numbers: []int{}}},
		{
			name:  "string_slice",
			input: sliceStruct{Tags: []string{"a", "b", "c"}},
		},
		{
			name:  "int_slice",
			input: sliceStruct{Numbers: []int{10, 20, 30}},
		},
		{
			name: "slice_of_pointers",
			input: sliceStruct{Ptrs: []*plainStruct{
				{Name: "x", Value: 1},
				nil,
				{Name: "y", Value: 2},
			}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cloned := cloneVia(t, tc.input, nil)

			if !reflect.DeepEqual(cloned, tc.input) {
				t.Fatalf("slice clone mismatch:\ngot  %+v\nwant %+v", cloned, tc.input)
			}

			// Independence: mutating original slice elements must not affect clone.
			if len(tc.input.Tags) > 0 {
				tc.input.Tags[0] = "mutated"
				if cloned.Tags[0] == "mutated" {
					t.Error("string slice element not independently cloned")
				}
			}
		})
	}
}

func TestEngine_Maps(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input mapStruct
	}{
		{name: "nil_maps", input: mapStruct{}},
		{name: "empty_map", input: mapStruct{Counts: map[string]int{}}},
		{
			name:  "string_int_map",
			input: mapStruct{Counts: map[string]int{"a": 1, "b": 2, "c": 3}},
		},
		{
			name: "map_of_pointers",
			input: mapStruct{Records: map[int]*plainStruct{
				1: {Name: "first"},
				2: {Name: "second"},
			}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cloned := cloneVia(t, tc.input, nil)

			if !reflect.DeepEqual(cloned, tc.input) {
				t.Fatalf("map clone mismatch:\ngot  %+v\nwant %+v", cloned, tc.input)
			}

			// Independence.
			if tc.input.Counts != nil {
				tc.input.Counts["a"] = 9999
				if cloned.Counts["a"] == 9999 {
					t.Error("map value not independently cloned")
				}
			}
		})
	}
}

func TestEngine_Arrays(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "int_array",
			check: func(t *testing.T) {
				src := [4]int{10, 20, 30, 40}
				cloned := cloneVia(t, src, nil)
				if cloned != src {
					t.Errorf("int array mismatch: got %v, want %v", cloned, src)
				}
			},
		},
		{
			name: "struct_array",
			check: func(t *testing.T) {
				src := [2]plainStruct{{Name: "a"}, {Name: "b"}}
				cloned := cloneVia(t, src, nil)
				if !reflect.DeepEqual(cloned, src) {
					t.Errorf("struct array mismatch")
				}
			},
		},
		{
			name: "zero_length_array",
			check: func(t *testing.T) {
				src := [0]int{}
				cloned := cloneVia(t, src, nil)
				if cloned != src {
					t.Error("zero-length array mismatch")
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

func TestEngine_Interface(t *testing.T) {
	t.Parallel()

	type holder struct {
		Anything any
	}

	testCases := []struct {
		name  string
		input holder
	}{
		{name: "nil_interface", input: holder{Anything: nil}},
		{name: "int_in_interface", input: holder{Anything: 42}},
		{name: "string_in_interface", input: holder{Anything: "hello"}},
		{name: "struct_in_interface", input: holder{Anything: plainStruct{Name: "boxed"}}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cloned := cloneVia(t, tc.input, nil)
			if !reflect.DeepEqual(cloned, tc.input) {
				t.Fatalf("interface clone mismatch:\ngot  %+v\nwant %+v", cloned, tc.input)
			}
		})
	}
}

func TestEngine_NilValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "nil_pointer",
			check: func(t *testing.T) {
				var src *plainStruct
				cloned := cloneVia(t, src, nil)
				if cloned != nil {
					t.Error("nil pointer should clone to nil")
				}
			},
		},
		{
			name: "nil_slice",
			check: func(t *testing.T) {
				var src []int
				cloned := cloneVia(t, src, nil)
				if cloned != nil {
					t.Error("nil slice should clone to nil")
				}
			},
		},
		{
			name: "nil_map",
			check: func(t *testing.T) {
				var src map[string]int
				cloned := cloneVia(t, src, nil)
				if cloned != nil {
					t.Error("nil map should clone to nil")
				}
			},
		},
		{
			name: "empty_slice_is_not_nil",
			check: func(t *testing.T) {
				src := []int{}
				cloned := cloneVia(t, src, nil)
				if cloned == nil {
					t.Error("empty slice should clone to non-nil empty slice")
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

func TestEngine_StructTags(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input withTags
		check func(t *testing.T, cloned withTags, original withTags)
	}{
		{
			name:  "doppel_dash_skips_field",
			input: withTags{Normal: "kept", Skipped: "should_be_empty"},
			check: func(t *testing.T, cloned withTags, _ withTags) {
				if cloned.Skipped != "" {
					t.Errorf("doppel:\"-\" field should be zero: got %q", cloned.Skipped)
				}
				if cloned.Normal != "kept" {
					t.Errorf("Normal field should be cloned: got %q", cloned.Normal)
				}
			},
		},
		{
			name:  "doppel_shallow_shares_backing_array",
			input: withTags{Shallow: []string{"x", "y", "z"}, Deep: []string{"a", "b"}},
			check: func(t *testing.T, cloned withTags, original withTags) {
				// Shallow: mutation of original IS visible in clone (shared backing array).
				original.Shallow[0] = "mutated"
				if cloned.Shallow[0] != "mutated" {
					t.Error("shallow field should share backing array with original")
				}
				// Deep: mutation of original is NOT visible in clone.
				original.Deep[0] = "mutated"
				if cloned.Deep[0] == "mutated" {
					t.Error("deep field should not share backing array with original")
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			eng := engine.New(nil)
			clonedVal, err := eng.Clone(reflect.ValueOf(tc.input))
			requireNoError(t, err)
			cloned := clonedVal.Interface().(withTags)
			tc.check(t, cloned, tc.input)
		})
	}
}

func TestEngine_UnexportedFieldsSkipped(t *testing.T) {
	t.Parallel()

	// The engine skips unexported fields. Exported fields must still be cloned.
	src := withUnexported{
		Exported:   "visible",
		unexported: 42, // will be zero in clone
		innerSlice: []string{"a"},
	}

	eng := engine.New(nil)
	clonedVal, err := eng.Clone(reflect.ValueOf(src))
	requireNoError(t, err)
	cloned := clonedVal.Interface().(withUnexported)

	if cloned.Exported != src.Exported {
		t.Errorf("Exported field: got %q, want %q", cloned.Exported, src.Exported)
	}
	// unexported and innerSlice are zero in the clone — this is documented behaviour.
	// (reflect cannot access them; implement SelfClonable to include them.)
}

func TestEngine_SelfClonablePrecedesReflection(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "pointer_receiver_clone_is_called",
			check: func(t *testing.T) {
				src := &selfClonable{Data: "payload"}
				eng := engine.New(nil)
				clonedVal, err := eng.Clone(reflect.ValueOf(src))
				requireNoError(t, err)
				cloned := clonedVal.Interface().(*selfClonable)
				if !cloned.cloneCalled {
					t.Error("SelfClonable.Clone() was not invoked")
				}
				if cloned.Data != "payload_cloned" {
					t.Errorf("Data: got %q, want %q", cloned.Data, "payload_cloned")
				}
			},
		},
		{
			name: "error_from_self_clonable_propagates",
			check: func(t *testing.T) {
				src := &failingClonable{}
				eng := engine.New(nil)
				_, err := eng.Clone(reflect.ValueOf(src))
				if err == nil {
					t.Fatal("expected error from SelfClonable.Clone(), got nil")
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

func TestEngine_TypeLookupPrecedesSelfClonable(t *testing.T) {
	t.Parallel()

	// selfClonable implements Clone(), but a registered TypeLookup handler must
	// win because registry > SelfClonable in the priority chain.
	lookup := newStubbedLookup()
	lookup.register(
		reflect.TypeOf(&selfClonable{}),
		func(src reflect.Value) (reflect.Value, error) {
			sc := src.Interface().(*selfClonable)
			return reflect.ValueOf(&selfClonable{Data: sc.Data + "_from_lookup"}), nil
		},
	)

	src := &selfClonable{Data: "original"}
	eng := engine.New(lookup)
	clonedVal, err := eng.Clone(reflect.ValueOf(src))
	requireNoError(t, err)
	cloned := clonedVal.Interface().(*selfClonable)

	if cloned.cloneCalled {
		t.Error("SelfClonable.Clone() should NOT have been called when TypeLookup wins")
	}
	if cloned.Data != "original_from_lookup" {
		t.Errorf("Data: got %q, want %q", cloned.Data, "original_from_lookup")
	}
}

func TestEngine_SharedPointerPreserved(t *testing.T) {
	t.Parallel()

	// Two fields pointing to the same allocation — the clone must too.
	shared := &plainStruct{Name: "shared", Value: 42}
	type diamond struct {
		Left  *plainStruct
		Right *plainStruct
	}

	src := diamond{Left: shared, Right: shared}
	cloned := cloneVia(t, src, nil)

	if cloned.Left != cloned.Right {
		t.Error("shared reference not preserved: Left and Right should point to the same clone allocation")
	}
	if cloned.Left == src.Left {
		t.Error("clone should not share the original's allocation")
	}
}

func TestEngine_CyclicPointer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		build func() *cyclicNode
		check func(t *testing.T, cloned *cyclicNode)
	}{
		{
			name: "self_loop",
			build: func() *cyclicNode {
				n := &cyclicNode{ID: 1}
				n.Next = n
				return n
			},
			check: func(t *testing.T, cloned *cyclicNode) {
				if cloned.ID != 1 {
					t.Errorf("ID: got %d, want 1", cloned.ID)
				}
				if cloned.Next != cloned {
					t.Error("self-loop not preserved in clone")
				}
			},
		},
		{
			name: "two_node_cycle",
			build: func() *cyclicNode {
				a := &cyclicNode{ID: 1}
				b := &cyclicNode{ID: 2}
				a.Next = b
				b.Next = a
				return a
			},
			check: func(t *testing.T, cloned *cyclicNode) {
				if cloned.Next == nil || cloned.Next.ID != 2 {
					t.Error("second node missing or wrong ID")
				}
				if cloned.Next.Next != cloned {
					t.Error("two-node cycle not preserved")
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			src := tc.build()
			eng := engine.New(nil)
			clonedVal, err := eng.Clone(reflect.ValueOf(src))
			requireNoError(t, err)
			tc.check(t, clonedVal.Interface().(*cyclicNode))
		})
	}
}

func TestEngine_TypeLookupErrorPropagates(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("lookup failure")
	lookup := newStubbedLookup()
	lookup.register(
		reflect.TypeOf(plainStruct{}),
		func(_ reflect.Value) (reflect.Value, error) {
			return reflect.Value{}, sentinel
		},
	)

	eng := engine.New(lookup)
	_, err := eng.Clone(reflect.ValueOf(plainStruct{Name: "test"}))

	if err == nil {
		t.Fatal("expected error from TypeLookup handler, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("errors.Is failed: %v", err)
	}
}

func TestEngine_DeepNestedComposition(t *testing.T) {
	t.Parallel()

	// A realistic aggregate with nesting across all supported kinds.
	type inner struct {
		Tag   string
		Count int
	}
	type outer struct {
		Name     string
		Items    []inner
		Lookup   map[string]*inner
		Metadata map[string]any
	}

	src := outer{
		Name: "root",
		Items: []inner{
			{Tag: "alpha", Count: 1},
			{Tag: "beta", Count: 2},
		},
		Lookup: map[string]*inner{
			"x": {Tag: "x-inner", Count: 10},
		},
		Metadata: map[string]any{
			"version": 3,
			"labels":  []string{"a", "b"},
		},
	}

	cloned := cloneVia(t, src, nil)

	if !reflect.DeepEqual(cloned, src) {
		t.Fatalf("deep nested clone mismatch:\ngot  %+v\nwant %+v", cloned, src)
	}

	// Independence checks.
	src.Items[0].Tag = "mutated"
	if cloned.Items[0].Tag == "mutated" {
		t.Error("cloned slice element not independent")
	}

	src.Lookup["x"].Count = 999
	if cloned.Lookup["x"].Count == 999 {
		t.Error("cloned map pointer value not independent")
	}
}

func TestEngine_Concurrency(t *testing.T) {
	t.Parallel()

	src := nestedStruct{
		Meta:  plainStruct{Name: "concurrent", Value: 1},
		Child: &plainStruct{Name: "child", Value: 2},
		Count: 99,
	}

	const goroutineCount = 50
	errCh := make(chan error, goroutineCount)
	var wg sync.WaitGroup

	for range goroutineCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			eng := engine.New(nil)
			clonedVal, err := eng.Clone(reflect.ValueOf(src))
			if err != nil {
				errCh <- err
				return
			}
			cloned := clonedVal.Interface().(nestedStruct)
			if !reflect.DeepEqual(cloned, src) {
				errCh <- fmt.Errorf("concurrent clone mismatch")
			}
		}()
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Error(err)
	}
}

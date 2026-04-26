package manual_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/seyallius/doppel/manual"
)

// ---------------------------------------------------------------------------
// CloneMap — fallible value cloner
// ---------------------------------------------------------------------------

func TestCloneMap_StringInt(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		src        map[string]int
		wantResult map[string]int
		wantNil    bool
		wantErr    bool
	}{
		{
			name:    "nil_map_returns_nil",
			src:     nil,
			wantNil: true,
		},
		{
			name:       "empty_map_returns_empty_non_nil",
			src:        map[string]int{},
			wantResult: map[string]int{},
		},
		{
			name:       "single_entry",
			src:        map[string]int{"score": 100},
			wantResult: map[string]int{"score": 100},
		},
		{
			name:       "multiple_entries_all_copied",
			src:        map[string]int{"a": 1, "b": 2, "c": 3},
			wantResult: map[string]int{"a": 1, "b": 2, "c": 3},
		},
		{
			name:       "value_transformation_applied",
			src:        map[string]int{"x": 5, "y": 10},
			wantResult: map[string]int{}, // overridden in cloneVal below
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cloneVal := manual.Identity[int]

			// For the transformation test, double each value.
			if tc.name == "value_transformation_applied" {
				cloneVal = func(v int) (int, error) { return v * 2, nil }
				tc.wantResult = map[string]int{"x": 10, "y": 20}
			}

			got, err := manual.CloneMap(tc.src, cloneVal)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if !reflect.DeepEqual(got, tc.wantResult) {
				t.Fatalf("result mismatch:\ngot  %v\nwant %v", got, tc.wantResult)
			}
		})
	}
}

func TestCloneMap_IntString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		src        map[int]string
		wantResult map[int]string
		wantNil    bool
	}{
		{name: "nil", src: nil, wantNil: true},
		{name: "empty", src: map[int]string{}, wantResult: map[int]string{}},
		{name: "populated", src: map[int]string{1: "one", 2: "two"}, wantResult: map[int]string{1: "one", 2: "two"}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := manual.CloneMap(tc.src, manual.Identity[string])
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if !reflect.DeepEqual(got, tc.wantResult) {
				t.Fatalf("got %v, want %v", got, tc.wantResult)
			}
		})
	}
}

func TestCloneMap_Independence(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		src    map[string]int
		mutate func(map[string]int)
		check  func(t *testing.T, cloned map[string]int)
	}{
		{
			name:   "add_key_to_original_does_not_appear_in_clone",
			src:    map[string]int{"existing": 1},
			mutate: func(m map[string]int) { m["new_key"] = 99 },
			check: func(t *testing.T, cloned map[string]int) {
				if _, found := cloned["new_key"]; found {
					t.Error("clone contains key added to original after cloning")
				}
			},
		},
		{
			name:   "mutate_original_value_does_not_change_clone",
			src:    map[string]int{"counter": 10},
			mutate: func(m map[string]int) { m["counter"] = 999 },
			check: func(t *testing.T, cloned map[string]int) {
				if cloned["counter"] != 10 {
					t.Errorf("clone value changed: got %d, want 10", cloned["counter"])
				}
			},
		},
		{
			name:   "delete_key_from_original_does_not_affect_clone",
			src:    map[string]int{"keep": 7, "delete_me": 0},
			mutate: func(m map[string]int) { delete(m, "delete_me") },
			check: func(t *testing.T, cloned map[string]int) {
				if _, found := cloned["delete_me"]; !found {
					t.Error("key was removed from clone when original was mutated")
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cloned, err := manual.CloneMap(tc.src, manual.Identity[int])
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tc.mutate(tc.src)
			tc.check(t, cloned)
		})
	}
}

func TestCloneMap_ErrorPropagation(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("value clone failure")
	cloneVal := func(v int) (int, error) {
		if v < 0 {
			return 0, sentinel
		}
		return v, nil
	}

	src := map[string]int{"positive": 5, "negative": -1}
	_, err := manual.CloneMap(src, cloneVal)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("errors.Is failed: %v", err)
	}
}

func TestCloneMap_ConditionalClone(t *testing.T) {
	t.Parallel()

	// Demonstrates the Phase 3 field-level use-case preview: the caller
	// controls which values make it into the clone.
	src := map[string]int{
		"alpha": 10,
		"beta":  3,
		"gamma": 15,
		"delta": 1,
	}

	// Clone only entries where the value is >= 10 (copy or zero-out).
	cloneVal := func(v int) (int, error) {
		if v < 10 {
			return 0, nil // zero-out below-threshold values
		}
		return v, nil
	}

	cloned, err := manual.CloneMap(src, cloneVal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[string]int{"alpha": 10, "beta": 0, "gamma": 15, "delta": 0}
	if !reflect.DeepEqual(cloned, want) {
		t.Fatalf("conditional clone mismatch:\ngot  %v\nwant %v", cloned, want)
	}
}

// ---------------------------------------------------------------------------
// CloneMapOf — infallible value cloner
// ---------------------------------------------------------------------------

func TestCloneMapOf(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		src        map[string]int
		wantResult map[string]int
		wantNil    bool
	}{
		{name: "nil_returns_nil", src: nil, wantNil: true},
		{name: "empty_returns_empty", src: map[string]int{}, wantResult: map[string]int{}},
		{name: "values_copied", src: map[string]int{"a": 1, "b": 2}, wantResult: map[string]int{"a": 1, "b": 2}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := manual.CloneMapOf(tc.src, manual.IdentityValue[int])

			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if !reflect.DeepEqual(got, tc.wantResult) {
				t.Fatalf("got %v, want %v", got, tc.wantResult)
			}
		})
	}
}

func TestCloneMapOf_Independence(t *testing.T) {
	t.Parallel()

	src := map[string]int{"x": 1}
	cloned := manual.CloneMapOf(src, manual.IdentityValue[int])

	src["x"] = 999
	if cloned["x"] == 999 {
		t.Error("CloneMapOf clone shares storage with original")
	}
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkCloneMap_StringInt_50(b *testing.B) {
	src := makeStringIntMap(50)
	b.ResetTimer()
	for range b.N {
		_, _ = manual.CloneMap(src, manual.Identity[int])
	}
}

func BenchmarkCloneMap_StringInt_500(b *testing.B) {
	src := makeStringIntMap(500)
	b.ResetTimer()
	for range b.N {
		_, _ = manual.CloneMap(src, manual.Identity[int])
	}
}

func BenchmarkCloneMapOf_StringInt_500(b *testing.B) {
	src := makeStringIntMap(500)
	b.ResetTimer()
	for range b.N {
		_ = manual.CloneMapOf(src, manual.IdentityValue[int])
	}
}

func BenchmarkShallowCopy_StringInt_500(b *testing.B) {
	src := makeStringIntMap(500)
	b.ResetTimer()
	for range b.N {
		dst := make(map[string]int, len(src))
		for k, v := range src {
			dst[k] = v
		}
		_ = dst
	}
}

// makeStringIntMap creates a map[string]int with n entries for benchmarking.
func makeStringIntMap(n int) map[string]int {
	m := make(map[string]int, n)
	for idx := 0; idx < n; idx++ {
		m[makeStringSlice(1)[0]] = idx // reuse helper for key uniqueness
		m[key(idx)] = idx * 7
	}
	return m
}

func key(idx int) string {
	return "key_" + string(rune('a'+idx%26)) + "_" + string(rune('0'+idx%10))
}

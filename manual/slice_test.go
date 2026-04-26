package manual_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/seyallius/doppel/manual"
)

// ---------------------------------------------------------------------------
// CloneSlice — fallible element cloner
// ---------------------------------------------------------------------------

func TestCloneSlice(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		src        []string
		cloneElem  func(string) (string, error)
		wantResult []string
		wantNil    bool
		wantErr    bool
	}{
		{
			name:      "nil_slice_returns_nil",
			src:       nil,
			cloneElem: manual.Identity[string],
			wantNil:   true,
		},
		{
			name:       "empty_slice_returns_empty_non_nil",
			src:        []string{},
			cloneElem:  manual.Identity[string],
			wantResult: []string{},
		},
		{
			name:       "single_element",
			src:        []string{"hello"},
			cloneElem:  manual.Identity[string],
			wantResult: []string{"hello"},
		},
		{
			name:       "multiple_elements_order_preserved",
			src:        []string{"a", "b", "c", "d"},
			cloneElem:  manual.Identity[string],
			wantResult: []string{"a", "b", "c", "d"},
		},
		{
			name:       "element_transformation_applied",
			src:        []string{"x", "y"},
			cloneElem:  func(s string) (string, error) { return s + "_copy", nil },
			wantResult: []string{"x_copy", "y_copy"},
		},
		{
			name: "error_from_cloneElem_propagates",
			src:  []string{"ok", "fail", "ok"},
			cloneElem: func(s string) (string, error) {
				if s == "fail" {
					return "", errors.New("synthetic failure")
				}
				return s, nil
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := manual.CloneSlice(tc.src, tc.cloneElem)

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
					t.Fatalf("expected nil slice, got %v", got)
				}
				return
			}

			if !reflect.DeepEqual(got, tc.wantResult) {
				t.Fatalf("result mismatch:\ngot  %v\nwant %v", got, tc.wantResult)
			}
		})
	}
}

func TestCloneSlice_Int(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		src        []int
		wantResult []int
		wantNil    bool
	}{
		{name: "nil", src: nil, wantNil: true},
		{name: "empty", src: []int{}, wantResult: []int{}},
		{name: "single", src: []int{42}, wantResult: []int{42}},
		{name: "many", src: []int{1, 2, 3, 4, 5}, wantResult: []int{1, 2, 3, 4, 5}},
		{name: "negatives", src: []int{-3, 0, 7}, wantResult: []int{-3, 0, 7}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := manual.CloneSlice(tc.src, manual.Identity[int])
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

func TestCloneSlice_Independence(t *testing.T) {
	t.Parallel()

	// Mutating the original after cloning must not affect the clone.
	testCases := []struct {
		name   string
		src    []string
		mutate func([]string)
		check  func(t *testing.T, cloned []string)
	}{
		{
			name:   "append_to_original_does_not_grow_clone",
			src:    []string{"a", "b"},
			mutate: func(s []string) { _ = append(s, "c") }, // harmless but explicit
			check: func(t *testing.T, cloned []string) {
				if len(cloned) != 2 {
					t.Errorf("expected len 2, got %d", len(cloned))
				}
			},
		},
		{
			name:   "overwrite_original_element_does_not_affect_clone",
			src:    []string{"original", "value"},
			mutate: func(s []string) { s[0] = "mutated" },
			check: func(t *testing.T, cloned []string) {
				if cloned[0] != "original" {
					t.Errorf("clone was affected by mutation: got %q, want %q", cloned[0], "original")
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cloned, err := manual.CloneSlice(tc.src, manual.Identity[string])
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tc.mutate(tc.src)
			tc.check(t, cloned)
		})
	}
}

func TestCloneSlice_ErrorContext(t *testing.T) {
	t.Parallel()

	// The error returned for index 2 must mention that index in its message.
	sentinel := errors.New("boom")
	callCount := 0
	cloneElem := func(v int) (int, error) {
		callCount++
		if callCount == 3 { // third element = index 2
			return 0, sentinel
		}
		return v, nil
	}

	_, err := manual.CloneSlice([]int{10, 20, 30, 40}, cloneElem)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("errors.Is failed: %v", err)
	}

	errMsg := err.Error()
	if !contains(errMsg, "2") {
		t.Errorf("expected error message to contain index 2, got: %s", errMsg)
	}
}

// ---------------------------------------------------------------------------
// CloneSliceOf — infallible element cloner
// ---------------------------------------------------------------------------

func TestCloneSliceOf(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		src        []string
		wantResult []string
		wantNil    bool
	}{
		{name: "nil_returns_nil", src: nil, wantNil: true},
		{name: "empty_returns_empty", src: []string{}, wantResult: []string{}},
		{name: "values_copied", src: []string{"x", "y", "z"}, wantResult: []string{"x", "y", "z"}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := manual.CloneSliceOf(tc.src, manual.IdentityValue[string])

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

func TestCloneSliceOf_Independence(t *testing.T) {
	t.Parallel()

	src := []int{1, 2, 3}
	cloned := manual.CloneSliceOf(src, manual.IdentityValue[int])

	src[0] = 999
	if cloned[0] == 999 {
		t.Error("CloneSliceOf clone shares memory with original")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// contains reports whether sub is a substring of s.
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkCloneSlice_Strings_10(b *testing.B) {
	src := makeStringSlice(10)
	b.ResetTimer()
	for range b.N {
		_, _ = manual.CloneSlice(src, manual.Identity[string])
	}
}

func BenchmarkCloneSlice_Strings_100(b *testing.B) {
	src := makeStringSlice(100)
	b.ResetTimer()
	for range b.N {
		_, _ = manual.CloneSlice(src, manual.Identity[string])
	}
}

func BenchmarkCloneSlice_Strings_1000(b *testing.B) {
	src := makeStringSlice(1000)
	b.ResetTimer()
	for range b.N {
		_, _ = manual.CloneSlice(src, manual.Identity[string])
	}
}

func BenchmarkCloneSliceOf_Strings_1000(b *testing.B) {
	src := makeStringSlice(1000)
	b.ResetTimer()
	for range b.N {
		_ = manual.CloneSliceOf(src, manual.IdentityValue[string])
	}
}

func BenchmarkShallowCopy_Strings_1000(b *testing.B) {
	src := makeStringSlice(1000)
	b.ResetTimer()
	for range b.N {
		dst := make([]string, len(src))
		copy(dst, src)
		_ = dst
	}
}

func BenchmarkCloneSlice_Ints_1000(b *testing.B) {
	src := makeIntSlice(1000)
	b.ResetTimer()
	for range b.N {
		_, _ = manual.CloneSlice(src, manual.Identity[int])
	}
}

func BenchmarkShallowCopy_Ints_1000(b *testing.B) {
	src := makeIntSlice(1000)
	b.ResetTimer()
	for range b.N {
		dst := make([]int, len(src))
		copy(dst, src)
		_ = dst
	}
}

// ---------------------------------------------------------------------------
// Benchmark helpers
// ---------------------------------------------------------------------------

func makeStringSlice(size int) []string {
	slice := make([]string, size)
	for idx := range slice {
		slice[idx] = fmt.Sprintf("element_%d", idx)
	}
	return slice
}

func makeIntSlice(size int) []int {
	slice := make([]int, size)
	for idx := range slice {
		slice[idx] = idx * 3
	}
	return slice
}

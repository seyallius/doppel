package manual_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/seyallius/doppel/manual"
)

// --- CloneSlice — fallible element cloner --------------------

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
			mutate: func(s []string) { _ = append(s, "c") },
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

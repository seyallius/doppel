package manual_test

import (
	"errors"
	"testing"

	"github.com/seyallius/doppel/manual"
)

// --- ClonePointer — fallible value cloner --------------------

func TestClonePointer_Int(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		src       *int
		wantValue *int // nil means expect nil result
		wantErr   bool
	}{
		{
			name:      "nil_pointer_returns_nil",
			src:       nil,
			wantValue: nil,
		},
		{
			name:      "pointer_to_zero",
			src:       intPointer(0),
			wantValue: intPointer(0),
		},
		{
			name:      "pointer_to_positive_int",
			src:       intPointer(42),
			wantValue: intPointer(42),
		},
		{
			name:      "pointer_to_negative_int",
			src:       intPointer(-99),
			wantValue: intPointer(-99),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := manual.ClonePointer(tc.src, manual.Identity[int])

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.wantValue == nil {
				if got != nil {
					t.Fatalf("expected nil pointer, got %v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("expected non-nil pointer, got nil")
			}
			if *got != *tc.wantValue {
				t.Fatalf("value mismatch: got %d, want %d", *got, *tc.wantValue)
			}
		})
	}
}

func TestClonePointer_Isolation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		src    *int
		mutate func(*int)
		verify func(t *testing.T, cloned *int)
	}{
		{
			name:   "mutating_original_does_not_affect_clone",
			src:    intPointer(10),
			mutate: func(p *int) { *p = 999 },
			verify: func(t *testing.T, cloned *int) {
				if *cloned != 10 {
					t.Errorf("clone was affected by original mutation: got %d, want 10", *cloned)
				}
			},
		},
		{
			name:   "clone_and_original_have_different_addresses",
			src:    intPointer(7),
			mutate: func(_ *int) {},
			verify: func(t *testing.T, cloned *int) {
				// The test fixture already re-clones below; we just check the address.
				if cloned == nil {
					t.Fatal("cloned pointer is nil")
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cloned, err := manual.ClonePointer(tc.src, manual.Identity[int])
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.src != nil && cloned == tc.src {
				t.Error("clone must be a different allocation from original")
			}

			tc.mutate(tc.src)
			tc.verify(t, cloned)
		})
	}
}

func TestClonePointer_ErrorPropagation(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("clone failed")
	failingClone := func(v int) (int, error) {
		return 0, sentinel
	}

	src := intPointer(5)
	_, err := manual.ClonePointer(src, failingClone)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("errors.Is failed: %v", err)
	}
}

func TestClonePointer_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		src       *string
		wantValue *string
	}{
		{name: "nil", src: nil, wantValue: nil},
		{name: "empty_string", src: stringPointer(""), wantValue: stringPointer("")},
		{name: "non_empty", src: stringPointer("hello"), wantValue: stringPointer("hello")},
		{name: "unicode", src: stringPointer("日本語"), wantValue: stringPointer("日本語")},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := manual.ClonePointer(tc.src, manual.Identity[string])
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.wantValue == nil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("unexpected nil result")
			}
			if *got != *tc.wantValue {
				t.Fatalf("got %q, want %q", *got, *tc.wantValue)
			}
		})
	}
}

// --- ClonePointerOf — infallible value cloner --------------------

func TestClonePointerOf(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		src        *int
		wantResult *int
	}{
		{name: "nil_returns_nil", src: nil, wantResult: nil},
		{name: "value_copied", src: intPointer(77), wantResult: intPointer(77)},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := manual.ClonePointerOf(tc.src, manual.IdentityValue[int])

			if tc.wantResult == nil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if got == nil || *got != *tc.wantResult {
				t.Fatalf("got %v, want %v", derefInt(got), derefInt(tc.wantResult))
			}
			if got == tc.src {
				t.Error("ClonePointerOf returned the same pointer as src")
			}
		})
	}
}

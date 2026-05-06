package core_test

import (
	"testing"
)

func TestParseTagValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want TagValue
	}{
		{"skip", "-", TagSkip},
		{"shallow", "shallow", TagShallow},
		{"clone", "clone", TagClone},
		{"deep", "deep", TagDeep},
		{"empty", "empty", TagEmpty},
		{"empty_string", "", TagDeep},
		{"unknown", "something_else", TagDeep},
		{"readonly_rejected", "readonly", TagDeep},
		{"mixed_case_rejected", "Shallow", TagDeep},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ParseTagValue(tc.raw)
			if got != tc.want {
				t.Errorf("ParseTagValue(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestParseTag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want TagDirective
	}{
		{"skip", "-", TagDirective{Skip: true}},
		{"shallow", "shallow", TagDirective{Shallow: true}},
		{"clone", "clone", TagDirective{Clone: true}},
		{"deep", "deep", TagDirective{Deep: true}},
		{"empty", "empty", TagDirective{Empty: true}},
		{"default", "", TagDirective{Deep: true}},
		{"unknown", "foo", TagDirective{Deep: true}},
		{"readonly_falls_back", "readonly", TagDirective{Deep: true}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ParseTag(tc.raw)
			if got != tc.want {
				t.Errorf("ParseTag(%q) = %+v, want %+v", tc.raw, got, tc.want)
			}
		})
	}
}

func TestTagValueConstants(t *testing.T) {
	t.Parallel()

	// Verify constants are not empty strings
	if TagSkip == "" {
		t.Error("TagSkip should not be empty")
	}
	if TagShallow == "" {
		t.Error("TagShallow should not be empty")
	}
	if TagClone == "" {
		t.Error("TagClone should not be empty")
	}
	if TagDeep == "" {
		t.Error("TagDeep should not be empty")
	}
	if TagEmpty == "" {
		t.Error("TagEmpty should not be empty")
	}

	// Verify mutual exclusivity
	all := []TagValue{TagSkip, TagShallow, TagClone, TagDeep, TagEmpty}
	seen := make(map[TagValue]bool)
	for _, v := range all {
		if seen[v] {
			t.Errorf("duplicate TagValue: %q", v)
		}
		seen[v] = true
	}

	// Verify each constant parses to itself
	for _, v := range all {
		parsed := ParseTagValue(string(v))
		if parsed != v {
			t.Errorf("ParseTagValue(%q) = %q, want %q", string(v), parsed, v)
		}
	}
}

func TestTagKey(t *testing.T) {
	if TagKey != "doppel" {
		t.Errorf("TagKey = %q, want %q", TagKey, "doppel")
	}
}

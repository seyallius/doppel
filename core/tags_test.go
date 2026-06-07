package core_test

import (
	"testing"

	"github.com/seyallius/doppel/core"
)

func TestParseTagValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want core.TagValue
	}{
		{"skip", "-", core.TagSkip},
		{"shallow", "shallow", core.TagShallow},
		{"clone", "clone", core.TagClone},
		{"deep", "deep", core.TagDeep},
		{"empty", "empty", core.TagEmpty},
		{"empty_string", "", core.TagDeep},
		{"unknown", "something_else", core.TagDeep},
		{"readonly_rejected", "readonly", core.TagDeep},
		{"mixed_case_rejected", "Shallow", core.TagDeep},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := core.ParseTagValue(tc.raw)
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
		want core.TagDirective
	}{
		{"skip", "-", core.TagDirective{Skip: true}},
		{"shallow", "shallow", core.TagDirective{Shallow: true}},
		{"clone", "clone", core.TagDirective{Clone: true}},
		{"deep", "deep", core.TagDirective{Deep: true}},
		{"empty", "empty", core.TagDirective{Empty: true}},
		{"default", "", core.TagDirective{Deep: true}},
		{"unknown", "foo", core.TagDirective{Deep: true}},
		{"readonly_falls_back", "readonly", core.TagDirective{Deep: true}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := core.ParseTag(tc.raw)
			if got != tc.want {
				t.Errorf("ParseTag(%q) = %+v, want %+v", tc.raw, got, tc.want)
			}
		})
	}
}

func TestTagValueConstants(t *testing.T) {
	t.Parallel()

	// Verify constants are not empty strings
	if core.TagSkip == "" {
		t.Error("TagSkip should not be empty")
	}
	if core.TagShallow == "" {
		t.Error("TagShallow should not be empty")
	}
	if core.TagClone == "" {
		t.Error("TagClone should not be empty")
	}
	if core.TagDeep == "" {
		t.Error("TagDeep should not be empty")
	}
	if core.TagEmpty == "" {
		t.Error("TagEmpty should not be empty")
	}

	// Verify mutual exclusivity
	all := []core.TagValue{core.TagSkip, core.TagShallow, core.TagClone, core.TagDeep, core.TagEmpty}
	seen := make(map[core.TagValue]bool)
	for _, v := range all {
		if seen[v] {
			t.Errorf("duplicate TagValue: %q", v)
		}
		seen[v] = true
	}

	// Verify each constant parses to itself
	for _, v := range all {
		parsed := core.ParseTagValue(string(v))
		if parsed != v {
			t.Errorf("ParseTagValue(%q) = %q, want %q", string(v), parsed, v)
		}
	}
}

func TestTagKey(t *testing.T) {
	if core.TagKey != "doppel" {
		t.Errorf("TagKey = %q, want %q", core.TagKey, "doppel")
	}
}

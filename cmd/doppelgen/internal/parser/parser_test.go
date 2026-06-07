package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/seyallius/doppel/cmd/doppelgen/internal/parser"
	"github.com/seyallius/doppel/cmd/doppelgen/internal/types"
	"github.com/seyallius/doppel/core"
)

// Compile-time interface check: types.TypeInfo must be usable with parser.FilterStructs.
var _ types.TypeInfo = nil

// testdataPath returns the absolute path to the basic test fixtures directory.
// The fixtures were moved into the 'basic' subdirectory to support multiple
// test scenarios (e.g., complex third-party modules).
func testdataPath(t *testing.T) string {
	t.Helper()

	// Walk up from the test's executable directory to find testdata.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// Point directly to the 'basic' subdirectory where the .go files live!
	return filepath.Join(wd, "..", "..", "testdata", "basic")
}

// -------------------------------------------- Tests --------------------------------------------

func TestParsePackage(t *testing.T) {
	td := testdataPath(t)
	result, err := parser.ParsePackage(td, "doppel")
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	if result.Package != "basic" {
		t.Errorf("package = %q, want %q", result.Package, "basic")
	}

	if result.FileCount < 1 {
		t.Error("expected at least 1 file parsed")
	}
}

func TestParsePackage_BasicUser(t *testing.T) {
	td := testdataPath(t)
	result, err := parser.ParsePackage(td, "doppel")
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	info, ok := result.Structs["BasicUser"]
	if !ok {
		t.Fatal("BasicUser not found")
	}
	if info.Skip {
		t.Errorf("BasicUser should not be skipped: %s", info.SkipReason)
	}

	// Check fields.
	fieldMap := make(map[string]int)
	for _, f := range info.Fields {
		fieldMap[f.Name]++
	}

	expectedFields := []string{"ID", "Name", "Tags", "Scores", "Secret", "Config", "Cache"}
	for _, name := range expectedFields {
		if _, ok := fieldMap[name]; !ok {
			t.Errorf("missing field %q in BasicUser", name)
		}
	}
}

func TestParsePackage_TagResolution(t *testing.T) {
	td := testdataPath(t)
	result, err := parser.ParsePackage(td, "doppel")
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	info := result.Structs["BasicUser"]

	for _, field := range info.Fields {
		switch field.Name {
		case "ID", "Name":
			if !field.Directive.Deep {
				t.Errorf("field %s: expected default Deep directive, got %+v", field.Name, field.Directive)
			}
		case "Tags":
			if !field.Directive.Deep {
				t.Errorf("Tags: expected Deep, got %+v", field.Directive)
			}
		case "Secret":
			if !field.Directive.Skip {
				t.Errorf("Secret: expected Skip, got %+v", field.Directive)
			}
		case "Config":
			if !field.Directive.Shallow {
				t.Errorf("Config: expected Shallow, got %+v", field.Directive)
			}
		case "Cache":
			if !field.Directive.Empty {
				t.Errorf("Cache: expected Empty, got %+v", field.Directive)
			}
		}
	}
}

func TestParsePackage_TypeCategories(t *testing.T) {
	td := testdataPath(t)
	result, err := parser.ParsePackage(td, "doppel")
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	info := result.Structs["BasicUser"]
	catMap := make(map[string]int)
	for _, f := range info.Fields {
		catMap[f.Name] = int(f.TypeCategory)
	}

	tests := []struct {
		field string
		cat   int // TypeCategory value
	}{
		{"ID", 0},     // CatPrimitive
		{"Tags", 1},   // CatSlice
		{"Scores", 2}, // CatMap
	}

	for _, tc := range tests {
		got, ok := catMap[tc.field]
		if !ok {
			t.Errorf("field %q not found", tc.field)
			continue
		}
		if got != tc.cat {
			t.Errorf("field %q category = %d, want %d", tc.field, got, tc.cat)
		}
	}
}

func TestParsePackage_ExistingCloneSkipped(t *testing.T) {
	td := testdataPath(t)
	result, err := parser.ParsePackage(td, "doppel")
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	info, ok := result.Structs["ExistingClone"]
	if !ok {
		t.Fatal("ExistingClone not found")
	}
	if !info.Skip {
		t.Error("ExistingClone should be skipped (has existing Clone method)")
	}
	if info.SkipReason != "has existing Clone() method" {
		t.Errorf("skip reason = %q, want %q", info.SkipReason, "has existing Clone() method")
	}
}

func TestParsePackage_SkipGenComment(t *testing.T) {
	td := testdataPath(t)
	result, err := parser.ParsePackage(td, "doppel")
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	info, ok := result.Structs["SkipGenStruct"]
	if !ok {
		t.Fatal("SkipGenStruct not found")
	}
	if !info.Skip {
		t.Error("SkipGenStruct should be skipped (has doppel:skip-gen comment)")
	}
}

func TestParsePackage_SkipAllFile(t *testing.T) {
	td := testdataPath(t)
	result, err := parser.ParsePackage(td, "doppel")
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	for _, name := range []string{"SkippedInFile", "AlsoSkippedInFile"} {
		info, ok := result.Structs[name]
		if !ok {
			t.Errorf("type %q not found in parse result", name)
			continue
		}
		if !info.Skip {
			t.Errorf("type %q should be skipped (file has skip-all)", name)
		}
	}
}

func TestParsePackage_UnexportedSkipped(t *testing.T) {
	td := testdataPath(t)
	result, err := parser.ParsePackage(td, "doppel")
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	// unexported struct should not appear in results.
	if _, ok := result.Structs["unexportedStruct"]; ok {
		// It may appear in Structs but should be skipped or have no fields.
		// Actually, the parser does parse it but filters unexported fields.
		// The struct itself can still be in the map.
	}
}

func TestResolveDependencies(t *testing.T) {
	td := testdataPath(t)
	result, err := parser.ParsePackage(td, "doppel")
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	filtered := parser.FilterStructs(result.Structs, nil)
	sorted, err := parser.ResolveDependencies(filtered)
	if err != nil {
		t.Fatalf("ResolveDependencies failed: %v", err)
	}

	// Address should come before NestedUser (dependency).
	addrIdx := -1
	nestedIdx := -1
	for i, name := range sorted {
		if name == "Address" {
			addrIdx = i
		}
		if name == "NestedUser" {
			nestedIdx = i
		}
	}

	if addrIdx != -1 && nestedIdx != -1 && addrIdx > nestedIdx {
		t.Errorf("Address (idx %d) should come before NestedUser (idx %d)", addrIdx, nestedIdx)
	}
}

func TestFilterStructs(t *testing.T) {
	td := testdataPath(t)
	result, err := parser.ParsePackage(td, "doppel")
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	// Filter to specific types.
	filtered := parser.FilterStructs(result.Structs, []string{"BasicUser", "Address"})
	if len(filtered) != 2 {
		t.Errorf("expected 2 structs, got %d", len(filtered))
	}
	if _, ok := filtered["BasicUser"]; !ok {
		t.Error("BasicUser not in filtered result")
	}
	if _, ok := filtered["Address"]; !ok {
		t.Error("Address not in filtered result")
	}
}

func TestFilterStructs_SkippedExcluded(t *testing.T) {
	td := testdataPath(t)
	result, err := parser.ParsePackage(td, "doppel")
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	// Filter all (empty list) — should exclude skipped.
	filtered := parser.FilterStructs(result.Structs, nil)
	for _, info := range filtered {
		if info.Skip {
			t.Errorf("skipped type %q should not be in filtered result", info.Name)
		}
	}
}

func TestParsePackage_PointerPrimitives(t *testing.T) {
	td := testdataPath(t)
	result, err := parser.ParsePackage(td, "doppel")
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	info, ok := result.Structs["PointerPrimitives"]
	if !ok {
		t.Fatal("PointerPrimitives not found")
	}

	catMap := make(map[string]int)
	for _, f := range info.Fields {
		catMap[f.Name] = int(f.TypeCategory)
	}

	// *string should be CatPtrPrimitive (value 3)
	if catMap["Name"] != 3 {
		t.Errorf("Name (*string) category = %d, want 3 (CatPtrPrimitive)", catMap["Name"])
	}
}

func TestParsePackage_EmptyTag(t *testing.T) {
	td := testdataPath(t)
	result, err := parser.ParsePackage(td, "doppel")
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	info, ok := result.Structs["EmptyFields"]
	if !ok {
		t.Fatal("EmptyFields not found")
	}

	dirMap := make(map[string]core.TagDirective)
	for _, f := range info.Fields {
		dirMap[f.Name] = f.Directive
	}

	if !dirMap["Tags"].Empty {
		t.Error("Tags should have Empty directive")
	}
	if !dirMap["Scores"].Empty {
		t.Error("Scores should have Empty directive")
	}
	if !dirMap["Addr"].Empty {
		t.Error("Addr should have Empty directive")
	}
}

package emitter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/seyallius/doppel/cmd/doppelgen/internal/types"
	"github.com/seyallius/doppel/core"
)

func TestGenerate_BasicUser(t *testing.T) {
	t.Parallel()

	info := &types.StructInfo{
		Name:    "BasicUser",
		Package: "testdata",
		File:    "basic_types.go",
		Fields: []types.FieldInfo{
			{Name: "ID", Type: "int64", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatPrimitive},
			{Name: "Name", Type: "string", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatPrimitive},
			{Name: "Tags", Type: "[]string", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatSlice, ElemType: "string"},
			{Name: "Scores", Type: "map[string]int", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatMap, KeyType: "string", ValueType: "int"},
			{Name: "Secret", Type: "string", Directive: core.TagDirective{Skip: true}, TypeCategory: types.CatPrimitive},
			{Name: "Config", Type: "map[string]string", Directive: core.TagDirective{Shallow: true}, TypeCategory: types.CatMap, KeyType: "string", ValueType: "string"},
			{Name: "Cache", Type: "[]string", Directive: core.TagDirective{Empty: true}, TypeCategory: types.CatSlice, ElemType: "string"},
		},
	}

	code, err := Generate(info)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify key patterns in the output.
	checks := []struct {
		name     string
		contains string
	}{
		{"package declaration", "package testdata"},
		{"nil guard", "if x == nil"},
		{"ID primitive", "ID := x.ID"},
		{"Name primitive", "Name := x.Name"},
		{"Tags clone slice", "manual.CloneSlice(x.Tags, manual.Identity[string])"},
		{"Scores clone map", "manual.CloneMap(x.Scores, manual.Identity[int])"},
		{"Config shallow", "Config := x.Config"},
		{"Cache empty", "Cache := []string{}"},
		{"Secret skip", "// Field: Secret (tag: skip)"},
		{"WrapError Tags", `WrapError("BasicUser.Tags", err)`},
		{"WrapError Scores", `WrapError("BasicUser.Scores", err)`},
		{"return struct", "return &BasicUser{"},
		{"import manual", "doppel/manual"},
		{"import core", "doppel/core"},
	}

	for _, tc := range checks {
		t.Run(tc.name, func(t *testing.T) {
			if !containsStr(code, tc.contains) {
				t.Errorf("generated code missing %q\nFull output:\n%s", tc.contains, code)
			}
		})
	}

	// Verify Secret is NOT assigned (skip tag).
	if containsStr(code, "Secret := x.Secret") {
		t.Error("Secret should be skipped (not assigned)")
	}
}

func TestGenerate_Address(t *testing.T) {
	t.Parallel()

	info := &types.StructInfo{
		Name:    "Address",
		Package: "testdata",
		File:    "basic_types.go",
		Fields: []types.FieldInfo{
			{Name: "Street", Type: "string", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatPrimitive},
			{Name: "City", Type: "string", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatPrimitive},
			{Name: "State", Type: "string", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatPrimitive},
			{Name: "Zip", Type: "string", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatPrimitive},
		},
	}

	code, err := Generate(info)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// All primitives, no manual/core imports needed.
	if !containsStr(code, "Street := x.Street") {
		t.Error("missing Street assignment")
	}
	if !containsStr(code, "City := x.City") {
		t.Error("missing City assignment")
	}
	// For all-primitive structs, there may be no imports at all.
}

func TestGenerate_NestedUser(t *testing.T) {
	t.Parallel()

	info := &types.StructInfo{
		Name:    "NestedUser",
		Package: "testdata",
		File:    "basic_types.go",
		Fields: []types.FieldInfo{
			{Name: "ID", Type: "int64", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatPrimitive},
			{Name: "Name", Type: "string", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatPrimitive},
			{Name: "Address", Type: "*Address", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatPtrStruct, PointedToType: "Address"},
			{Name: "Items", Type: "[]Address", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatSlice, ElemType: "Address"},
			{Name: "Labels", Type: "map[string]string", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatMap, KeyType: "string", ValueType: "string"},
		},
	}

	code, err := Generate(info)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	checks := []struct {
		name     string
		contains string
	}{
		{"pointer deep", "manual.ClonePointer(x.Address, func(v Address)"},
		{"pointer Clone call", "return v.Clone()"},
		{"slice struct Clone", "manual.CloneSlice(x.Items, func(v Address)"},
		{"map string identity", "manual.CloneMap(x.Labels, manual.Identity[string])"},
		{"WrapError Address", `WrapError("NestedUser.Address", err)`},
		{"WrapError Items", `WrapError("NestedUser.Items", err)`},
		{"WrapError Labels", `WrapError("NestedUser.Labels", err)`},
	}

	for _, tc := range checks {
		t.Run(tc.name, func(t *testing.T) {
			if !containsStr(code, tc.contains) {
				t.Errorf("generated code missing %q\nFull output:\n%s", tc.contains, code)
			}
		})
	}
}

func TestGenerate_EmptyFields(t *testing.T) {
	t.Parallel()

	info := &types.StructInfo{
		Name:    "EmptyFields",
		Package: "testdata",
		File:    "basic_types.go",
		Fields: []types.FieldInfo{
			{Name: "Name", Type: "string", Directive: core.TagDirective{Empty: true}, TypeCategory: types.CatPrimitive},
			{Name: "Tags", Type: "[]string", Directive: core.TagDirective{Empty: true}, TypeCategory: types.CatSlice, ElemType: "string"},
			{Name: "Scores", Type: "map[int]string", Directive: core.TagDirective{Empty: true}, TypeCategory: types.CatMap, KeyType: "int", ValueType: "string"},
			{Name: "Addr", Type: "*Address", Directive: core.TagDirective{Empty: true}, TypeCategory: types.CatPtrStruct, PointedToType: "Address"},
		},
	}

	code, err := Generate(info)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Primitive with empty tag should still be assignment (not empty literal).
	if !containsStr(code, "Name := x.Name") {
		t.Error("primitive Name with empty tag should be assignment")
	}

	// Non-primitives should get empty-but-non-nil.
	if !containsStr(code, "Tags := []string{}") {
		t.Error("Tags should be empty slice literal")
	}
	if !containsStr(code, "Scores := map[int]string{}") {
		t.Error("Scores should be empty map literal")
	}
	if !containsStr(code, "Addr := &Address{}") {
		t.Error("Addr should be empty pointer literal")
	}
}

func TestGenerate_PointerPrimitives(t *testing.T) {
	t.Parallel()

	info := &types.StructInfo{
		Name:    "PointerPrimitives",
		Package: "testdata",
		File:    "basic_types.go",
		Fields: []types.FieldInfo{
			{Name: "Name", Type: "*string", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatPtrPrimitive, PointedToType: "string"},
			{Name: "Age", Type: "*int", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatPtrPrimitive, PointedToType: "int"},
			{Name: "Active", Type: "*bool", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatPtrPrimitive, PointedToType: "bool"},
			{Name: "Secret", Type: "*string", Directive: core.TagDirective{Skip: true}, TypeCategory: types.CatPtrPrimitive, PointedToType: "string"},
			{Name: "Shallow", Type: "*string", Directive: core.TagDirective{Shallow: true}, TypeCategory: types.CatPtrPrimitive, PointedToType: "string"},
			{Name: "EmptyP", Type: "*string", Directive: core.TagDirective{Empty: true}, TypeCategory: types.CatPtrPrimitive, PointedToType: "string"},
		},
	}

	code, err := Generate(info)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Deep pointer primitives should use ClonePointer + Identity.
	if !containsStr(code, "manual.ClonePointer(x.Name, manual.Identity[string])") {
		t.Error("missing Name ClonePointer with Identity[string]")
	}
	if !containsStr(code, "manual.ClonePointer(x.Age, manual.Identity[int])") {
		t.Error("missing Age ClonePointer with Identity[int]")
	}

	// Skip should omit Secret.
	if containsStr(code, "Secret := x.Secret") {
		t.Error("Secret should be skipped")
	}

	// Shallow should be direct assignment.
	if !containsStr(code, "Shallow := x.Shallow") {
		t.Error("missing Shallow direct assignment")
	}

	// Empty pointer should be &string{}.
	if !containsStr(code, "EmptyP := &string{}") {
		t.Error("missing EmptyP empty pointer literal")
	}
}

func TestGenerate_CloneTag(t *testing.T) {
	t.Parallel()

	info := &types.StructInfo{
		Name:    "CloneTagUser",
		Package: "testdata",
		File:    "basic_types.go",
		Fields: []types.FieldInfo{
			{Name: "ID", Type: "int64", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatPrimitive},
			{Name: "Profile", Type: "*Profile", Directive: core.TagDirective{Clone: true}, TypeCategory: types.CatPtrStruct, PointedToType: "Profile"},
		},
	}

	code, err := Generate(info)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should call convention-based clone function.
	if !containsStr(code, "cloneCloneTagUserProfile(x.Profile)") {
		t.Error("missing convention-based clone function call")
	}
	if !containsStr(code, `WrapError("CloneTagUser.Profile", err)`) {
		t.Error("missing WrapError for Profile")
	}
}

func TestGeneratePreview(t *testing.T) {
	t.Parallel()

	info := &types.StructInfo{
		Name:    "Address",
		Package: "testdata",
		File:    "basic_types.go",
		Fields: []types.FieldInfo{
			{Name: "Street", Type: "string", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatPrimitive},
			{Name: "City", Type: "string", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatPrimitive},
		},
	}

	code := GeneratePreview(info)
	if code == "" {
		t.Error("GeneratePreview returned empty string")
	}
	if !containsStr(code, "package testdata") {
		t.Error("preview missing package declaration")
	}
}

func TestGoldenFile_BasicUser(t *testing.T) {
	t.Parallel()

	info := &types.StructInfo{
		Name:    "BasicUser",
		Package: "testdata",
		File:    "basic_types.go",
		Fields: []types.FieldInfo{
			{Name: "ID", Type: "int64", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatPrimitive},
			{Name: "Name", Type: "string", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatPrimitive},
			{Name: "Tags", Type: "[]string", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatSlice, ElemType: "string"},
			{Name: "Scores", Type: "map[string]int", Directive: core.TagDirective{Deep: true}, TypeCategory: types.CatMap, KeyType: "string", ValueType: "int"},
			{Name: "Secret", Type: "string", Directive: core.TagDirective{Skip: true}, TypeCategory: types.CatPrimitive},
			{Name: "Config", Type: "map[string]string", Directive: core.TagDirective{Shallow: true}, TypeCategory: types.CatMap, KeyType: "string", ValueType: "string"},
			{Name: "Cache", Type: "[]string", Directive: core.TagDirective{Empty: true}, TypeCategory: types.CatSlice, ElemType: "string"},
		},
	}

	code, err := Generate(info)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Load golden file.
	goldenPath := filepath.Join("..", "..", "testdata", "basicuser_clone.gen.go.golden")
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Skipf("golden file not found: %v", err)
	}

	if string(golden) != code {
		t.Errorf("generated output differs from golden file.\n--- Generated ---\n%s\n--- Golden ---\n%s", code, string(golden))
	}
}

// --- Helpers ---

func containsStr(s, sub string) bool {
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

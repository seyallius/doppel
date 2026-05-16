package emitter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/seyallius/doppel/cmd/doppelgen/internal/emitter"
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

	code, err := emitter.Generate(info)
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
		// Variable names are now prefixed with "cloned" to avoid field/type name conflicts.
		{"ID primitive", "clonedID := x.ID"},
		{"Name primitive", "clonedName := x.Name"},
		{"Tags clone slice", "manual.CloneSlice(x.Tags, manual.Identity[string])"},
		{"Scores clone map", "manual.CloneMap(x.Scores, manual.Identity[int])"},
		{"Config shallow", "clonedConfig := x.Config"},
		{"Cache empty", "clonedCache := []string{}"},
		{"Secret skip", "// Field: Secret (tag: skip)"},
		{"WrapError Tags", `WrapError("BasicUser.Tags", err)`},
		{"WrapError Scores", `WrapError("BasicUser.Scores", err)`},
		{"return struct", "return &BasicUser{"},
		// Return statement must map field names → cloned variable names.
		{"return ID", "ID: clonedID,"},
		{"return Name", "Name: clonedName,"},
		{"return Tags", "Tags: clonedTags,"},
		{"return Scores", "Scores: clonedScores,"},
		{"return Config", "Config: clonedConfig,"},
		{"return Cache", "Cache: clonedCache,"},
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
	if containsStr(code, "Secret := x.Secret") || containsStr(code, "clonedSecret := x.Secret") {
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

	code, err := emitter.Generate(info)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// All primitives use cloned-prefixed variables.
	if !containsStr(code, "clonedStreet := x.Street") {
		t.Error("missing clonedStreet assignment")
	}
	if !containsStr(code, "clonedCity := x.City") {
		t.Error("missing clonedCity assignment")
	}
	// Return must use cloned vars.
	if !containsStr(code, "Street: clonedStreet,") {
		t.Error("return missing Street: clonedStreet")
	}
	if !containsStr(code, "City: clonedCity,") {
		t.Error("return missing City: clonedCity")
	}
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

	code, err := emitter.Generate(info)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	checks := []struct {
		name     string
		contains string
	}{
		// CatPtrStruct: ClonePointer with a closure that dereferences the *Address return.
		{"pointer deep ClonePointer", "manual.ClonePointer(x.Address, func(v Address) (Address, error)"},
		{"pointer closure dereference", "return *cloned, nil"},
		{"pointer nil guard in closure", "if cloned == nil {"},
		// CatSlice[Address]: same closure pattern.
		{"slice struct CloneSlice", "manual.CloneSlice(x.Items, func(v Address) (Address, error)"},
		// CatMap[string]string: primitive value, uses Identity.
		{"map string identity", "manual.CloneMap(x.Labels, manual.Identity[string])"},
		{"WrapError Address", `WrapError("NestedUser.Address", err)`},
		{"WrapError Items", `WrapError("NestedUser.Items", err)`},
		{"WrapError Labels", `WrapError("NestedUser.Labels", err)`},
		// Return uses cloned-prefixed variables — no field/type name collision.
		{"return Address", "Address: clonedAddress,"},
		{"return Items", "Items: clonedItems,"},
		{"return Labels", "Labels: clonedLabels,"},
	}

	for _, tc := range checks {
		t.Run(tc.name, func(t *testing.T) {
			if !containsStr(code, tc.contains) {
				t.Errorf("generated code missing %q\nFull output:\n%s", tc.contains, code)
			}
		})
	}

	// Verify there is no bare "Address :=" (the old name-collision pattern).
	if containsStr(code, "\nAddress, err :=") || containsStr(code, "\tAddress, err :=") {
		t.Error("variable named 'Address' conflicts with struct type; should use 'clonedAddress'")
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

	code, err := emitter.Generate(info)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Primitive with empty tag should still be assignment.
	if !containsStr(code, "clonedName := x.Name") {
		t.Error("primitive Name with empty tag should be direct assignment to clonedName")
	}

	// Non-primitives should get empty-but-non-nil using cloned-prefixed vars.
	if !containsStr(code, "clonedTags := []string{}") {
		t.Error("Tags should be empty slice literal assigned to clonedTags")
	}
	if !containsStr(code, "clonedScores := map[int]string{}") {
		t.Error("Scores should be empty map literal assigned to clonedScores")
	}
	if !containsStr(code, "clonedAddr := &Address{}") {
		t.Error("Addr should be empty pointer literal assigned to clonedAddr")
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

	code, err := emitter.Generate(info)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Deep pointer primitives should use ClonePointer + Identity with cloned-prefixed vars.
	if !containsStr(code, "manual.ClonePointer(x.Name, manual.Identity[string])") {
		t.Error("missing clonedName ClonePointer with Identity[string]")
	}
	if !containsStr(code, "manual.ClonePointer(x.Age, manual.Identity[int])") {
		t.Error("missing clonedAge ClonePointer with Identity[int]")
	}

	// Skip should omit Secret entirely.
	if containsStr(code, "clonedSecret") {
		t.Error("Secret should be skipped (not assigned)")
	}

	// Shallow should be direct assignment.
	if !containsStr(code, "clonedShallow := x.Shallow") {
		t.Error("missing clonedShallow direct assignment")
	}

	// Empty pointer-to-primitive is treated as primitive category — direct assignment.
	if !containsStr(code, "clonedEmptyP := x.EmptyP") {
		t.Error("missing clonedEmptyP direct assignment (pointer-to-primitive with empty tag)")
	}

	// Return should map correctly.
	if !containsStr(code, "Name: clonedName,") {
		t.Error("return missing Name: clonedName")
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

	code, err := emitter.Generate(info)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Convention-based clone function call with cloned-prefixed variable.
	if !containsStr(code, "CloneCloneTagUserProfile(x.Profile)") {
		t.Error("missing convention-based clone function call")
	}
	if !containsStr(code, `WrapError("CloneTagUser.Profile", err)`) {
		t.Error("missing WrapError for Profile")
	}
	if !containsStr(code, "Profile: clonedProfile,") {
		t.Error("return missing Profile: clonedProfile")
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

	code := emitter.GeneratePreview(info)
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

	code, err := emitter.Generate(info)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Load golden file.
	goldenPath := filepath.Join("..", "..", "testdata", "basicuser.clone_gen.go.golden")
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

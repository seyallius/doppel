// Package parser_test. recursive_test.go - Tests for cross-package dependency resolution,
// third-party detection, module utilities, and the ParseProject orchestrator.
package parser_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seyallius/doppel/cmd/doppelgen/internal/emitter"
	"github.com/seyallius/doppel/cmd/doppelgen/internal/parser"
	"github.com/seyallius/doppel/cmd/doppelgen/internal/types"
	"github.com/seyallius/doppel/core"
)

// -------------------------------------------- Module Utilities --------------------------------------------

func TestFindModuleRoot_FindsGoMod(t *testing.T) {
	t.Parallel()

	// Walk up from a nested testdata directory — should find the project root go.mod.
	td := testdataPath(t)
	root, err := parser.FindModuleRoot(td)
	if err != nil {
		t.Fatalf("FindModuleRoot(%q) failed: %v", td, err)
	}

	// The root must contain a go.mod.
	if _, statErr := os.Stat(filepath.Join(root, "go.mod")); statErr != nil {
		t.Errorf("FindModuleRoot returned %q which has no go.mod", root)
	}
}

func TestFindModuleRoot_NoGoMod(t *testing.T) {
	t.Parallel()

	// Use a temp directory guaranteed to have no go.mod.
	tmp := t.TempDir()
	_, err := parser.FindModuleRoot(tmp)
	if err == nil {
		t.Error("expected error when no go.mod exists, got nil")
	}
}

func TestParseModulePath_ExtractsPath(t *testing.T) {
	t.Parallel()

	// Find the real module root for this project.
	td := testdataPath(t)
	root, err := parser.FindModuleRoot(td)
	if err != nil {
		t.Skipf("module root not found: %v", err)
	}

	modulePath, err := parser.ParseModulePath(root)
	if err != nil {
		t.Fatalf("ParseModulePath failed: %v", err)
	}

	if modulePath == "" {
		t.Error("ParseModulePath returned empty string")
	}

	// The module path should start with a recognizable prefix.
	if len(modulePath) < 3 {
		t.Errorf("module path looks suspiciously short: %q", modulePath)
	}
}

func TestParseModulePath_MissingGoMod(t *testing.T) {
	t.Parallel()

	_, err := parser.ParseModulePath(t.TempDir())
	if err == nil {
		t.Error("expected error for directory without go.mod, got nil")
	}
}

// -------------------------------------------- TypeResolver --------------------------------------------

func TestTypeResolver_IsProjectInternal_True(t *testing.T) {
	t.Parallel()

	root, path := moduleRootAndPath(t)
	resolver := parser.NewTypeResolver(path, root)

	// The main package itself should be internal.
	if !resolver.IsProjectInternal(path) {
		t.Errorf("IsProjectInternal(%q) = false, want true for module root import path", path)
	}

	// A sub-package should also be internal.
	subPkg := path + "/manual"
	if !resolver.IsProjectInternal(subPkg) {
		t.Errorf("IsProjectInternal(%q) = false, want true", subPkg)
	}
}

func TestTypeResolver_IsProjectInternal_False(t *testing.T) {
	t.Parallel()

	root, path := moduleRootAndPath(t)
	resolver := parser.NewTypeResolver(path, root)

	thirdParty := []string{
		"github.com/spf13/cobra",
		"golang.org/x/tools/imports",
		"github.com/completely/unrelated/pkg",
	}

	for _, tp := range thirdParty {
		if resolver.IsProjectInternal(tp) {
			t.Errorf("IsProjectInternal(%q) = true, want false (should be third-party)", tp)
		}
	}
}

func TestTypeResolver_IsProjectInternal_PrefixCollision(t *testing.T) {
	t.Parallel()

	// "github.com/foo/bar" must NOT match "github.com/foo/bar-extended".
	resolver := parser.NewTypeResolver("github.com/foo/bar", t.TempDir())

	if resolver.IsProjectInternal("github.com/foo/bar-extended") {
		t.Error("IsProjectInternal matched an extended prefix — prefix collision not guarded")
	}
}

func TestTypeResolver_GetOrParse_Caches(t *testing.T) {
	t.Parallel()

	root, path := moduleRootAndPath(t)
	resolver := parser.NewTypeResolver(path, root)

	// Parse core package (known to exist in the project).
	corePath := path + "/core"
	r1, err := resolver.GetOrParse(corePath, "doppel")
	if err != nil {
		t.Skipf("could not parse core package: %v", err)
	}

	// Second call must return the exact same pointer (cached).
	r2, err := resolver.GetOrParse(corePath, "doppel")
	if err != nil {
		t.Fatalf("second GetOrParse failed: %v", err)
	}

	if r1 != r2 {
		t.Error("GetOrParse did not cache result — second call returned different pointer")
	}
}

// -------------------------------------------- Cross-Package Field Detection --------------------------------------------

func TestParsePackage_CrossPackageField_NotUnknown(t *testing.T) {
	t.Parallel()

	// The testdata package has a struct with a cross-package field.
	// After the parser fix, its TypeCategory must NOT be CatUnknown.
	td := testdataPath(t)
	root, modulePath := moduleRootAndPath(t)
	resolver := parser.NewTypeResolver(modulePath, root)

	result, err := parser.ParsePackageWithResolver(td, "doppel", resolver)
	if err != nil {
		t.Fatalf("ParsePackageWithResolver failed: %v", err)
	}

	// CrossPkgUser is defined in testdata with a cross-package field.
	info, ok := result.Structs["CrossPkgUser"]
	if !ok {
		t.Skip("CrossPkgUser not found in testdata — add it to run this test")
	}

	for _, field := range info.Fields {
		if field.TypeCategory == types.CatUnknown {
			t.Errorf("field %q has CatUnknown after resolver-aware parse", field.Name)
		}
	}
}

func TestParsePackage_ThirdPartyField_Detected(t *testing.T) {
	t.Parallel()

	// ThirdPartyUser has a field of type cobra.Command (third-party).
	td := testdataPath(t)
	root, modulePath := moduleRootAndPath(t)
	resolver := parser.NewTypeResolver(modulePath, root)

	result, err := parser.ParsePackageWithResolver(td, "doppel", resolver)
	if err != nil {
		t.Fatalf("ParsePackageWithResolver failed: %v", err)
	}

	info, ok := result.Structs["ThirdPartyUser"]
	if !ok {
		t.Skip("ThirdPartyUser not found in testdata — add it to run this test")
	}

	for _, field := range info.Fields {
		if field.Type == "cobra.Command" || field.Type == "*cobra.Command" {
			if !field.IsThirdParty {
				t.Errorf("field %q (%s) should be IsThirdParty=true", field.Name, field.Type)
			}
			if field.TypeCategory != types.CatThirdPartyStruct &&
				field.TypeCategory != types.CatThirdPartyPtrStruct {
				t.Errorf("field %q TypeCategory = %v, want CatThirdPartyStruct/PtrStruct", field.Name, field.TypeCategory)
			}
		}
	}
}

// -------------------------------------------- ParseProject (Orchestrator) --------------------------------------------

func TestParseProject_DiscoversInternalDeps(t *testing.T) {
	t.Parallel()

	td := testdataPath(t)
	result, err := parser.ParseProject(td, "doppel", "")
	if err != nil {
		t.Fatalf("ParseProject failed: %v", err)
	}

	if len(result.Packages) < 1 {
		t.Error("ParseProject returned 0 packages")
	}

	if len(result.Structs) < 1 {
		t.Error("ParseProject returned 0 structs")
	}
}

func TestParseProject_TopologicalOrderValid(t *testing.T) {
	t.Parallel()

	td := testdataPath(t)
	result, err := parser.ParseProject(td, "doppel", "")
	if err != nil {
		t.Fatalf("ParseProject failed: %v", err)
	}

	// Every key in TopologicalOrder must exist in Structs.
	for _, key := range result.TopologicalOrder {
		if _, ok := result.Structs[key]; !ok {
			t.Errorf("TopologicalOrder contains %q which is not in Structs", key)
		}
	}
}

func TestParseProject_ModuleRootOverride(t *testing.T) {
	t.Parallel()

	td := testdataPath(t)
	root, _ := moduleRootAndPath(t)

	// Override with the correct module root — should succeed.
	result, err := parser.ParseProject(td, "doppel", root)
	if err != nil {
		t.Fatalf("ParseProject with override failed: %v", err)
	}

	if len(result.Structs) == 0 {
		t.Error("ParseProject with override returned 0 structs")
	}
}

func TestParseProject_FallbackOnMissingGoMod(t *testing.T) {
	t.Parallel()

	// Copy testdata to a temp dir that has no go.mod — ParseProject should fall back
	// to single-package mode gracefully rather than returning an error.
	tmp := t.TempDir()

	// Write a minimal Go file so the parser finds something.
	goSrc := `package testfallback
type Simple struct {
	Name string
}
`
	if err := os.WriteFile(filepath.Join(tmp, "simple.go"), []byte(goSrc), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	result, err := parser.ParseProject(tmp, "doppel", "")
	if err != nil {
		t.Fatalf("ParseProject fallback failed: %v", err)
	}

	if result == nil {
		t.Error("ParseProject fallback returned nil result")
	}
}

// -------------------------------------------- Emitter: Third-Party Convention Stubs --------------------------------------------

func TestGenerate_ThirdPartyField_EmitsConventionStub(t *testing.T) {
	t.Parallel()

	info := &types.StructInfo{
		Name:    "Session",
		Package: "testdata",
		File:    "session.go",
		Fields: []types.FieldInfo{
			{
				Name:         "ID",
				Type:         "int64",
				Directive:    core.TagDirective{Deep: true},
				TypeCategory: types.CatPrimitive,
			},
			{
				Name:         "Credential",
				Type:         "webauthn.Credential",
				Directive:    core.TagDirective{Deep: true},
				TypeCategory: types.CatThirdPartyStruct,
				IsThirdParty: true,
				ImportPath:   "github.com/go-webauthn/webauthn/protocol",
			},
		},
	}

	code, err := emitter.Generate(info)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	checks := []string{
		"CloneSessionCredential(x.Credential)", // convention call
		"//todo(CloneSessionCredential):",      // todo comment
		`WrapError("Session.Credential", err)`, // error wrapping
		"clonedCredential",                     // cloned-prefixed var
	}
	for _, want := range checks {
		if !strings.Contains(code, want) {
			t.Errorf("third-party field: generated code missing %q\nOutput:\n%s", want, code)
		}
	}

	// Must NOT import the third-party package in the generated file.
	if strings.Contains(code, `"github.com/go-webauthn/webauthn/protocol"`) {
		t.Error("generated file should not import third-party package — convention function is user-provided")
	}
}

func TestGenerate_ThirdPartySliceElem_EmitsConventionStub(t *testing.T) {
	t.Parallel()

	info := &types.StructInfo{
		Name:    "Batch",
		Package: "testdata",
		File:    "batch.go",
		Fields: []types.FieldInfo{
			{
				Name:             "Items",
				Type:             "[]extpkg.Item",
				Directive:        core.TagDirective{Deep: true},
				TypeCategory:     types.CatSlice,
				ElemType:         "extpkg.Item",
				ElemImportPath:   "github.com/external/extpkg",
				ElemIsThirdParty: true,
			},
		},
	}

	code, err := emitter.Generate(info)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	checks := []string{
		"CloneBatchItemsElem",                             // elem convention function name
		"manual.CloneSlice(x.Items, CloneBatchItemsElem)", // CloneSlice wired to convention fn
		"//todo(CloneBatchItemsElem):",                    // todo comment
	}
	for _, want := range checks {
		if !strings.Contains(code, want) {
			t.Errorf("third-party slice elem: missing %q\nOutput:\n%s", want, code)
		}
	}
}

// -------------------------------------------- Helpers --------------------------------------------

// moduleRootAndPath returns the detected module root directory and module path
// for the current project. Skips the test if detection fails.
func moduleRootAndPath(t *testing.T) (root, modulePath string) {
	t.Helper()

	td := testdataPath(t)
	var err error
	root, err = parser.FindModuleRoot(td)
	if err != nil {
		t.Skipf("module root not found: %v", err)
	}

	modulePath, err = parser.ParseModulePath(root)
	if err != nil {
		t.Skipf("module path not found: %v", err)
	}

	return root, modulePath
}

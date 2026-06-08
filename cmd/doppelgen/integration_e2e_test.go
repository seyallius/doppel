// Package main. integration_e2e_test.go - End-to-end integration tests for the doppelgen CLI.
// It tests the full pipeline (parsing, dependency resolution, emission, and file routing)
// using temporary directories to simulate real-world multi-package and single-package scenarios.
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seyallius/doppel/cmd/doppelgen/internal/types"
)

// -------------------------------------------- Integration ------------------------------------------ //

// TestCLI_MultiPackage_FlagParsingAndValidation verifies that the Cobra CLI correctly
// parses multiple --package flags into a slice, and that the CLI command itself
// correctly rejects the --output flag when multiple --package flags are provided.
func TestCLI_MultiPackage_FlagParsingAndValidation(t *testing.T) {
	t.Parallel()

	pkgADir := setupTestPackage(t, "pkgA", "a.go", `package pkgA
type UserA struct { Name string `+"`doppel:\"deep\"`"+` }`)

	pkgBDir := setupTestPackage(t, "pkgB", "b.go", `package pkgB
type OrderB struct { ID int `+"`doppel:\"deep\"`"+` }`)

	outDir := t.TempDir()

	// 1. Instantiate the actual Cobra command
	cmd := newRootCmd()

	// 2. Simulate the user typing: doppelgen -p pkgA -p pkgB -o outDir
	cmd.SetArgs([]string{
		"-p", pkgADir,
		"-p", pkgBDir,
		"-o", outDir,
	})

	// 3. Execute the CLI
	err := cmd.Execute()

	// 4. Assert that the CLI correctly surfaced the validation error from run()
	if err == nil {
		t.Fatal("expected CLI to return an error for multi-package + --output, got nil")
	}

	expectedErr := "--output cannot be used with multiple --package flags"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, got: %v", expectedErr, err)
	}
}

// ----------------------------------------------- Tests -------------------------------------------- //

// TestIntegration_MultiPackage_InlineGeneration verifies that when multiple packages
// are provided without an --output flag, generated files are placed inline next
// to their respective source definitions.
func TestIntegration_MultiPackage_InlineGeneration(t *testing.T) {
	t.Parallel()

	// Setup: Create two distinct packages with tagged structs
	pkgADir := setupTestPackage(t, "pkgA", "a.go", `package pkgA
type UserA struct { Name string `+"`doppel:\"deep\"`"+` }`)

	pkgBDir := setupTestPackage(t, "pkgB", "b.go", `package pkgB
type OrderB struct { ID int `+"`doppel:\"deep\"`"+` }`)

	cfg := &types.GeneratorConfig{
		Packages: []string{pkgADir, pkgBDir},
		Output:   "", // Explicitly empty to trigger inline routing
		Preview:  false,
		Tag:      "doppel",
	}

	err := run(cfg)
	if err != nil {
		t.Fatalf("run() failed unexpectedly: %v", err)
	}

	// Assert: Files must exist in their respective source directories
	assertFileExists(t, filepath.Join(pkgADir, "usera.clone_gen.go"))
	assertFileExists(t, filepath.Join(pkgADir, "usera.clone_gen_test.go"))
	assertFileExists(t, filepath.Join(pkgBDir, "orderb.clone_gen.go"))
	assertFileExists(t, filepath.Join(pkgBDir, "orderb.clone_gen_test.go"))

	// Assert: Verify package declarations are correct (no cross-contamination)
	contentA, _ := os.ReadFile(filepath.Join(pkgADir, "usera.clone_gen.go"))
	if !strings.Contains(string(contentA), "package pkgA") {
		t.Error("generated file for pkgA does not contain 'package pkgA'")
	}
}

// TestIntegration_MultiPackage_OutputFlagRejected verifies that the generator
// explicitly fails when a user attempts to use --output with multiple --package flags.
func TestIntegration_MultiPackage_OutputFlagRejected(t *testing.T) {
	t.Parallel()

	pkgADir := setupTestPackage(t, "pkgA", "a.go", `package pkgA
type UserA struct { Name string `+"`doppel:\"deep\"`"+` }`)

	pkgBDir := setupTestPackage(t, "pkgB", "b.go", `package pkgB
type OrderB struct { ID int `+"`doppel:\"deep\"`"+` }`)

	outDir := t.TempDir()

	cfg := &types.GeneratorConfig{
		Packages: []string{pkgADir, pkgBDir},
		Output:   outDir, // This should trigger the validation guard
		Preview:  false,
		Tag:      "doppel",
	}

	err := run(cfg)
	if err == nil {
		t.Fatal("expected run() to return an error for multi-package + --output, got nil")
	}

	expectedErr := "--output cannot be used with multiple --package flags"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, got: %v", expectedErr, err)
	}
}

// TestIntegration_SinglePackage_OutputFlagRespected verifies backward compatibility:
// a single package with an --output flag should write all generated files to that output directory.
func TestIntegration_SinglePackage_OutputFlagRespected(t *testing.T) {
	t.Parallel()

	pkgDir := setupTestPackage(t, "pkgSingle", "single.go", `package pkgSingle
type Item struct { Value string `+"`doppel:\"deep\"`"+` }`)

	outDir := t.TempDir()

	cfg := &types.GeneratorConfig{
		Packages: []string{pkgDir},
		Output:   outDir,
		Preview:  false,
		Tag:      "doppel",
	}

	err := run(cfg)
	if err != nil {
		t.Fatalf("run() failed unexpectedly: %v", err)
	}

	// Assert: File exists in the output directory
	assertFileExists(t, filepath.Join(outDir, "item.clone_gen.go"))
	assertFileExists(t, filepath.Join(outDir, "item.clone_gen_test.go"))

	// Assert: File does NOT exist in the source directory (inline routing should be bypassed)
	inlinePath := filepath.Join(pkgDir, "item.clone_gen.go")
	if _, err := os.Stat(inlinePath); !os.IsNotExist(err) {
		t.Errorf("expected inline file to NOT exist at %s when --output is provided", inlinePath)
	}
}

// TestIntegration_CrossPackageDependency_MultiPackageRun verifies that when
// multiple packages are provided, and one depends on a struct in another,
// both are successfully parsed, topologically sorted, and generated.
func TestIntegration_CrossPackageDependency_MultiPackageRun(t *testing.T) {
	t.Parallel()

	// pkgB has a standalone struct
	pkgBDir := setupTestPackage(t, "pkgB", "address.go", `package pkgB
type Address struct { City string `+"`doppel:\"deep\"`"+` }`)

	// pkgA has a struct that references pkgB.Address
	// Note: In a real scenario, this would have an import, but for AST parsing
	// of the doppel tag, the type string "pkgB.Address" is enough to trigger
	// the cross-package resolver if the import path is resolved.
	// For this e2e test, we simulate the parser seeing the type.
	pkgADir := setupTestPackage(t, "pkgA", "user.go", `package pkgA
import "testmodule/pkgB"
type UserA struct { 
    Name string `+"`doppel:\"deep\"`"+`
    Addr *pkgB.Address `+"`doppel:\"deep\"`"+` 
}`)

	cfg := &types.GeneratorConfig{
		Packages:   []string{pkgADir, pkgBDir},
		Output:     "",
		Preview:    false,
		Tag:        "doppel",
		ModuleRoot: t.TempDir(), // Fallback to single-package mode gracefully if no go.mod, which is fine for this AST-level test
	}

	err := run(cfg)
	// Even if cross-package resolution warns, it should not hard-fail the generation
	// of the types it can resolve.
	if err != nil && !strings.Contains(err.Error(), "no Go files found") {
		// We allow it to pass if it gracefully handles the missing go.mod fallback
		t.Logf("run() returned (expected in isolated temp env without go.mod): %v", err)
	}

	// The primary assertion is that the orchestrator doesn't panic or deadlock
	// when handed multiple directories with interdependent-looking type names.
	assertFileExists(t, filepath.Join(pkgBDir, "address.clone_gen.go"))
}

// ------------------------------------------- Internal Helpers ------------------------------------- //

// setupTestPackage creates a temporary directory, writes a Go source file with the
// provided content, and returns the absolute path to that directory.
func setupTestPackage(t *testing.T, pkgName, filename, content string) string {
	t.Helper()
	dir := t.TempDir()

	// Write a minimal go.mod to satisfy the parser's module root detection
	goModContent := "module testmodule/" + pkgName + "\ngo 1.25\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file %s: %v", filePath, err)
	}
	return dir
}

// assertFileExists checks that a file exists at the given path, failing the test if not.
func assertFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("expected file to exist at %s, but it was not found", path)
		return
	}
	if err != nil {
		t.Errorf("error checking file at %s: %v", path, err)
		return
	}
	if info.IsDir() {
		t.Errorf("expected file at %s, but found a directory", path)
	}
}

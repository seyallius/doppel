// Package parser. module.go - provides utilities for Go module detection and parsing.
// It walks the filesystem upward to locate the nearest go.mod file and extracts
// the declared module path — enabling the resolver to distinguish project-internal
// types from third-party dependencies at code-generation time.
package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// -------------------------------------------- Public Functions --------------------------------------------

// FindModuleRoot walks up the filesystem starting from startDir until it finds a
// directory containing a go.mod file. Returns the absolute path of that directory.
//
// Returns an error if no go.mod is found before reaching the filesystem root.
// This mirrors the behavior of the `go` tool itself.
//
// Example:
//
//	root, err := FindModuleRoot("/home/user/project/cmd/doppelgen")
//	// root == "/home/user/project"
func FindModuleRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path of %q: %w", startDir, err)
	}

	for {
		candidate := filepath.Join(dir, "go.mod")
		if _, statErr := os.Stat(candidate); statErr == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding go.mod.
			return "", fmt.Errorf("go.mod not found searching upward from %q", startDir)
		}

		dir = parent
	}
}

// ParseModulePath reads the go.mod located in moduleRoot and extracts the module
// path declared on the "module" directive line.
//
// Example (given go.mod: "module github.com/seyallius/doppel"):
//
//	path, err := ParseModulePath("/home/user/project")
//	// path == "github.com/seyallius/doppel"
func ParseModulePath(moduleRoot string) (string, error) {
	gomodPath := filepath.Join(moduleRoot, "go.mod")

	f, err := os.Open(gomodPath)
	if err != nil {
		return "", fmt.Errorf("open go.mod at %q: %w", gomodPath, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// The module directive is always "module <path>" on its own line.
		if strings.HasPrefix(line, "module ") {
			path := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			if path == "" {
				return "", fmt.Errorf("empty module path in %q", gomodPath)
			}
			return path, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan go.mod at %q: %w", gomodPath, err)
	}

	return "", fmt.Errorf("module directive not found in %q", gomodPath)
}

// -------------------------------------------- Internal Helpers --------------------------------------------

// importPathToLocalDir maps a full Go import path to a local filesystem directory
// by stripping the module prefix and joining the remainder with the module root.
//
// Example:
//
//	dir := importPathToLocalDir("github.com/seyallius/doppel/manual",
//	                            "github.com/seyallius/doppel",
//	                            "/home/user/project")
//	// dir == "/home/user/project/manual"
func importPathToLocalDir(importPath, modulePath, moduleRoot string) (string, error) {
	if !strings.HasPrefix(importPath, modulePath) {
		return "", fmt.Errorf("import path %q is not under module %q", importPath, modulePath)
	}

	rel := strings.TrimPrefix(importPath, modulePath)
	rel = strings.TrimPrefix(rel, "/")

	return filepath.Join(moduleRoot, filepath.FromSlash(rel)), nil
}

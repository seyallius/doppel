// Package parser. recursive.go - implements ParseProject, the cross-package parse
// orchestrator for doppelgen. It discovers all project-internal types referenced
// by structs in the target package, parses their packages transitively, builds a
// unified cross-package dependency graph, and returns a topologically-sorted
// generation plan.
//
// Algorithm (fixed-point BFS):
//
//  1. Parse the initial target package.
//  2. Inspect every FieldInfo for non-primitive types with a non-empty ImportPath
//     that is project-internal (IsThirdParty == false).
//  3. For each such ImportPath not yet in the parsed set, enqueue it.
//  4. Parse the enqueued package (via TypeResolver cache), merge its structs.
//  5. Repeat until the queue is empty (fixed point).
//  6. Topologically sort the merged struct set across all packages.
package parser

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/seyallius/doppel/cmd/doppelgen/internal/types"
)

// -------------------------------------------- Types --------------------------------------------

// ProjectParseResult holds the aggregate output of a recursive parse run across
// all project-internal packages reachable from the initial target directory.
type ProjectParseResult struct {
	// Structs is the union of all parsed structs across all packages, keyed by
	// "<packageName>.<TypeName>" to avoid collisions across packages.
	Structs types.TypeInfo

	// TopologicalOrder is the dependency-resolved list of struct keys in the order
	// they must be generated (dependencies before dependents).
	TopologicalOrder []string

	// Packages is a map of import path → ParseResult for every package parsed.
	Packages map[string]*ParseResult

	// InitialPackage is the import path of the target package passed to ParseProject.
	InitialPackage string

	// Skipped holds all types that were explicitly opted out of generation.
	Skipped []SkipInfo
}

// -------------------------------------------- Public Functions --------------------------------------------

// ParseProject is the top-level recursive parse orchestrator. It:
//  1. Detects the Go module root and module path (or uses overrides from cfg).
//  2. Creates a TypeResolver.
//  3. Parses the initial package.
//  4. Transitively discovers and parses project-internal dependencies.
//  5. Returns a ProjectParseResult ready for multi-package code generation.
//
// If module detection fails (e.g. the directory is outside any module), ParseProject
// falls back to single-package mode (no recursive discovery, no third-party detection).
func ParseProject(targetDir, tagKey string, moduleRootOverride string) (*ProjectParseResult, error) {
	// ── Step 1: Module detection ──────────────────────────────────────────
	moduleRoot := moduleRootOverride
	if moduleRoot == "" {
		var err error
		moduleRoot, err = FindModuleRoot(targetDir)
		if err != nil {
			// Graceful fallback: single-package mode.
			return parseProjectFallback(targetDir, tagKey)
		}
	}

	modulePath, err := ParseModulePath(moduleRoot)
	if err != nil {
		return parseProjectFallback(targetDir, tagKey)
	}

	resolver := NewTypeResolver(modulePath, moduleRoot)

	// Resolve the target package's import path (needed as a cache key).
	// We use the parsed package's directory relative to the module root.
	initialImportPath, err := dirToImportPath(targetDir, moduleRoot, modulePath)
	if err != nil {
		return parseProjectFallback(targetDir, tagKey)
	}

	// ── Step 2: Parse initial package ────────────────────────────────────
	initialResult, err := ParsePackageWithResolver(targetDir, tagKey, resolver)
	if err != nil {
		return nil, fmt.Errorf("parse initial package at %q: %w", targetDir, err)
	}

	// Seed the resolver cache with the initial parse result so it isn't re-parsed.
	resolver.mu.Lock()
	resolver.cache[initialImportPath] = initialResult
	resolver.mu.Unlock()

	// ── Step 3: BFS over project-internal dependencies ────────────────────
	parsed := map[string]*ParseResult{
		initialImportPath: initialResult,
	}
	queue := collectInternalImportPaths(initialResult, resolver)
	visited := map[string]bool{initialImportPath: true}

	for len(queue) > 0 {
		importPath := queue[0]
		queue = queue[1:]

		if visited[importPath] {
			continue
		}
		visited[importPath] = true

		depResult, depErr := resolver.GetOrParse(importPath, tagKey)
		if depErr != nil {
			// Warn but don't abort — partial generation is better than none.
			_, _ = fmt.Printf("doppelgen: warning: could not parse %q: %v\n", importPath, depErr)
			continue
		}

		parsed[importPath] = depResult
		newDeps := collectInternalImportPaths(depResult, resolver)
		for _, dep := range newDeps {
			if !visited[dep] {
				queue = append(queue, dep)
			}
		}
	}

	// ── Step 4: Merge structs across packages ─────────────────────────────
	merged := make(types.TypeInfo)
	var skipped []SkipInfo

	for importPath, result := range parsed {
		pkgName := result.Package
		for typeName, info := range result.Structs {
			// Key = "pkgName.TypeName" for cross-package disambiguation.
			// Exception: initial package types keep their plain name for backward compat.
			key := typeName
			if importPath != initialImportPath {
				key = pkgName + "." + typeName
			}
			merged[key] = info
		}
		skipped = append(skipped, result.Skipped...)
	}

	// ── Step 5: Topological sort across all packages ───────────────────────
	sorted, topoErr := ResolveDependencies(merged)
	if topoErr != nil {
		return nil, fmt.Errorf("resolve cross-package dependencies: %w", topoErr)
	}

	return &ProjectParseResult{
		Structs:          merged,
		TopologicalOrder: sorted,
		Packages:         parsed,
		InitialPackage:   initialImportPath,
		Skipped:          skipped,
	}, nil
}

// -------------------------------------------- Internal Helpers --------------------------------------------

// parseProjectFallback wraps the original single-package ParsePackage for cases
// where module detection is unavailable (e.g. outside a Go module).
func parseProjectFallback(targetDir, tagKey string) (*ProjectParseResult, error) {
	result, err := ParsePackage(targetDir, tagKey)
	if err != nil {
		return nil, err
	}

	filtered := FilterStructs(result.Structs, nil)
	sorted, err := ResolveDependencies(filtered)
	if err != nil {
		return nil, err
	}

	return &ProjectParseResult{
		Structs:          result.Structs,
		TopologicalOrder: sorted,
		Packages:         map[string]*ParseResult{"": result},
		InitialPackage:   "",
		Skipped:          result.Skipped,
	}, nil
}

// collectInternalImportPaths scans all fields in a ParseResult and returns the
// import paths of project-internal types that should be enqueued for parsing.
// Each import path is deduplicated before returning.
func collectInternalImportPaths(result *ParseResult, resolver *TypeResolver) []string {
	seen := make(map[string]bool)
	var paths []string

	enqueue := func(importPath string, isThirdParty bool) {
		if importPath == "" || isThirdParty || seen[importPath] {
			return
		}
		if resolver.IsProjectInternal(importPath) {
			seen[importPath] = true
			paths = append(paths, importPath)
		}
	}

	for _, info := range result.Structs {
		if info.Skip {
			continue
		}
		for _, field := range info.Fields {
			enqueue(field.ImportPath, field.IsThirdParty)
			enqueue(field.ElemImportPath, field.ElemIsThirdParty)
			enqueue(field.ValueImportPath, field.ValueIsThirdParty)
		}
	}

	sort.Strings(paths) // deterministic ordering
	return paths
}

// dirToImportPath converts a filesystem directory to a Go import path by computing
// the relative path from the module root and joining it with the module path.
//
// Example:
//
//	dirToImportPath("/home/user/project/manual",
//	                "/home/user/project",
//	                "github.com/seyallius/doppel")
//	→ "github.com/seyallius/doppel/manual"
func dirToImportPath(dir, moduleRoot, modulePath string) (string, error) {
	// Use filepath.Rel to get the path relative to module root.
	rel, err := relPath(moduleRoot, dir)
	if err != nil {
		return "", fmt.Errorf("rel path from %q to %q: %w", moduleRoot, dir, err)
	}

	if rel == "." {
		return modulePath, nil
	}

	// Convert OS path separators to forward slashes for import path.
	return modulePath + "/" + toSlash(rel), nil
}

// relPath returns the relative path from base to target.
// It's a wrapper around filepath.Rel that handles edge cases.
func relPath(base, target string) (string, error) {
	// Clean the paths to handle any redundant separators or relative elements.
	base = filepath.Clean(base)
	target = filepath.Clean(target)

	rel, err := filepath.Rel(base, target)
	if err != nil {
		return "", err
	}

	// On Windows, filepath.Rel might return paths with backslashes,
	// but we want to keep them as-is for now since toSlash will handle conversion.
	// However, we need to ensure "." is returned as is.
	if rel == "." {
		return rel, nil
	}

	return rel, nil
}

// toSlash converts OS-specific path separators to forward slashes.
// This is useful for converting filesystem paths to import paths.
func toSlash(path string) string {
	if path == "" {
		return ""
	}

	// Use filepath.ToSlash which handles all OS-specific separators.
	// On Unix systems, this does nothing (since separator is already '/').
	// On Windows, it converts backslashes to forward slashes.
	return filepath.ToSlash(path)
}

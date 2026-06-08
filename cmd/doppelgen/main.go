// Package main. main.go - Implements the doppelgen CLI tool — a code generator that reads
// Go source files with doppel struct tags and emits Clone() method implementations.
//
// Usage:
//
//	doppelgen --type=User,Order --package=mypackage --output=./generated
//	doppelgen --type=User --preview
//	doppelgen --package=. --output=./
//	doppelgen --package=./models --module-root=/home/user/project
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/seyallius/doppel/cmd/doppelgen/internal/emitter"
	"github.com/seyallius/doppel/cmd/doppelgen/internal/parser"
	"github.com/seyallius/doppel/cmd/doppelgen/internal/types"
)

// -------------------------------------------- Main --------------------------------------------

//go:generate go run github.com/seyallius/doppel/cmd/doppelgen --package=./testdata/complex/object --preview

// main is the entry point for the doppelgen CLI. It creates the cobra root command
// and executes it, handling top-level error printing and exit codes.
func main() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// run orchestrates the CLI workflow: directory resolution, recursive AST parsing
// (with cross-package discovery), dependency sorting, code generation, and file
// output (or preview mode).
//
// run uses parser.ParseProject instead of parser.ParsePackage.
// ParseProject auto-detects the module root, resolves cross-package internal types
// transitively, and flags third-party types for convention-function stubs.
// If module detection fails gracefully, single-package mode is used transparently.
func run(cfg *types.GeneratorConfig) error {
	// ── Step 1: Resolve absolute paths for all target directories ─────────
	var absPackages []string
	for _, pkg := range cfg.Packages {
		absDir, err := filepath.Abs(pkg)
		if err != nil {
			return fmt.Errorf("resolve target directory %q: %w", pkg, err)
		}
		absPackages = append(absPackages, absDir)
	}

	// Detect module root from the first package to share across all parses (optimization)
	moduleRoot := cfg.ModuleRoot
	if moduleRoot == "" && len(absPackages) > 0 {
		if detectedRoot, err := parser.FindModuleRoot(absPackages[0]); err == nil {
			moduleRoot = detectedRoot
		}
	}

	// ── Step 2: Parse all packages and merge results ──────────────────────
	mergedStructs := make(types.TypeInfo)
	var allSkipped []parser.SkipInfo

	for _, absDir := range absPackages {
		projectResult, err := parser.ParseProject(absDir, cfg.Tag, moduleRoot)
		if err != nil {
			return fmt.Errorf("parse project at %q: %w", absDir, err)
		}
		for key, info := range projectResult.Structs {
			mergedStructs[key] = info
		}
		allSkipped = append(allSkipped, projectResult.Skipped...)
	}

	if len(mergedStructs) == 0 {
		return fmt.Errorf("no Go files found in provided packages")
	}

	// ── Step 3: Filter to requested types ─────────────────────────────────
	filtered := parser.FilterStructs(mergedStructs, cfg.TypeNames)
	if len(filtered) == 0 {
		return fmt.Errorf(
			"no structs eligible for generation (found %d structs, %d skipped)",
			len(mergedStructs), len(allSkipped),
		)
	}

	// ── Step 4: Report skipped types in preview mode ──────────────────────
	if len(allSkipped) > 0 && cfg.Preview {
		_, _ = fmt.Fprintf(os.Stderr, "// Skipped types:\n")
		for _, s := range allSkipped {
			_, _ = fmt.Fprintf(os.Stderr, "//   %s: %s (%s)\n", s.TypeName, s.Reason, s.File)
		}
		_, _ = fmt.Fprintf(os.Stderr, "\n")
	}

	// ── Step 5: Unified topological sort across all packages ──────────────
	sortedKeys, err := parser.ResolveDependencies(filtered)
	if err != nil {
		return fmt.Errorf("resolve cross-package dependencies: %w", err)
	}

	// ── Step 6: Generate Clone() methods ──────────────────────────────────
	for _, typeName := range sortedKeys {
		info := filtered[typeName]
		code, genErr := emitter.Generate(info)
		if genErr != nil {
			return fmt.Errorf("generate %s: %w", typeName, genErr)
		}

		if cfg.Preview {
			_, _ = fmt.Fprintf(os.Stdout, "// --- %s.clone_gen.go ---\n", strings.ToLower(plainTypeName(typeName)))
			_, _ = fmt.Fprintln(os.Stdout, code)
			_, _ = fmt.Fprintln(os.Stdout)
		}
	}

	// ── Step 7: Write files (non-preview mode) ────────────────────────────
	if !cfg.Preview {
		for _, typeName := range sortedKeys {
			info := filtered[typeName]
			code, genErr := emitter.Generate(info)
			if genErr != nil {
				return fmt.Errorf("generate %s: %w", typeName, genErr)
			}

			outputDirForStruct := determineOutputDirectory(cfg, info)
			if mkErr := os.MkdirAll(outputDirForStruct, 0755); mkErr != nil {
				return fmt.Errorf("create output directory for %s: %w", typeName, mkErr)
			}

			// Use the plain type name (without "pkg." prefix) as the file name.
			baseName := plainTypeName(typeName)
			fileName := filepath.Join(outputDirForStruct, fmt.Sprintf("%s.clone_gen.go", strings.ToLower(baseName)))

			if writeErr := os.WriteFile(fileName, []byte(code), 0644); writeErr != nil {
				return fmt.Errorf("write %s: %w", fileName, writeErr)
			}
			_, _ = fmt.Fprintf(os.Stdout, "  ✓ %s\n", fileName)

			// Generate companion test file.
			testCode, testErr := emitter.GenerateTest(info)
			if testErr != nil {
				_, _ = fmt.Fprintf(os.Stderr, "  ⚠ generate test for %s: %v\n", typeName, testErr)
			} else {
				testFileName := filepath.Join(outputDirForStruct, fmt.Sprintf("%s.clone_gen_test.go", strings.ToLower(baseName)))
				if writeErr := os.WriteFile(testFileName, []byte(testCode), 0644); writeErr != nil {
					return fmt.Errorf("write %s: %w", testFileName, writeErr)
				}
				_, _ = fmt.Fprintf(os.Stdout, "  ✓ %s\n", testFileName)
			}
		}

		_, _ = fmt.Fprintf(os.Stdout,
			"\nGenerated %d Clone() implementation(s) + test file(s)\n",
			len(sortedKeys))
	}

	return nil
}

// determineOutputDirectory selects the correct output directory for a generated Clone() file.
//
// Rules:
//   - If cfg.Output is non-empty: always use cfg.Output (user explicitly requested single dir).
//   - Otherwise: use the directory of the source file (filepath.Dir(info.File)).
//
// This ensures generated files are colocated with their source definitions by default,
// inherently supporting multi-package generation while preserving backward compatibility.
func determineOutputDirectory(cfg *types.GeneratorConfig, info *types.StructInfo) string {
	if cfg.Output != "" {
		// User explicitly requested all files in one directory — respect that choice.
		return cfg.Output
	}
	if info.File != "" {
		return filepath.Dir(info.File)
	}
	// Fallback (should rarely happen if parser populates info.File correctly)
	return "."
}

// plainTypeName strips the "pkg." prefix from a cross-package qualified key
// (e.g. "pkgB.Address" → "Address"). Used for file naming.
func plainTypeName(qualifiedName string) string {
	if idx := strings.LastIndex(qualifiedName, "."); idx >= 0 {
		return qualifiedName[idx+1:]
	}
	return qualifiedName
}

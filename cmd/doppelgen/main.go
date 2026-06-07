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
	// ── Step 1: Resolve target directory ─────────────────────────────────
	targetDir := cfg.Package
	if targetDir == "" {
		targetDir = "."
	}

	var err error
	targetDir, err = filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("resolve target directory: %w", err)
	}

	// ── Step 2: Parse project (recursive, cross-package aware) ───────────
	projectResult, err := parser.ParseProject(targetDir, cfg.Tag, cfg.ModuleRoot)
	if err != nil {
		return fmt.Errorf("parse project: %w", err)
	}

	// Derive a flat ParseResult for compat with FilterStructs / ResolveDependencies.
	// We use projectResult.Structs (merged across all packages) and
	// projectResult.TopologicalOrder (cross-package topo sort).
	if len(projectResult.Structs) == 0 {
		return fmt.Errorf("no Go files found in %q", targetDir)
	}

	// ── Step 3: Filter to requested types ────────────────────────────────
	filtered := parser.FilterStructs(projectResult.Structs, cfg.TypeNames)
	if len(filtered) == 0 {
		return fmt.Errorf(
			"no structs eligible for generation in %q (found %d structs, %d skipped)",
			targetDir, len(projectResult.Structs), len(projectResult.Skipped),
		)
	}

	// ── Step 4: Report skipped types in preview mode ──────────────────────
	if len(projectResult.Skipped) > 0 && cfg.Preview {
		_, _ = fmt.Fprintf(os.Stderr, "// Skipped types:\n")
		for _, s := range projectResult.Skipped {
			_, _ = fmt.Fprintf(os.Stderr, "//   %s: %s (%s)\n", s.TypeName, s.Reason, s.File)
		}
		_, _ = fmt.Fprintf(os.Stderr, "\n")
	}

	// ── Step 5: Use pre-computed topological order (cross-package aware) ──
	// Filter the topo order to only include types in the filtered set.
	var sortedKeys []string
	for _, key := range projectResult.TopologicalOrder {
		if _, ok := filtered[key]; ok {
			sortedKeys = append(sortedKeys, key)
		}
	}

	// ── Step 6: Generate Clone() methods ──────────────────────────────────
	for _, typeName := range sortedKeys {
		info := filtered[typeName]

		code, genErr := emitter.Generate(info)
		if genErr != nil {
			return fmt.Errorf("generate %s: %w", typeName, genErr)
		}

		if cfg.Preview {
			_, _ = fmt.Fprintf(os.Stdout, "// --- %s.clone_gen.go ---\n", strings.ToLower(typeName))
			_, _ = fmt.Fprintln(os.Stdout, code)
			_, _ = fmt.Fprintln(os.Stdout)
		}
	}

	// ── Step 7: Write files (non-preview mode) ────────────────────────────
	if !cfg.Preview {
		// Create default output dir only if explicitly requested via --output.
		if cfg.Output != "" {
			if mkErr := os.MkdirAll(cfg.Output, 0755); mkErr != nil {
				return fmt.Errorf("create output directory: %w", mkErr)
			}
		}

		for _, typeName := range sortedKeys {
			info := filtered[typeName]

			code, genErr := emitter.Generate(info)
			if genErr != nil {
				return fmt.Errorf("generate %s: %w", typeName, genErr)
			}

			outputDirForStruct := determineOutputDirectory(cfg, typeName, info, targetDir)
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
//   - If typeName contains "." (cross-package, e.g., "auth.Role"): use filepath.Dir(info.File).
//   - Otherwise (initial package type): use targetDir.
//
// This ensures generated files are colocated with their source definitions by default,
// while preserving backward compatibility when --output is explicitly set. (◕‿◕)
func determineOutputDirectory(cfg *types.GeneratorConfig, typeName string, info *types.StructInfo, targetDir string) string {
	if cfg.Output != "" {
		// User explicitly requested all files in one directory — respect that choice.
		return cfg.Output
	}
	if strings.Contains(typeName, ".") && info.File != "" {
		// Cross-package internal type: write to the directory where the struct is defined.
		// info.File contains the absolute path to the source file, so filepath.Dir gives us the package dir.
		return filepath.Dir(info.File)
	}
	// Initial package type or fallback: use the target directory.
	return targetDir
}

// plainTypeName strips the "pkg." prefix from a cross-package qualified key
// (e.g. "pkgB.Address" → "Address"). Used for file naming.
func plainTypeName(qualifiedName string) string {
	if idx := strings.LastIndex(qualifiedName, "."); idx >= 0 {
		return qualifiedName[idx+1:]
	}
	return qualifiedName
}

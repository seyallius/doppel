// Package main. main.go - Implements the doppelgen CLI tool — a code generator that reads
// Go source files with doppel struct tags and emits Clone() method implementations.
//
// Usage:
//
//	doppelgen -type=User,Order -package=mypackage -output=./generated
//	doppelgen -type=User -preview
//	doppelgen -package=. -output=./
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

// main is the entry point for the doppelgen CLI. It creates the cobra root command
// and executes it, handling top-level error printing and exit codes.
func main() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// run orchestrates the CLI workflow: directory resolution, AST parsing,
// dependency sorting, code generation, and file output (or preview).
func run(cfg *types.GeneratorConfig) error {
	// Determine the directory to parse.
	targetDir := cfg.Package
	if targetDir == "" {
		targetDir = "."
	}
	targetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("resolve target directory: %w", err)
	}

	// Parse the package.
	result, err := parser.ParsePackage(targetDir, cfg.Tag)
	if err != nil {
		return fmt.Errorf("parse package: %w", err)
	}

	if result.FileCount == 0 {
		return fmt.Errorf("no Go files found in %q", targetDir)
	}

	// Filter to requested types or all tagged structs.
	filtered := parser.FilterStructs(result.Structs, cfg.TypeNames)
	if len(filtered) == 0 {
		return fmt.Errorf("no structs eligible for generation in %q (found %d structs, %d skipped)",
			targetDir, len(result.Structs), len(result.Skipped))
	}

	// Log skipped types.
	if len(result.Skipped) > 0 && cfg.Preview {
		_, _ = fmt.Fprintf(os.Stderr, "// Skipped types:\n")
		for _, s := range result.Skipped {
			_, _ = fmt.Fprintf(os.Stderr, "//   %s: %s (%s)\n", s.TypeName, s.Reason, s.File)
		}
		_, _ = fmt.Fprintf(os.Stderr, "\n")
	}

	// Resolve dependencies and topologically sort.
	sorted, err := parser.ResolveDependencies(filtered)
	if err != nil {
		return fmt.Errorf("resolve dependencies: %w", err)
	}

	// Generate Clone() methods.
	for _, typeName := range sorted {
		info := filtered[typeName]
		var code string
		code, err = emitter.Generate(info)
		if err != nil {
			return fmt.Errorf("generate %s: %w", typeName, err)
		}

		if cfg.Preview {
			_, _ = fmt.Fprintf(os.Stdout, "// --- %s_clone.gen.go ---\n", strings.ToLower(typeName))
			_, _ = fmt.Fprintln(os.Stdout, code)
			_, _ = fmt.Fprintln(os.Stdout)
		}
	}

	// Write files unless in preview mode.
	if !cfg.Preview {
		outputDir := cfg.Output
		if outputDir == "" {
			outputDir = targetDir
		}

		if err = os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}

		for _, typeName := range sorted {
			info := filtered[typeName]
			var code string
			code, err = emitter.Generate(info)
			if err != nil {
				return fmt.Errorf("generate %s: %w", typeName, err)
			}

			fileName := filepath.Join(outputDir, fmt.Sprintf("%s_clone.gen.go", strings.ToLower(typeName)))
			if err = os.WriteFile(fileName, []byte(code), 0644); err != nil {
				return fmt.Errorf("write %s: %w", fileName, err)
			}

			_, _ = fmt.Fprintf(os.Stdout, "  ✓ %s\n", fileName)
		}

		_, _ = fmt.Fprintf(os.Stdout, "\nGenerated %d Clone() implementation(s) in %s\n", len(sorted), outputDir)
	}

	return nil
}

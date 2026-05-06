//go:build ignore

// This file is an integration test that verifies the full parser → emitter → compile → run pipeline.
// It is run via `go run` rather than `go test` because it generates code dynamically.
package doppelgen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/seyallius/doppel/cmd/doppelgen/internal/emitter"
	"github.com/seyallius/doppel/cmd/doppelgen/internal/parser"
	"github.com/seyallius/doppel/cmd/doppelgen/internal/types"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("PASS: all integration checks")
}

func run() error {
	// 1. Parse test fixtures.
	testdataDir := filepath.Join("testdata")
	result, err := parser.ParsePackage(testdataDir, "doppel")
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	fmt.Printf("Parsed %d structs from %d files\n", len(result.Structs), result.FileCount)

	// 2. Filter non-skipped.
	filtered := parser.FilterStructs(result.Structs, nil)
	fmt.Printf("Filtered to %d non-skipped structs\n", len(filtered))

	if len(filtered) == 0 {
		return fmt.Errorf("no structs to generate")
	}

	// 3. Resolve dependencies.
	sorted, err := parser.ResolveDependencies(filtered)
	if err != nil {
		return fmt.Errorf("resolve deps: %w", err)
	}
	fmt.Printf("Topological order: %v\n", sorted)

	// 4. Generate code for each struct.
	for _, typeName := range sorted {
		info := filtered[typeName]
		code, err := emitter.Generate(info)
		if err != nil {
			return fmt.Errorf("generate %s: %w", typeName, err)
		}

		// 4a. Check for expected patterns in generated code.
		if err := verifyGeneratedCode(code, typeName, info); err != nil {
			return fmt.Errorf("verify %s: %w", typeName, err)
		}

		fmt.Printf("  ✓ %s generated (%d bytes)\n", typeName, len(code))
	}

	// 5. Log skipped types.
	if len(result.Skipped) > 0 {
		fmt.Printf("\nSkipped types (%d):\n", len(result.Skipped))
		for _, s := range result.Skipped {
			fmt.Printf("  - %s: %s\n", s.TypeName, s.Reason)
		}
	}

	return nil
}

func verifyGeneratedCode(code, typeName string, info *types.StructInfo) error {
	// Must contain package declaration.
	if !contains(code, "package testdata") {
		return fmt.Errorf("missing package declaration")
	}

	// Must contain nil guard.
	if !contains(code, "if x == nil") {
		return fmt.Errorf("missing nil guard")
	}

	// Must contain method signature.
	expectedSig := fmt.Sprintf("func (x *%s) Clone() (*%s, error)", typeName, typeName)
	if !contains(code, expectedSig) {
		return fmt.Errorf("missing method signature: %s", expectedSig)
	}

	// Must contain return statement.
	expectedReturn := fmt.Sprintf("return &%s{", typeName)
	if !contains(code, expectedReturn) {
		return fmt.Errorf("missing return statement")
	}

	// All field names must appear in the return statement.
	for _, field := range info.Fields {
		if field.Directive.Skip {
			continue // skipped fields are not in the return
		}
		if !contains(code, fmt.Sprintf("%s: %s,", field.Name, field.Name)) {
			return fmt.Errorf("field %s missing from return statement", field.Name)
		}
	}

	// Check tag-specific patterns.
	for _, field := range info.Fields {
		switch {
		case field.Directive.Skip:
			if contains(code, fmt.Sprintf("%s := x.%s", field.Name, field.Name)) {
				return fmt.Errorf("skipped field %s should not be assigned", field.Name)
			}
		case field.Directive.Shallow:
			if !contains(code, fmt.Sprintf("%s := x.%s", field.Name, field.Name)) {
				return fmt.Errorf("shallow field %s should be assigned from x", field.Name)
			}
		case field.Directive.Empty && field.TypeCategory != types.CatPrimitive:
			if !contains(code, fmt.Sprintf("%s := %s{", field.Name, field.Type)) &&
				!contains(code, fmt.Sprintf("%s := &%s{", field.Name, field.PointedToType)) {
				return fmt.Errorf("empty field %s should have empty literal", field.Name)
			}
		case field.TypeCategory == types.CatSlice && field.Directive.Deep:
			if !contains(code, "manual.CloneSlice") {
				return fmt.Errorf("slice field should use CloneSlice")
			}
		case field.TypeCategory == types.CatMap && field.Directive.Deep:
			if !contains(code, "manual.CloneMap") {
				return fmt.Errorf("map field should use CloneMap")
			}
		}
	}

	return nil
}

func contains(s, sub string) bool {
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

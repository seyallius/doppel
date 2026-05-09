// Package main. cli.go - Handles cobra command definition, flag registration, validation,
// and configuration mapping for the doppelgen tool.
package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/seyallius/doppel/cmd/doppelgen/internal/types"
	"github.com/spf13/cobra"
)

// -------------------------------------------- Public API --------------------------------------------

// newRootCmd creates and returns the root cobra command for doppelgen.
// It registers flags for type filtering, package targeting, output directory, preview mode, and custom tag keys.
// If type names are provided, they are validated against Go identifier rules.
// Returns an error during execution if any type name is invalid.
func newRootCmd() *cobra.Command {
	var (
		typeNames string
		pkg       string
		output    string
		preview   bool
		tag       string
	)

	cmd := &cobra.Command{
		Use:   "doppelgen",
		Short: "Generate Clone() method implementations from doppel struct tags",
		Long: `doppelgen is a code generator that reads Go source files with doppel struct tags
and emits Clone() method implementations. It analyses struct fields, resolves type
dependencies, and produces idiomatic Go code with proper deep/shallow/empty/skip semantics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &types.GeneratorConfig{
				Output:  output,
				Preview: preview,
				Tag:     tag,
				Package: pkg,
			}

			if typeNames != "" {
				cfg.TypeNames = splitComma(typeNames)
				// Validate that type names are valid identifiers.
				for _, name := range cfg.TypeNames {
					if !isValidGoIdent(name) {
						return fmt.Errorf("invalid type name %q: must be a valid Go identifier", name)
					}
				}
			}

			return run(cfg)
		},
	}

	// Register flags — matching the original flagSet interface.
	cmd.Flags().StringVar(&typeNames, "type", "", "Comma-separated list of type names to generate (default: all tagged structs)")
	cmd.Flags().StringVar(&pkg, "package", "", "Target package directory (default: current directory)")
	cmd.Flags().StringVar(&output, "output", "", "Output directory for generated files (default: package directory)")
	cmd.Flags().BoolVar(&preview, "preview", false, "Print generated code to stdout without writing files")
	cmd.Flags().StringVar(&tag, "tag", "doppel", "Struct tag key to look for")

	return cmd
}

// -------------------------------------------- Internal Helpers --------------------------------------------

// splitComma splits a comma-separated string into a sorted slice of trimmed, non-empty strings.
// It safely handles extra whitespace and ignores empty segments.
func splitComma(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	sort.Strings(result)
	return result
}

// isValidGoIdent checks whether the provided string conforms to Go's identifier syntax rules.
// Identifiers must start with a letter, followed by any number of letters, digits, or underscores.
func isValidGoIdent(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if r < 'A' || (r > 'Z' && r < 'a') || r > 'z' {
				return false
			}
		} else {
			if r < '0' || (r > '9' && r < 'A') || (r > 'Z' && r < 'a') || r > 'z' {
				if r != '_' {
					return false
				}
			}
		}
	}
	return true
}

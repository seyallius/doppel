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
// It registers flags for type filtering, package targeting, output directory, preview mode,
// custom tag keys, and an optional module-root override for cross-package resolution.
// If type names are provided, they are validated against Go identifier rules.
// Returns an error during execution if any type name is invalid.
func newRootCmd() *cobra.Command {
	var (
		typeNames  string
		packages   []string
		output     string
		preview    bool
		tag        string
		moduleRoot string
	)

	cmd := &cobra.Command{
		Use:   "doppelgen",
		Short: "Generate Clone() method implementations from doppel struct tags",
		Long: `doppelgen is a code generator that reads Go source files with doppel struct tags
and emits Clone() method implementations. It analyses struct fields, resolves type
dependencies across packages, and produces idiomatic Go code with proper
deep/shallow/empty/skip semantics.

Cross-package support:
  - Project-internal types (under the same Go module) are parsed recursively,
    and Clone() implementations are generated for their structs automatically.
  - Third-party types (from external modules) receive convention-function stubs
    with //todo comments indicating the expected function signature.
  - Multiple packages can be specified using multiple --package flags.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &types.GeneratorConfig{
				Output:     output,
				Preview:    preview,
				Tag:        tag,
				Packages:   packages,
				ModuleRoot: moduleRoot,
			}

			// Validate type names (CLI-specific syntax check)
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
	cmd.Flags().StringVarP(&typeNames, "type", "t", "", "Comma-separated list of type names to generate (default: all tagged structs)")
	cmd.Flags().StringArrayVarP(&packages, "package", "p", nil, "Target package directory (can be specified multiple times)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output directory for generated files (default: package directory, locked for multi-package)")
	cmd.Flags().BoolVar(&preview, "preview", false, "Print generated code to stdout without writing files")
	cmd.Flags().StringVar(&tag, "tag", "doppel", "Struct tag key to look for")
	cmd.Flags().StringVar(&moduleRoot, "module-root", "",
		"Override the Go module root directory. Defaults to auto-detection by walking up from --package until go.mod is found.\n"+
			"Use this when go.mod is not in a parent of --package, or when running in unusual project layouts.")

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

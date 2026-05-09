// Package main. cli.go - Handles command-line flag parsing, validation,
// and configuration mapping for the doppelgen tool.
package main

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/seyallius/doppel/cmd/doppelgen/internal/types"
)

// -------------------------------------------- Public API --------------------------------------------

// parseFlags parses the provided CLI arguments and returns a fully populated GeneratorConfig.
// It registers flags for type filtering, package targeting, output directory, preview mode, and custom tag keys.
// If type names are provided, they are validated against Go identifier rules.
// Returns an error if flag parsing fails or if any type name is invalid.
func parseFlags(args []string) (*types.GeneratorConfig, error) {
	fs := flag.NewFlagSet("doppelgen", flag.ContinueOnError)

	var (
		typeNames string
		pkg       string
		output    string
		preview   bool
		tag       string
	)

	fs.StringVar(&typeNames, "type", "", "Comma-separated list of type names to generate (default: all tagged structs)")
	fs.StringVar(&pkg, "package", "", "Target package directory (default: current directory)")
	fs.StringVar(&output, "output", "", "Output directory for generated files (default: package directory)")
	fs.BoolVar(&preview, "preview", false, "Print generated code to stdout without writing files")
	fs.StringVar(&tag, "tag", "doppel", "Struct tag key to look for (default: doppel)")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("parse flags: %w", err)
	}

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
				return nil, fmt.Errorf("invalid type name %q: must be a valid Go identifier", name)
			}
		}
	}

	return cfg, nil
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

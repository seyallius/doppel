// Package main. generate-nav.go - Injects DRY navigation HTML into documentation markdown files.
package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// -------------------------------------------- Types, Variables & Constants --------------------------------------------

// NavConfig defines the navigation structure loaded from navigation.json.
type NavConfig struct {
	BaseURL string   `json:"baseUrl"`
	Docs    []NavDoc `json:"docs"`
}

// NavDoc represents a single documentation page's navigation metadata.
type NavDoc struct {
	File      string `json:"file"`
	Title     string `json:"title"`
	Prev      string `json:"prev"`
	PrevLabel string `json:"prevLabel"`
	Next      string `json:"next"`
	NextLabel string `json:"nextLabel"`
}

const navStartMarker = "<!-- Navigation (AUTO-GENERATED - DO NOT EDIT) -->"

// navTemplateStr is the reusable navigation component evaluated via text/template.
//
//go:embed navigation.html
var navTemplateStr string

// -------------------------------------------- Main --------------------------------------------

// main is the entry point. Loads navigation.json, evaluates the template for each doc,
// and writes the result back to the markdown file only if content actually changed.
func main() {
	config, err := loadConfig("docs/navigation.json")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "❌ Error loading config: %v\n", err)
		os.Exit(1)
	}

	tmpl, err := template.New("nav").Parse(navTemplateStr)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "❌ Error parsing template: %v\n", err)
		os.Exit(1)
	}

	for _, doc := range config.Docs {
		// Inject the start marker into the doc for template evaluation
		docData := struct {
			NavDoc
			NavStart string
		}{
			NavDoc:   doc,
			NavStart: navStartMarker,
		}

		if err = injectNavigation(doc.File, tmpl, docData); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "❌ Error processing %s: %v\n", doc.File, err)
			os.Exit(1)
		}
	}
	fmt.Println("✨ Navigation injection complete!")
}

// -------------------------------------------- Internal Helpers --------------------------------------------

// loadConfig reads and unmarshals the navigation.json file.
func loadConfig(path string) (*NavConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	var config NavConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}
	return &config, nil
}

// injectNavigation reads a markdown file, strips any existing auto-generated navigation,
// renders the new navigation block, and writes it back only if content changed.
func injectNavigation(filename string, tmpl *template.Template, data any) error {
	filePath := filepath.Join("docs", filename)
	original, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	content := string(original)

	// Strip existing navigation block (if any) by finding the start marker
	if idx := strings.Index(content, navStartMarker); idx != -1 {
		content = content[:idx]
	}

	// Normalize trailing whitespace to ensure exactly one blank line before nav
	content = strings.TrimRight(content, " \t\r\n")

	// Execute template into a buffer
	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	// Assemble final content with predictable spacing
	finalContent := content + "\n\n" + buf.String() + "\n"

	// Idempotent write: only touch disk if content actually changed
	if string(original) == finalContent {
		fmt.Printf("(✓ %s is already up to date)\n", filename)
		return nil
	}

	if err = os.WriteFile(filePath, []byte(finalContent), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	fmt.Printf("✓ Updated %s\n", filename)
	return nil
}

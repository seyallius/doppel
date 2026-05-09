// Package parser. parser.go reads Go source files, extracts struct definitions with doppel struct tags,
// resolves type information, and builds dependency graphs for safe code generation.
//
// It uses go/ast and go/token for parsing — no external dependencies.
package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
	"unicode"

	"github.com/seyallius/doppel/cmd/doppelgen/internal/types"
	"github.com/seyallius/doppel/core"
)

// -------------------------------------------- Types, Variables & Constants --------------------------------------------

const (
	skipGenComment      = "doppel:skip-gen"             // skipGenComment is the opt-out marker for individual struct types.
	skipAllComment      = "doppel:skip-all"             // skipAllComment is the opt-out marker for entire files.
	skipGenerateComment = "go:generate doppelgen -skip" // skipGenerateComment is the alternative opt-out marker (go:generate convention).
)

// ParseResult holds the complete output of a parse run, including discovered structs,
// package metadata, file counts, and any types that were intentionally skipped.
type ParseResult struct {
	Structs   types.TypeInfo // all discovered structs keyed by name
	Package   string         // package name
	FileCount int            // number of files parsed
	Skipped   []SkipInfo     // types that were skipped and why
}

// SkipInfo records why a type was skipped from generation, providing context for debugging or reporting.
type SkipInfo struct {
	TypeName string
	Reason   string
	File     string
}

// -------------------------------------------- Public API --------------------------------------------

// ParsePackage parses all .go files in the given directory and returns the discovered struct types
// with their doppel tag information. It handles multi-package edge cases by selecting the first
// non-test package if multiple are found.
func ParsePackage(dir string, tagKey string) (*ParseResult, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse directory %q: %w", dir, err)
	}

	// Expect a single package
	if len(pkgs) != 1 {
		// Pick the first non-test package
		for name, pkg := range pkgs {
			if !strings.HasSuffix(name, "_test") {
				return parsePackageFiles(fset, pkg, dir, tagKey)
			}
		}
		return nil, fmt.Errorf("expected exactly one package, found %d", len(pkgs))
	}

	// Use the first (and only) package
	for _, pkg := range pkgs {
		return parsePackageFiles(fset, pkg, dir, tagKey)
	}

	return nil, fmt.Errorf("no packages found in %q", dir)
}

// ParseFiles parses specific Go files by path and returns discovered struct types.
// Useful for targeted generation or when bypassing directory-based package resolution.
func ParseFiles(files []string, tagKey string) (*ParseResult, error) {
	fset := token.NewFileSet()
	var astFiles []*ast.File

	for _, f := range files {
		af, err := parser.ParseFile(fset, f, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse file %q: %w", f, err)
		}
		astFiles = append(astFiles, af)
	}

	if len(astFiles) == 0 {
		return &ParseResult{}, nil
	}

	return parseASTFiles(fset, astFiles, tagKey)
}

// HasExistingClone checks whether a struct type already has a Clone() method defined in the parsed files.
// This prevents the generator from overwriting user-defined implementations.
func (r *ParseResult) HasExistingClone(typeName string) bool {
	// This is populated during parsing via method detection.
	// The parser marks structs with existing Clone() methods.
	if info, ok := r.Structs[typeName]; ok {
		return info.Skip && info.SkipReason == "has existing Clone() method"
	}
	return false
}

// -------------------------------------------- Internal --------------------------------------------

// parsePackageFiles converts an ast.Package into a slice of ast.File and delegates to parseASTFiles.
func parsePackageFiles(fset *token.FileSet, pkg *ast.Package, dir string, tagKey string) (*ParseResult, error) {
	var files []*ast.File
	for _, f := range pkg.Files {
		files = append(files, f)
	}
	return parseASTFiles(fset, files, tagKey)
}

// parseASTFiles iterates over a slice of AST files, extracts struct declarations, resolves tags,
// categorizes field types, and builds the final ParseResult.
func parseASTFiles(fset *token.FileSet, files []*ast.File, tagKey string) (*ParseResult, error) {
	result := &ParseResult{
		Structs: make(types.TypeInfo),
		Skipped: []SkipInfo{},
	}

	if len(files) == 0 {
		return result, nil
	}

	// Determine package name from first file
	result.Package = files[0].Name.Name
	result.FileCount = len(files)

	// Collect all method declarations to detect existing Clone() methods.
	existingCloneTypes := collectExistingCloneMethods(files)

	// Collect all type declarations with doppel tags.
	for _, file := range files {
		filePath := fset.Position(file.Pos()).Filename
		if filePath == "" {
			filePath = file.Name.Name
		}

		// Check for file-level skip-all comment.
		if hasFileSkipComment(file) {
			// Skip all types in this file.
			collectSkippedTypes(file, filePath, skipAllComment, result)
			continue
		}

		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				typeName := typeSpec.Name.Name
				doc := extractDocComment(genDecl, typeSpec)

				structInfo := &types.StructInfo{
					Name:    typeName,
					File:    filePath,
					Package: result.Package,
					Doc:     doc,
					Fields:  []types.FieldInfo{},
				}

				// Check skip-gen heuristics.
				if existingCloneTypes[typeName] {
					structInfo.Skip = true
					structInfo.SkipReason = "has existing Clone() method"
					result.Skipped = append(result.Skipped, SkipInfo{
						TypeName: typeName,
						Reason:   structInfo.SkipReason,
						File:     filePath,
					})
					result.Structs[typeName] = structInfo
					continue
				}

				if hasSkipGenComment(doc) {
					structInfo.Skip = true
					structInfo.SkipReason = "has doppel:skip-gen comment"
					result.Skipped = append(result.Skipped, SkipInfo{
						TypeName: typeName,
						Reason:   structInfo.SkipReason,
						File:     filePath,
					})
					result.Structs[typeName] = structInfo
					continue
				}

				// Parse fields.
				for _, field := range structType.Fields.List {
					fieldInfos := parseField(field, typeName, filePath, tagKey)
					structInfo.Fields = append(structInfo.Fields, fieldInfos...)
				}

				result.Structs[typeName] = structInfo
			}
		}
	}

	return result, nil
}

// collectExistingCloneMethods scans all files for method declarations named Clone with a pointer receiver,
// recording which types already have a Clone implementation.
func collectExistingCloneMethods(files []*ast.File) map[string]bool {
	result := make(map[string]bool)

	for _, file := range files {
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if funcDecl.Name.Name != "Clone" {
				continue
			}
			if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
				continue
			}

			// Extract type name from receiver: *T or T.
			recvType := extractReceiverTypeName(funcDecl.Recv.List[0].Type)
			if recvType != "" {
				result[recvType] = true
			}
		}
	}

	return result
}

// extractReceiverTypeName extracts the base type name from a receiver expression.
// Handles pointer receivers (*T) and value receivers (T) by unwrapping StarExpr.
func extractReceiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return extractReceiverTypeName(t.X)
	case *ast.Ident:
		return t.Name
	default:
		return ""
	}
}

// hasFileSkipComment checks if the file's package-level doc comment contains the skip-all marker.
func hasFileSkipComment(file *ast.File) bool {
	if file.Doc == nil {
		return false
	}
	for _, c := range file.Doc.List {
		if strings.Contains(c.Text, skipAllComment) {
			return true
		}
	}
	return false
}

// collectSkippedTypes adds all struct types in a file as skipped entries, typically triggered by file-level opt-out comments.
func collectSkippedTypes(file *ast.File, filePath string, reason string, result *ParseResult) {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if _, ok := typeSpec.Type.(*ast.StructType); !ok {
				continue
			}
			typeName := typeSpec.Name.Name
			result.Structs[typeName] = &types.StructInfo{
				Name:       typeName,
				File:       filePath,
				Package:    result.Package,
				Skip:       true,
				SkipReason: reason,
			}
			result.Skipped = append(result.Skipped, SkipInfo{
				TypeName: typeName,
				Reason:   reason,
				File:     filePath,
			})
		}
	}
}

// hasSkipGenComment checks if a doc comment contains the struct-level skip-gen marker.
func hasSkipGenComment(doc string) bool {
	return strings.Contains(doc, skipGenComment) ||
		strings.Contains(doc, skipGenerateComment)
}

// extractDocComment extracts the doc comment for a type declaration, preferring type-specific docs over group-level docs.
func extractDocComment(genDecl *ast.GenDecl, typeSpec *ast.TypeSpec) string {
	if typeSpec.Doc != nil {
		return typeSpec.Doc.Text()
	}
	if genDecl.Doc != nil && len(genDecl.Specs) == 1 {
		return genDecl.Doc.Text()
	}
	return ""
}

// parseField converts an ast.Field into one or more FieldInfo entries.
// A single ast.Field can declare multiple names (e.g., `x, y int`). It resolves tags, categorizes types, and filters unexported fields.
func parseField(field *ast.Field, structName string, filePath string, tagKey string) []types.FieldInfo {
	var result []types.FieldInfo

	tagValue := ""
	if field.Tag != nil {
		tagValue = reflectTagValue(field.Tag.Value, tagKey)
	}

	doc := ""
	if field.Comment != nil {
		doc = field.Comment.Text()
	}

	typeStr := formatType(field.Type)

	// Handle multiple names in a single field declaration.
	names := field.Names
	if len(names) == 0 {
		// Embedded field
		names = []*ast.Ident{{Name: embeddedFieldName(field.Type)}}
	}

	for _, name := range names {
		if !isExported(name.Name) {
			continue // skip unexported fields
		}

		fi := types.FieldInfo{
			Name:      name.Name,
			Type:      typeStr,
			Tag:       tagValue,
			Doc:       doc,
			File:      filePath,
			Directive: core.ParseTag(tagValue),
		}

		resolveTypeCategory(&fi)
		result = append(result, fi)
	}

	return result
}

// resolveTypeCategory analyzes the field's Go type string and sets the TypeCategory and related metadata.
// This drives which cloning strategy (deep, shallow, empty, etc.) the emitter will use.
func resolveTypeCategory(fi *types.FieldInfo) {
	t := fi.Type

	// Pointer to struct: *StructName
	if strings.HasPrefix(t, "*") {
		inner := t[1:]
		if isBuiltinPrimitive(inner) {
			fi.TypeCategory = types.CatPtrPrimitive
			fi.PointedToType = inner
		} else if isValidIdent(inner) {
			fi.TypeCategory = types.CatPtrStruct
			fi.PointedToType = inner
		} else {
			fi.TypeCategory = types.CatUnknown
		}
		return
	}

	// Slice: []T
	if strings.HasPrefix(t, "[]") {
		fi.TypeCategory = types.CatSlice
		fi.ElemType = t[2:]
		return
	}

	// Map: map[K]V
	if strings.HasPrefix(t, "map[") {
		fi.TypeCategory = types.CatMap
		key, val := parseMapType(t)
		fi.KeyType = key
		fi.ValueType = val
		return
	}

	// Primitive or struct
	if isBuiltinPrimitive(t) {
		fi.TypeCategory = types.CatPrimitive
		return
	}

	if isValidIdent(t) {
		fi.TypeCategory = types.CatStruct
		return
	}

	// Interface or complex type
	if strings.HasPrefix(t, "interface") {
		fi.TypeCategory = types.CatInterface
		return
	}

	fi.TypeCategory = types.CatUnknown
}

// parseMapType extracts key and value types from a "map[K]V" string representation.
// It safely handles nested brackets in complex key types.
func parseMapType(s string) (string, string) {
	// s = "map[K]V"
	// Find matching bracket.
	start := 4 // len("map[")
	depth := 1
	for i := start; i < len(s); i++ {
		if s[i] == '[' {
			depth++
		} else if s[i] == ']' {
			depth--
			if depth == 0 {
				return s[start:i], s[i+1:]
			}
		}
	}
	return s[start:], ""
}

// formatType converts an ast.Expr into its canonical Go type string representation.
// Handles identifiers, pointers, arrays, maps, interfaces, selectors, and anonymous structs.
func formatType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + formatType(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + formatType(t.Elt)
		}
		return fmt.Sprintf("[%s]%s", formatType(t.Len), formatType(t.Elt))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", formatType(t.Key), formatType(t.Value))
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.SelectorExpr:
		// pkg.Type — for cross-package types we just use the type name
		return fmt.Sprintf("%s.%s", formatType(t.X), t.Sel.Name)
	case *ast.StructType:
		return "struct{}"
	default:
		return "unknown"
	}
}

// reflectTagValue extracts the value for a given tag key from a raw struct tag string.
// Manually parses the `key:"value"` format to avoid importing reflect at runtime.
func reflectTagValue(rawTag, tagKey string) string {
	// Strip backticks.
	rawTag = strings.Trim(rawTag, "`")

	// Look for key:"value" pairs.
	for rawTag != "" {
		// Skip whitespace.
		i := 0
		for i < len(rawTag) && rawTag[i] == ' ' {
			i++
		}
		rawTag = rawTag[i:]
		if rawTag == "" {
			break
		}

		// Extract key.
		i = 0
		for i < len(rawTag) && rawTag[i] != ':' && rawTag[i] != ' ' && rawTag[i] != '"' {
			i++
		}
		key := rawTag[:i]
		rawTag = rawTag[i:]

		if rawTag == "" {
			break
		}

		// Expect ':'.
		if rawTag[0] != ':' {
			continue
		}
		rawTag = rawTag[1:]

		// Expect '"'.
		if len(rawTag) == 0 || rawTag[0] != '"' {
			continue
		}
		rawTag = rawTag[1:]

		// Extract value.
		i = 0
		for i < len(rawTag) && rawTag[i] != '"' {
			if rawTag[i] == '\\' {
				i++ // skip escaped char
			}
			i++
		}
		if i >= len(rawTag) {
			break
		}
		value := rawTag[:i]
		rawTag = rawTag[i+1:]

		if key == tagKey {
			return value
		}
	}

	return ""
}

// embeddedFieldName extracts a fallback name for an embedded (anonymous) field.
func embeddedFieldName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return embeddedFieldName(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	default:
		return ""
	}
}

// isExported reports whether the identifier starts with an uppercase letter, indicating it is exported.
func isExported(name string) bool {
	if name == "" {
		return false
	}
	r := []rune(name)
	return unicode.IsUpper(r[0])
}

// isBuiltinPrimitive reports whether the type name matches a Go builtin primitive type.
func isBuiltinPrimitive(t string) bool {
	switch t {
	case "bool", "byte", "complex64", "complex128",
		"error", "float32", "float64", "int", "int8",
		"int16", "int32", "int64", "rune", "string",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"uintptr":
		return true
	default:
		return false
	}
}

// isValidIdent reports whether the string is a valid Go identifier.
func isValidIdent(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' {
				return false
			}
		} else {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return false
			}
		}
	}
	return true
}

// ResolveDependencies builds a dependency graph from the parsed structs and returns a topologically sorted list.
// Inner/dependent types appear first to ensure safe generation order. Detects circular dependencies.
func ResolveDependencies(structs types.TypeInfo) ([]string, error) {
	// Build adjacency list: typeName -> set of dependency type names.
	deps := make(map[string]map[string]bool)
	for name, info := range structs {
		if info.Skip {
			continue
		}
		depSet := make(map[string]bool)
		for _, field := range info.Fields {
			switch field.TypeCategory {
			case types.CatPtrStruct:
				depSet[field.PointedToType] = true
			case types.CatStruct:
				depSet[field.Type] = true
			case types.CatSlice:
				// If element type is a known struct, it's a dependency.
				if _, ok := structs[field.ElemType]; ok && !isBuiltinPrimitive(field.ElemType) {
					depSet[field.ElemType] = true
				}
			case types.CatMap:
				// If value type is a known struct, it's a dependency.
				if _, ok := structs[field.ValueType]; ok && !isBuiltinPrimitive(field.ValueType) {
					depSet[field.ValueType] = true
				}
			}
		}
		deps[name] = depSet
	}

	// Topological sort using Kahn's algorithm.
	inDegree := make(map[string]int)
	for name := range deps {
		if _, ok := inDegree[name]; !ok {
			inDegree[name] = 0
		}
		for dep := range deps[name] {
			if _, ok := deps[dep]; !ok {
				// Dependency is not in our generation set — skip.
				continue
			}
			inDegree[dep]++
		}
	}

	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}
	sort.Strings(queue)

	var sorted []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		sorted = append(sorted, current)

		for dep := range deps[current] {
			if _, ok := deps[dep]; !ok {
				continue
			}
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if len(sorted) != len(deps) {
		// Circular dependency detected.
		var remaining []string
		for name := range deps {
			found := false
			for _, s := range sorted {
				if s == name {
					found = true
					break
				}
			}
			if !found {
				remaining = append(remaining, name)
			}
		}
		sort.Strings(remaining)
		return nil, fmt.Errorf("circular dependency detected among types: %s", strings.Join(remaining, ", "))
	}

	return sorted, nil
}

// FilterStructs returns only the structs matching the given type names.
// If typeNames is empty, all non-skipped structs are returned.
func FilterStructs(structs types.TypeInfo, typeNames []string) types.TypeInfo {
	if len(typeNames) == 0 {
		// Return all non-skipped structs.
		result := make(types.TypeInfo)
		for name, info := range structs {
			if !info.Skip {
				result[name] = info
			}
		}
		return result
	}

	result := make(types.TypeInfo)
	for _, name := range typeNames {
		if info, ok := structs[name]; ok && !info.Skip {
			result[name] = info
		}
	}
	return result
}

// HasDoppelTag reports whether a struct contains at least one field with a non-empty doppel tag.
func HasDoppelTag(info *types.StructInfo) bool {
	for _, field := range info.Fields {
		if field.Tag != "" {
			return true
		}
	}
	return false
}

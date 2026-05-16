// Package parser. parser_extensions.go - shows the key additions and changes to parser.go
// needed to support cross-package type resolution and third-party detection.
//
// These functions integrate with the TypeResolver introduced in resolver.go and
// should be merged into parser.go. They are grouped here for clarity.
package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/seyallius/doppel/cmd/doppelgen/internal/types"
	"github.com/seyallius/doppel/core"
)

// -------------------------------------------- Public Functions --------------------------------------------

// ParsePackageWithResolver is like ParsePackage but accepts a TypeResolver, enabling
// cross-package type detection and third-party classification.
//
// When resolver is nil, the function behaves identically to ParsePackage.
func ParsePackageWithResolver(dir string, tagKey string, resolver *TypeResolver) (*ParseResult, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse directory %q: %w", dir, err)
	}

	if len(pkgs) != 1 {
		for name, pkg := range pkgs {
			if !strings.HasSuffix(name, "_test") {
				return parsePackageFilesWithResolver(fset, pkg, dir, tagKey, resolver)
			}
		}
		return nil, fmt.Errorf("expected exactly one package, found %d", len(pkgs))
	}

	for _, pkg := range pkgs {
		return parsePackageFilesWithResolver(fset, pkg, dir, tagKey, resolver)
	}

	return nil, fmt.Errorf("no packages found in %q", dir)
}

// -------------------------------------------- Internal Helpers --------------------------------------------

// buildFileImports extracts all import declarations from an AST file and returns
// a map from alias (or inferred package name) to the full import path.
//
// Import aliases:
//   - Explicit alias:  import foo "github.com/bar/foo"  → {"foo": "github.com/bar/foo"}
//   - No alias:        import "github.com/bar/baz"       → {"baz": "github.com/bar/baz"}
//   - Blank/dot:       import _ "..." / import . "..."   → skipped
func buildFileImports(file *ast.File) map[string]string {
	result := make(map[string]string, len(file.Imports))

	for _, imp := range file.Imports {
		if imp.Path == nil {
			continue
		}
		path := strings.Trim(imp.Path.Value, `"`)

		// Determine the alias to use in source code.
		if imp.Name != nil {
			switch imp.Name.Name {
			case "_", ".":
				continue // skip blank and dot imports
			default:
				result[imp.Name.Name] = path
			}
		} else {
			// Default: last path segment (e.g. "github.com/foo/bar/baz" → "baz").
			parts := strings.Split(path, "/")
			alias := parts[len(parts)-1]
			result[alias] = path
		}
	}

	return result
}

// resolveImportInfo inspects an AST type expression for a SelectorExpr pattern
// (pkg.TypeName or *pkg.TypeName). When found, it looks up the full import path
// from the file's import map and determines whether the type is third-party.
//
// Returns ("", false) for local types that need no import.
func resolveImportInfo(
	expr ast.Expr,
	fileImports map[string]string,
	resolver *TypeResolver,
) (importPath string, isThirdParty bool) {
	// Unwrap pointer: *pkg.Type → pkg.Type
	inner := expr
	if star, ok := expr.(*ast.StarExpr); ok {
		inner = star.X
	}

	sel, ok := inner.(*ast.SelectorExpr)
	if !ok {
		return "", false // local type, no package qualifier
	}

	pkgAlias := formatType(sel.X)
	fullPath, found := fileImports[pkgAlias]
	if !found {
		return "", false
	}

	isThirdParty = (resolver == nil) || !resolver.IsProjectInternal(fullPath)

	if resolver != nil {
		resolver.WarnIfMissing(fullPath)
	}

	return fullPath, isThirdParty
}

// isSelectorType reports whether s is a cross-package qualified type name of the
// form "pkg.TypeName" — both sides must be valid Go identifiers.
//
// Examples:
//
//	isSelectorType("core.Address")     → true
//	isSelectorType("Address")          → false
//	isSelectorType("map[string]int")   → false
func isSelectorType(s string) bool {
	idx := strings.IndexByte(s, '.')
	if idx <= 0 || idx == len(s)-1 {
		return false
	}
	return isValidIdent(s[:idx]) && isValidIdent(s[idx+1:])
}

// parseField converts an ast.Field into one or more FieldInfo entries.
//
// KEY CHANGES vs original:
//   - Accepts fileImports (alias→importPath map built from the containing file).
//   - Accepts resolver (*TypeResolver, may be nil for backward compat).
//   - Resolves cross-package import paths and sets IsThirdParty, ElemIsThirdParty,
//     ValueIsThirdParty on each FieldInfo.
func parseFieldWithResolver(
	field *ast.Field,
	structName, filePath, tagKey string,
	fileImports map[string]string,
	resolver *TypeResolver,
) []types.FieldInfo {
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

	// ── Resolve cross-package info for the field's own type ──────────────
	importPath, isThirdParty := resolveImportInfo(field.Type, fileImports, resolver)

	// ── Resolve cross-package info for elem/value types inside composites ─
	elemImportPath, elemIsThirdParty := "", false
	valueImportPath, valueIsThirdParty := "", false

	switch t := field.Type.(type) {
	case *ast.ArrayType:
		// []T — check element type
		elemImportPath, elemIsThirdParty = resolveImportInfo(t.Elt, fileImports, resolver)
	case *ast.MapType:
		// map[K]V — check value type (keys are comparable value-types, no Clone needed)
		valueImportPath, valueIsThirdParty = resolveImportInfo(t.Value, fileImports, resolver)
	}

	// Handle multiple names in a single field declaration (e.g. "X, Y int").
	names := field.Names
	if len(names) == 0 {
		names = []*ast.Ident{{Name: embeddedFieldName(field.Type)}}
	}

	for _, name := range names {
		if !isExported(name.Name) {
			continue
		}

		fi := types.FieldInfo{
			Name:      name.Name,
			Type:      typeStr,
			Tag:       tagValue,
			Doc:       doc,
			File:      filePath,
			Directive: core.ParseTag(tagValue),

			// Cross-package resolution.
			ImportPath:        importPath,
			IsThirdParty:      isThirdParty,
			ElemImportPath:    elemImportPath,
			ElemIsThirdParty:  elemIsThirdParty,
			ValueImportPath:   valueImportPath,
			ValueIsThirdParty: valueIsThirdParty,
		}

		resolveTypeCategoryWithContext(&fi)
		result = append(result, fi)
	}

	return result
}

// resolveTypeCategoryWithContext is a drop-in replacement for resolveTypeCategory.
// It adds support for cross-package selector types (pkg.TypeName) and maps them
// to CatPtrStruct / CatStruct (project-internal) or CatThirdPartyPtrStruct /
// CatThirdPartyStruct (external modules) based on fi.IsThirdParty.
func resolveTypeCategoryWithContext(fi *types.FieldInfo) {
	t := fi.Type

	// ── Pointer types: *T, *pkg.T ─────────────────────────────────────────
	if strings.HasPrefix(t, "*") {
		inner := t[1:]
		switch {
		case isBuiltinPrimitive(inner):
			fi.TypeCategory = types.CatPtrPrimitive
			fi.PointedToType = inner
		case isValidIdent(inner):
			fi.TypeCategory = types.CatPtrStruct
			fi.PointedToType = inner
		case isSelectorType(inner):
			// Cross-package pointer: *pkg.TypeName
			fi.PointedToType = inner
			if fi.IsThirdParty {
				fi.TypeCategory = types.CatThirdPartyPtrStruct
			} else {
				fi.TypeCategory = types.CatPtrStruct
			}
		default:
			fi.TypeCategory = types.CatUnknown
		}
		return
	}

	// ── Slices: []T ───────────────────────────────────────────────────────
	if strings.HasPrefix(t, "[]") {
		fi.TypeCategory = types.CatSlice
		fi.ElemType = t[2:]
		return
	}

	// ── Maps: map[K]V ─────────────────────────────────────────────────────
	if strings.HasPrefix(t, "map[") {
		fi.TypeCategory = types.CatMap
		key, val := parseMapType(t)
		fi.KeyType = key
		fi.ValueType = val
		return
	}

	// ── Primitive value types ─────────────────────────────────────────────
	if isBuiltinPrimitive(t) {
		fi.TypeCategory = types.CatPrimitive
		return
	}

	// ── Local struct (same package) ───────────────────────────────────────
	if isValidIdent(t) {
		fi.TypeCategory = types.CatStruct
		return
	}

	// ── Cross-package struct: pkg.TypeName ────────────────────────────────
	if isSelectorType(t) {
		if fi.IsThirdParty {
			fi.TypeCategory = types.CatThirdPartyStruct
		} else {
			fi.TypeCategory = types.CatStruct // treat internal cross-pkg the same
		}
		return
	}

	// ── Interface ─────────────────────────────────────────────────────────
	if strings.HasPrefix(t, "interface") {
		fi.TypeCategory = types.CatInterface
		return
	}

	fi.TypeCategory = types.CatUnknown
}

// parsePackageFilesWithResolver is the internal version of parsePackageFiles that
// threads a TypeResolver through to parseASTFilesWithResolver.
func parsePackageFilesWithResolver(
	fset *token.FileSet,
	pkg *ast.Package,
	dir, tagKey string,
	resolver *TypeResolver,
) (*ParseResult, error) {
	var files []*ast.File
	for _, f := range pkg.Files {
		files = append(files, f)
	}
	return parseASTFilesWithResolver(fset, files, tagKey, resolver)
}

// parseASTFilesWithResolver iterates over a slice of AST files, extracts struct declarations, resolves tags,
// categorizes field types, and builds the final ParseResult.
//
// It replaces the inner loop of parseASTFiles, adding:
//  1. File-level import map construction (buildFileImports).
//  2. parseFieldWithResolver instead of parseField (threads imports + resolver).
//
// And rename parseASTFiles to parseASTFilesWithResolver, adding resolver *TypeResolver.
func parseASTFilesWithResolver(
	fset *token.FileSet,
	files []*ast.File,
	tagKey string,
	resolver *TypeResolver,
) (*ParseResult, error) {
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

	primitiveAliases := collectPrimitiveAliases(files)

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
					fileImports := buildFileImports(file)
					fieldInfos := parseFieldWithResolver(field, typeName, filePath, tagKey, fileImports, resolver)
					for i := range fieldInfos {
						resolveTypeCategoryWithAliases(&fieldInfos[i], primitiveAliases)
					}
					structInfo.Fields = append(structInfo.Fields, fieldInfos...)
				}

				result.Structs[typeName] = structInfo
			}
		}
	}

	return result, nil
}

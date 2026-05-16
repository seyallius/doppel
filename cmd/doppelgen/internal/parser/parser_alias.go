// Package parser. parser_alias.go - Adds a pre-pass to collect type alias declarations
// and map them to their underlying primitive types. During field categorization,
// now check this map first to ensure aliases are treated as primitives.
//
// Supported alias patterns:
//   - type MyString string          → CatPrimitive
//   - type MyInt int64              → CatPrimitive
//   - type MyBool bool              → CatPrimitive
//   - type MyStruct SomeOtherStruct → CatStruct (unchanged)
//
// Note: This only resolves aliases within the same package. Cross-package alias resolution
// would require full type-checking via go/types, which is out of scope for this lightweight AST parser.
package parser

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/seyallius/doppel/cmd/doppelgen/internal/types"
)

// -------------------------------------------- Types --------------------------------------------

// primitiveAliasMap holds a mapping from type alias names to their underlying primitive type.
// Populated during the pre-pass over type declarations in a package.
//
// Example: {"StepStatus": "string", "UserID": "int64"}
type primitiveAliasMap map[string]string

// -------------------------------------------- Public Functions --------------------------------------------

// collectPrimitiveAliases scans all type declarations in the given AST files and returns
// a map of alias name → underlying primitive type for aliases that wrap a builtin primitive.
//
// This is a lightweight pre-pass that enables correct categorization of fields like:
//
//	type User struct {
//	    Status StepStatus  // StepStatus is `type StepStatus string`
//	}
//
// Without this, StepStatus would be treated as a struct and emit invalid v.Clone() code.
func collectPrimitiveAliases(files []*ast.File) primitiveAliasMap {
	aliases := make(primitiveAliasMap)
	for _, file := range files {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				aliasName := typeSpec.Name.Name
				underlying := formatType(typeSpec.Type)
				if isBuiltinPrimitive(underlying) {
					aliases[aliasName] = underlying
				}
			}
		}
	}
	return aliases
}

// -------------------------------------------- Internal Helpers --------------------------------------------

// resolveTypeCategoryWithAliases is a drop-in replacement for resolveTypeCategory.
// It adds support for primitive type aliases by checking the aliases map first.
//
// Priority order:
//  1. Pointer types (*T, *pkg.T)
//  2. Composite types ([]T, map[K]V)
//  3. Primitive type aliases (via aliases map)
//  4. Builtin primitives
//  5. Local struct identifiers
//  6. Cross-package selector types (pkg.TypeName)
//  7. Interface types
//  8. Unknown
func resolveTypeCategoryWithAliases(fi *types.FieldInfo, aliases primitiveAliasMap) {
	t := fi.Type

	// ── Pointer types: *T, *pkg.T ─────────────────────────────────────────
	if strings.HasPrefix(t, "*") {
		inner := t[1:]
		switch {
		case isBuiltinPrimitive(inner):
			fi.TypeCategory = types.CatPtrPrimitive
			fi.PointedToType = inner
		case isValidIdent(inner):
			// Check if it's a primitive alias first!
			if underlying, ok := aliases[inner]; ok {
				fi.TypeCategory = types.CatPtrPrimitive
				fi.PointedToType = underlying
			} else {
				fi.TypeCategory = types.CatPtrStruct
				fi.PointedToType = inner
			}
		case isSelectorType(inner):
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

	// ── Primitive type aliases (NEW!) ─────────────────────────────────────
	if underlying, ok := aliases[t]; ok {
		fi.TypeCategory = types.CatPrimitive
		fi.UnderlyingType = underlying
		return
	}

	// ── Builtin primitive value types ─────────────────────────────────────
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
			fi.TypeCategory = types.CatStruct
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

// Package types. types.go - Defines the intermediate representation (IR) data structures
// used by the doppelgen code generator. These types bridge AST parsing and code emission.
package types

import "github.com/seyallius/doppel/core"

// -------------------------------------------- Types, Variables & Constants --------------------------------------------

// FieldInfo describes a single struct field and its cloning strategy as determined by the doppel tag and Go type analysis.
type FieldInfo struct {
	Name          string            // field name (e.g., "Tags")
	Type          string            // Go type expression (e.g., "[]string", "*Address")
	Tag           string            // raw tag value (e.g., "deep", "shallow", "-", "")
	Doc           string            // line comment or doc comment on the field
	File          string            // source file where the field was declared
	Directive     core.TagDirective // Resolved directive from parsing the tag.
	TypeCategory  TypeCategory      // TypeCategory is the resolved category of the field's type. This drives which manual helper (or clone pattern) to emit.
	ElemType      string            // Element type for slices and maps (e.g., "string" for []string).
	KeyType       string            // Key type for maps (e.g., "string" for map[string]int).
	ValueType     string            // Value type for maps (e.g., "int" for map[string]int).
	PointedToType string            // PointedToType is the dereferenced type for pointer fields (e.g., "Address" for *Address).
}

// TypeCategory classifies a field's Go type for clone strategy selection.
type TypeCategory int

const (
	CatPrimitive    TypeCategory = iota // CatPrimitive is a value type that needs no deep copy (int, string, bool, float, etc.).
	CatSlice                            // CatSlice is a slice type ([]T).
	CatMap                              // CatMap is a map type (map[K]V).
	CatPtrPrimitive                     // CatPtrPrimitive is a pointer to a primitive type (*int, *string, etc.).
	CatPtrStruct                        // CatPtrStruct is a pointer to a struct type (*Address, *User, etc.).
	CatStruct                           // CatStruct is a non-pointer struct type (embedded or by value).
	CatInterface                        // CatInterface is an interface type.
	CatUnknown                          // CatUnknown is a type that could not be classified.
)

// StructInfo describes a struct type eligible for Clone() generation, including metadata, fields, and skip status.
type StructInfo struct {
	Name       string      // type name (e.g., "User")
	File       string      // source file path
	Package    string      // package name
	Doc        string      // doc comment on the struct
	Fields     []FieldInfo // ordered field list
	Skip       bool        // true if generation should be skipped
	SkipReason string      // reason for skipping (e.g., "has existing Clone()", "doppel:skip-gen comment")
}

// TypeInfo maps type names to their StructInfo, used during parsing and dependency resolution.
type TypeInfo map[string]*StructInfo

// ImportSpec represents a Go import needed by the generated file.
type ImportSpec struct {
	Path  string // full import path (e.g., "github.com/seyallius/doppel/core")
	Alias string // optional alias (empty if not needed)
}

// GeneratorConfig holds the CLI configuration for a single generation run.
type GeneratorConfig struct {
	TypeNames []string // specific types to generate (empty = all tagged structs)
	Package   string   // target package name
	Output    string   // output directory
	Preview   bool     // print to stdout without writing files
	Tag       string   // tag key (default: "doppel")
}

// GenerationUnit represents a single file to be generated, containing one or more Clone() methods.
type GenerationUnit struct {
	FileName string        // output filename (e.g., "user_clone.gen.go")
	Package  string        // package name
	Imports  []ImportSpec  // required imports
	Structs  []*StructInfo // structs to generate Clone() for
	Source   string        // original source file name for the header
}

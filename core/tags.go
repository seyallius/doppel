// Package core. tags - defines struct tag constants and a lightweight tag parser
// for the doppel code generator.
//
// These types have NO runtime reflection — they are pure data structures that
// the generator uses to decide how to emit Clone() implementations.
//
// Supported tag directives:
//
//	doppel:"-"        Skip the field entirely; the clone receives the zero value.
//	doppel:"shallow"  Assign without recursing; the clone shares the field's value.
//	doppel:"clone"    Emit a deep clone for this field (generator will produce custom logic).
//	doppel:"deep"     Explicit deep copy (default behavior).
//	doppel:"empty"    Produce empty-but-non-nil value for collections and pointers.
//
// Usage in struct definitions:
//
//	type User struct {
//	    Name    string
//	    Secret  string           `doppel:"-"`
//	    Config  map[string]string `doppel:"shallow"`
//	    Address *Address         `doppel:"clone"`
//	    Tags    []string         `doppel:"deep"`
//	    Cache   []string         `doppel:"empty"`
//	}
package core

// -------------------------------------------- Types, Variables & Constants --------------------------------------------

// TagKey is the struct tag key consulted by the code generator.
const TagKey = "doppel"

// TagValue is a type-safe alias for doppel struct tag directives.
// Using this alias in switch statements prevents accidental typos.
type TagValue string

const (
	// TagSkip omits the field from the clone; the field receives its zero value.
	// For pointers, slices, and maps, the zero value is nil.
	TagSkip TagValue = "-"

	// TagShallow copies the field by direct assignment; the clone shares
	// the original's backing data. Use for large, immutable data.
	TagShallow TagValue = "shallow"

	// TagClone requires a user-provided clone function named clone<Type><Field>.
	// The generator emits a call to this function for the field.
	TagClone TagValue = "clone"

	// TagDeep performs a full recursive deep copy via manual.* helpers
	// or nested .Clone() calls. This is the default behavior when no tag
	// is specified or when the tag value is empty/unrecognized.
	TagDeep TagValue = "deep"

	// TagEmpty produces an empty-but-non-nil value for non-primitive types.
	//   - []T → []T{}
	//   - map[K]V → map[K]V{}
	//   - *Struct → &Struct{}
	// Primitive types ignore this tag (assignment is already non-nil).
	TagEmpty TagValue = "empty"
)

// TagDirective represents the parsed result of a doppel struct tag.
// Each boolean field is mutually exclusive — at most one will be true.
type TagDirective struct {
	Skip    bool // doppel:"-" — exclude from clone
	Shallow bool // doppel:"shallow" — shared reference
	Deep    bool // doppel:"deep" or default — full deep copy
	Clone   bool // doppel:"clone" — user-provided clone function
	Empty   bool // doppel:"empty" — empty-but-non-nil for collections/pointers
}

// -------------------------------------------- Public API --------------------------------------------

// ParseTagValue validates and converts a raw tag string to TagValue.
// Returns TagDeep as default for empty or unrecognized values.
func ParseTagValue(raw string) TagValue {
	switch TagValue(raw) {
	case TagSkip, TagShallow, TagClone, TagDeep, TagEmpty:
		return TagValue(raw)
	default:
		return TagDeep
	}
}

// ParseTag parses a doppel struct tag value into a TagDirective.
// This is a simple string-matching function with no reflection.
//
// Returns a TagDirective where at most one field is true.
// Empty or unrecognized values default to Deep (standard deep copy).
func ParseTag(tagValue string) TagDirective {
	switch ParseTagValue(tagValue) {
	case TagSkip:
		return TagDirective{Skip: true}
	case TagShallow:
		return TagDirective{Shallow: true}
	case TagClone:
		return TagDirective{Clone: true}
	case TagDeep:
		return TagDirective{Deep: true}
	case TagEmpty:
		return TagDirective{Empty: true}
	default:
		return TagDirective{Deep: true}
	}
}

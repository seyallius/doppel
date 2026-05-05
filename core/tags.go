// Package core. tags - defines struct tag constants and a lightweight tag parser
// for the future doppel code generator.
//
// These types have NO runtime reflection — they are pure data structures that
// the generator will use to decide how to emit Clone() implementations.
//
// Supported tag directives:
//
//	doppel:"-"        Skip the field entirely; the clone receives the zero value.
//	doppel:"shallow"  Assign without recursing; the clone shares the field's value.
//	doppel:"readonly" Same as shallow; communicates that the field is conceptually immutable.
//	doppel:"clone"    Emit a deep clone for this field (generator will produce custom logic).
//	doppel:"deep"     Explicit deep copy (same as default, for documentation purposes).
//
// Usage in struct definitions:
//
//	type User struct {
//	    Name    string
//	    Secret  string           `doppel:"-"`
//	    Config  map[string]string `doppel:"readonly"`
//	    Address *Address         `doppel:"clone"`
//	    Tags    []string         `doppel:"deep"`
//	}
//
// TODO: Generator integration — a future code generator will parse these tags
// and emit SelfClonable[T].Clone() implementations automatically.
package core

// -------------------------------------------- Types, Variables & Constants --------------------------------------------

// TagKey is the struct tag key consulted by the future code generator.
const TagKey = "doppel"

// TagDirective represents the parsed result of a doppel struct tag.
// Each field is mutually exclusive — at most one will be true.
type TagDirective struct {
	Skip    bool // doppel:"-" — exclude from clone
	Shallow bool // doppel:"shallow" or doppel:"readonly" — shared reference
	Deep    bool // doppel:"deep" or default — full deep copy
	Clone   bool // doppel:"clone" — generator emits custom clone logic
}

// -------------------------------------------- Public API --------------------------------------------

// ParseTag parses a doppel struct tag value into a TagDirective.
// This is a simple string-matching function with no reflection.
//
// Returns a TagDirective where at most one field is true.
// Empty or unrecognized values default to Deep (standard deep copy).
func ParseTag(tagValue string) TagDirective {
	switch tagValue {
	case "-":
		return TagDirective{Skip: true}
	case "shallow", "readonly":
		return TagDirective{Shallow: true}
	case "clone":
		return TagDirective{Clone: true}
	case "deep":
		return TagDirective{Deep: true}
	default:
		// Empty string or unrecognized — default to deep copy.
		return TagDirective{Deep: true}
	}
}

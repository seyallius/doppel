// Package engine. cycle - defines CyclePolicy and the Options struct that configure
// how Engine handles pointer cycles and shared sub-graph references during deep copy.
//
// Three policies are available:
//
//   - PreserveShared (default): shared pointer allocations in the original are
//     preserved as shared in the clone. Cyclic graphs are reproduced faithfully.
//     This is the same behavior as the Phase 4 engine — fully backward-compatible.
//
//   - BreakCycles: the first visit to a pointer is cloned normally; any back-edge
//     to the SAME allocation is replaced with a nil pointer in the clone. This
//     produces an acyclic clone of a cyclic graph. Shared (non-cyclic) sub-trees
//     are NOT deduplicated — each reference gets its own independent clone.
//
//   - ErrorOnCycle: returns a descriptive error the moment a back-edge (cycle) is
//     detected. Useful for strict validation contexts where a cyclic graph is a bug.
//     Non-cyclic shared references are cloned independently (no deduplication).
//
// Policy comparison:
//
//	Policy          | Cycles       | Shared refs   | Use-case
//	----------------|--------------|---------------|----------------------------------
//	PreserveShared  | reproduced   | deduplicated  | general-purpose (default)
//	BreakCycles     | broken→nil   | independent   | serialization, acyclic output
//	ErrorOnCycle    | error        | independent   | strict validation / assertion
package engine

import "fmt"

// -------------------------------------------- Types, Variables & Constants --------------------------------------------

// CyclePolicy controls how Engine responds when it encounters a pointer or map
// address that has already been visited in the current clone operation.
type CyclePolicy int

const (
	// PreserveShared reproduces the exact sharing topology of the original graph.
	// If two pointers point to the same allocation, their clones do too.
	// Cyclic back-edges resolve to the already-allocated clone node — no infinite
	// recursion, no data loss.
	//
	// This is the default policy and is fully backward-compatible with Phase 4.
	PreserveShared CyclePolicy = iota

	// BreakCycles clones the first visit to a pointer normally. Any subsequent
	// visit to the same address (a back-edge / cycle) is replaced with a nil
	// pointer in the clone. Non-cyclic shared references are cloned independently.
	//
	// Choose BreakCycles when you need an acyclic clone that won't loop but
	// can tolerate nil stubs in place of back-edges.
	BreakCycles

	// ErrorOnCycle returns a *CycleError the moment a back-edge is detected.
	// Non-cyclic shared references are cloned independently (each reference gets
	// its own copy). No deduplication is performed.
	//
	// Choose ErrorOnCycle when a cyclic graph represents a bug in your data model
	// and you want immediate, actionable feedback rather than silent behavior.
	ErrorOnCycle
)

// Options configures an Engine at construction time. All fields have safe zero values.
type Options struct {
	// CyclePolicy controls how cyclic and shared pointer references are handled.
	// The zero value is PreserveShared, which is the Phase 4 default behavior.
	CyclePolicy CyclePolicy
}

// CycleError is returned by Engine.Clone when CyclePolicy is ErrorOnCycle and
// a back-edge is detected. It records the pointer address and the Go type name
// of the cyclic value for easier debugging.
type CycleError struct {
	Addr     uintptr // raw pointer address of the back-edge
	TypeName string  // reflect.Type.String() of the value at that address
}

// -------------------------------------------- Public API --------------------------------------------

// Error implements the error interface for CycleError.
func (e *CycleError) Error() string {
	return fmt.Sprintf(
		"doppel/engine: cycle detected at address 0x%x (type %s); use BreakCycles or PreserveShared policy to handle cyclic graphs",
		e.Addr,
		e.TypeName,
	)
}

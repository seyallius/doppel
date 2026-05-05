// Package manual. map - provides CloneMap[K,V] — a generic helper for creating
// independent deep copies of maps, cloning values while preserving comparable keys.
//
// Key behaviors:
//   - Nil-safety: nil src → (nil, nil); empty non-nil src → fresh empty map.
//   - Independence: cloned map is a new map[K]V; mutations to src never affect the clone.
//   - Key handling: keys are comparable value types in Go and copied automatically; only values are cloned.
//   - Error context: on failure, returns an error annotated with the failing key.
//
// Choose CloneMap when value cloning can fail (e.g., nested structs).
// For primitive value types, pass manual.Identity[V] as the cloneVal function.
package manual

import "fmt"

// CloneMap creates an independent deep copy of src.
//
// Map keys in Go must be comparable. Comparable types are either primitive
// (string, numeric) or structs/arrays of comparable fields. Primitive keys
// are value types, so they do not require their own clone step — they are
// copied automatically during map iteration. CloneMap therefore only accepts
// a value cloner, not a key cloner.
//
// Nil-safety contract:
//   - A nil src returns (nil, nil) — nil is preserved as nil.
//   - An empty (non-nil) src returns a fresh empty map, not nil.
//
// On error, CloneMap returns nil and a contextual error that identifies
// the offending key.
func CloneMap[K comparable, V any](src map[K]V, cloneVal func(V) (V, error)) (map[K]V, error) {
	if src == nil {
		return nil, nil
	}

	dst := make(map[K]V, len(src))

	for key, val := range src {
		cloned, err := cloneVal(val)
		if err != nil {
			return nil, fmt.Errorf("doppel: CloneMap key [%v]: %w", key, err)
		}
		dst[key] = cloned
	}

	return dst, nil
}

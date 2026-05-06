package doppel_test

import "fmt"

func makeStringIntMap(n int) map[string]int {
	m := make(map[string]int, n)
	for idx := 0; idx < n; idx++ {
		m[makeStringSlice(1)[0]] = idx // reuse helper for key uniqueness
		m[key(idx)] = idx * 7
	}
	return m
}

func key(idx int) string {
	return "key_" + string(rune('a'+idx%26)) + "_" + string(rune('0'+idx%10))
}

func intPointer(v int) *int {
	return &v
}

func stringPointer(s string) *string {
	return &s
}

func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

func makeStringSlice(size int) []string {
	slice := make([]string, size)
	for idx := range slice {
		slice[idx] = fmt.Sprintf("element_%d", idx)
	}
	return slice
}

func makeIntSlice(size int) []int {
	slice := make([]int, size)
	for idx := range slice {
		slice[idx] = idx * 3
	}
	return slice
}

// contains reports whether sub is a substring of s.
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

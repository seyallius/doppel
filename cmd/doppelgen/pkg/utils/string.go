// Package utils. string - Provides utility for string.
package utils

import (
	"unicode"
)

// ToUpperAtN given the str, makes the position provided letter to uppercase.
func ToUpperAtN(position int, str string) string {
	if position < 0 || position >= len(str) {
		return str
	}

	runes := []rune(str)
	if position >= len(runes) {
		return str
	}

	runes[position] = unicode.ToUpper(runes[position])
	return string(runes)
}

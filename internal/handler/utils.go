package handler

import "strings"

// Contains checks if a slice has a given string case-insensitively.
func Contains(l []string, s string) bool {
	for _, e := range l {
		if strings.EqualFold(e, s) {
			return true
		}
	}

	return false
}

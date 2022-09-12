package tool

import (
	"regexp"
	"time"
)

// In returns true if a given value exists in the list.
func In[T comparable](value T, list ...T) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}
	return false
}

// Matches return true if a given string value matches regex provided
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Unique returns true if all given values in the list are unique.
func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}

// RFC3339 return true if value is compatible with layout RFC3339
func RFC3339(value string) bool {
	if _, err := time.Parse(time.RFC3339, value); err != nil {
		return false
	}

	return true
}

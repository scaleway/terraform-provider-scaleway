package datasource

import (
	"fmt"
)

// FindExact finds the first element in 'slice' matching the condition defined by 'finder'.
// It returns the first matching element and an error if either no match is found or multiple matches are found.
func FindExact[T any](slice []T, finder func(T) bool, searchName string) (T, error) {
	var found T

	var foundFlag bool

	for _, elem := range slice {
		if finder(elem) {
			if foundFlag {
				// More than one element found with the same search name
				var zero T

				return zero, fmt.Errorf("multiple elements found with the name %s", searchName)
			}

			found = elem
			foundFlag = true
		}
	}

	if !foundFlag {
		var zero T

		return zero, fmt.Errorf("no element found with the name %s", searchName)
	}

	return found, nil
}

package datasource

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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

// SingularDataSourceFindError returns a standard error message for a singular data source's non-nil resource find error.
func SingularDataSourceFindError(resourceType string, err error) error {
	if notFound(err) {
		if errors.Is(err, &TooManyResultsError{}) {
			return fmt.Errorf("multiple %[1]ss matched; use additional constraints to reduce matches to a single %[1]s", resourceType)
		}

		return fmt.Errorf("no matching %[1]s found", resourceType)
	}

	return fmt.Errorf("reading %s: %w", resourceType, err)
}

// notFound returns true if the error represents a "resource not found" condition.
// Specifically, notFound returns true if the error or a wrapped error is of type
// retry.NotFoundError.
func notFound(err error) bool {
	var e *retry.NotFoundError // nosemgrep:ci.is-not-found-error
	return errors.As(err, &e)
}

type TooManyResultsError struct {
	Count       int
	LastRequest interface{}
}

func (e *TooManyResultsError) Error() string {
	return fmt.Sprintf("too many results: wanted 1, got %d", e.Count)
}

func (e *TooManyResultsError) Is(err error) bool {
	_, ok := err.(*TooManyResultsError) //nolint:errorlint // Explicitly does *not* match down the error tree
	return ok
}

func (e *TooManyResultsError) As(target interface{}) bool {
	t, ok := target.(**retry.NotFoundError)
	if !ok {
		return false
	}

	*t = &retry.NotFoundError{
		Message:     e.Error(),
		LastRequest: e.LastRequest,
	}

	return true
}

var ErrTooManyResults = &TooManyResultsError{}

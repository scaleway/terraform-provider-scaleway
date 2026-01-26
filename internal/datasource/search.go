package datasource

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

var (
	ErrMultipleElementsFound = errors.New("multiple elements found")
	ErrNoElementFound       = errors.New("no element found")
	ErrMultipleMatches      = errors.New("multiple matches")
	ErrNoMatchesFound       = errors.New("no matches found")
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

				return zero, fmt.Errorf("%w with the name %s", ErrMultipleElementsFound, searchName)
			}

			found = elem
			foundFlag = true
		}
	}

	if !foundFlag {
		var zero T

		return zero, fmt.Errorf("%w with the name %s", ErrNoElementFound, searchName)
	}

	return found, nil
}

// SingularDataSourceFindError returns a standard error message for a singular data source's non-nil resource find error.
func SingularDataSourceFindError(resourceType string, err error) error {
	if notFound(err) {
		if errors.Is(err, &TooManyResultsError{}) {
			return fmt.Errorf("%w; use additional constraints to reduce matches to a single %[1]s", ErrMultipleMatches, resourceType)
		}

		return fmt.Errorf("%w", ErrNoMatchesFound)
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
	LastRequest any
	Count       int
}

func (e *TooManyResultsError) Error() string {
	return fmt.Sprintf("too many results: wanted 1, got %d", e.Count)
}

func (e *TooManyResultsError) Is(err error) bool {
	_, ok := err.(*TooManyResultsError)

	return ok
}

func (e *TooManyResultsError) As(target any) bool {
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

package meta

import "errors"

// ErrProjectIDNotFound is returned when no project ID can be detected
var ErrProjectIDNotFound = errors.New("could not detect project id")

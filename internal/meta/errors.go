package meta

import "errors"

// ErrProjectIDNotFound is returned when no region can be detected
var ErrProjectIDNotFound = errors.New("could not detect project id")

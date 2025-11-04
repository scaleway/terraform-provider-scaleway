package meta

import "errors"

// ErrProjectIDNotFound is returned when no project ID can be detected
var ErrProjectIDNotFound = errors.New("could not detect project id")

// ErrOrganizationIDNotFound is returned when no organization ID can be detected
var ErrOrganizationIDNotFound = errors.New("could not detect organization id")

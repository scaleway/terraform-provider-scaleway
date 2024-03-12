package regional

import "errors"

// ErrRegionNotFound is returned when no region can be detected
var ErrRegionNotFound = errors.New("could not detect region")

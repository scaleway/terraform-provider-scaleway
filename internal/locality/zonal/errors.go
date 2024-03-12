package zonal

import "errors"

// ErrZoneNotFound is returned when no zone can be detected
var ErrZoneNotFound = errors.New("could not detect zone. Scaleway uses regions and zones. For more information, refer to https://www.terraform.io/docs/providers/scaleway/guides/regions_and_zones.html")

package zonal

import (
	"fmt"
	"strings"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

// ID represents an ID that is linked with a zone, eg fr-par-1/11111111-1111-1111-1111-111111111111
type ID struct {
	ID   string
	Zone scw.Zone
}

func (z ID) String() string {
	return fmt.Sprintf("%s/%s", z.Zone, z.ID)
}

func NewID(zone scw.Zone, id string) ID {
	return ID{
		ID:   id,
		Zone: zone,
	}
}

func ExpandID(id interface{}) ID {
	zonedID := ID{}
	tab := strings.Split(id.(string), "/")

	if len(tab) != 2 {
		zonedID.ID = id.(string)
	} else {
		zone, _ := scw.ParseZone(tab[0])
		zonedID.ID = tab[1]
		zonedID.Zone = zone
	}

	return zonedID
}

// NewIDString constructs a unique identifier based on resource zone and id
func NewIDString(zone scw.Zone, id string) string {
	return fmt.Sprintf("%s/%s", zone, id)
}

// NewNestedIDString constructs a unique identifier based on resource zone, inner and outer IDs
func NewNestedIDString(zone scw.Zone, outerID, innerID string) string {
	return fmt.Sprintf("%s/%s/%s", zone, outerID, innerID)
}

// ParseID parses a zonedID and extracts the resource zone and id.
func ParseID(zonedID string) (zone scw.Zone, id string, err error) {
	rawZone, id, err := locality.ParseLocalizedID(zonedID)
	if err != nil {
		return zone, id, err
	}

	zone, err = scw.ParseZone(rawZone)

	return
}

// ParseNestedID parses a zonedNestedID and extracts the resource zone ,inner and outer ID.
func ParseNestedID(zonedNestedID string) (zone scw.Zone, outerID, innerID string, err error) {
	rawZone, innerID, outerID, err := locality.ParseLocalizedNestedID(zonedNestedID)
	if err != nil {
		return
	}

	zone, err = scw.ParseZone(rawZone)

	return
}

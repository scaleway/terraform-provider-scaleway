package regional

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

// ID represents an ID that is linked with a region, eg fr-par/11111111-1111-1111-1111-111111111111
type ID struct {
	ID     string
	Region scw.Region
}

func NewID(region scw.Region, id string) ID {
	return ID{
		ID:     id,
		Region: region,
	}
}

func NewIDStrings(region scw.Region, ids []string) []string {
	if ids == nil {
		return nil
	}

	flattenedIDs := make([]string, len(ids))
	for i, id := range ids {
		flattenedIDs[i] = NewIDString(region, id)
	}

	return flattenedIDs
}

func (z ID) String() string {
	return fmt.Sprintf("%s/%s", z.Region, z.ID)
}

func ExpandID(id any) ID {
	regionalID := ID{}
	tab := strings.Split(id.(string), "/")

	if len(tab) != 2 {
		regionalID.ID = id.(string)
	} else {
		region, _ := scw.ParseRegion(tab[0])
		regionalID.ID = tab[1]
		regionalID.Region = region
	}

	return regionalID
}

// NewIDString constructs a unique identifier based on resource region and id
func NewIDString(region scw.Region, id string) string {
	return fmt.Sprintf("%s/%s", region, id)
}

// ParseNestedID parses a regionalNestedID and extracts the resource region, inner and outer ID.
func ParseNestedID(regionalNestedID string) (region scw.Region, outerID, innerID string, err error) {
	loc, innerID, outerID, err := locality.ParseLocalizedNestedID(regionalNestedID)
	if err != nil {
		return
	}

	region, err = scw.ParseRegion(loc)

	return
}

// ParseID parses a regionalID and extracts the resource region and id.
func ParseID(regionalID string) (region scw.Region, id string, err error) {
	loc, id, err := locality.ParseLocalizedID(regionalID)
	if err != nil {
		return
	}

	region, err = scw.ParseRegion(loc)

	return
}

func ResolveRegionAndID(
	d *schema.ResourceData,
	fallbackDefaultRegion func(*schema.ResourceData) (scw.Region, error),
) (scw.Region, string, error) {
	if identity, err := d.Identity(); err == nil && identity != nil {
		if v := identity.Get("id"); v != nil {
			ID, _ := v.(string)
			if ID != "" {
				if rv := identity.Get("region"); rv != nil {
					if rstr, ok := rv.(string); ok && rstr != "" {
						return scw.Region(rstr), ID, nil
					}
				}

				if sid := d.Id(); sid != "" {
					if rFromState, _, err := ParseID(sid); err == nil && rFromState != "" {
						return rFromState, ID, nil
					}
				}

				if fallbackDefaultRegion != nil {
					if region, err := fallbackDefaultRegion(d); err == nil && region != "" {
						return region, ID, nil
					}
				}

				return "", "", fmt.Errorf("cannot resolve region for identity (id=%q)", ID)
			}
		}
	}

	if sid := d.Id(); sid != "" {
		region, ID, err := ParseID(sid)
		if err != nil {
			return "", "", err
		}

		if ID == "" {
			return "", "", fmt.Errorf("empty id parsed from state ID %q", sid)
		}

		return region, ID, nil
	}

	return "", "", errors.New("cannot resolve identity: both identity.id and state ID are empty")
}

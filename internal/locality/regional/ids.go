package regional

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	ctx context.Context,
	d *schema.ResourceData,
	fallbackDefaultRegion func(*schema.ResourceData) (scw.Region, error),
) (scw.Region, string, error) {
	identity, err := d.Identity()
	if err != nil {
		tflog.Warn(ctx, fmt.Sprintf("failed to read identity from ResourceData: %v", err))
	} else if identity != nil {
		if v := identity.Get("id"); v != nil {
			id, _ := v.(string)
			if id != "" {
				if rv := identity.Get("region"); rv != nil {
					if rstr, ok := rv.(string); ok && rstr != "" {
						return scw.Region(rstr), id, nil
					}
				}

				if sid := d.Id(); sid != "" {
					regionFromState, _, err := ParseID(sid)
					if err != nil {
						tflog.Warn(ctx, fmt.Sprintf("failed to parse region from state ID %q: %v", sid, err))
					} else if regionFromState != "" {
						return regionFromState, id, nil
					}
				}

				if fallbackDefaultRegion != nil {
					region, err := fallbackDefaultRegion(d)
					if err != nil {
						tflog.Warn(ctx, fmt.Sprintf("fallbackDefaultRegion error for ID %q: %v", id, err))
					} else if region != "" {
						return region, id, nil
					}
				}

				return "", "", fmt.Errorf("cannot resolve region for identity (id=%q)", id)
			}
		}
	}

	sid := d.Id()
	if sid == "" {
		tflog.Error(ctx, "cannot resolve identity: both identity.id and state ID are empty")

		return "", "", errors.New("cannot resolve identity: both identity.id and state ID are empty")
	}

	region, id, err := ParseID(sid)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("failed to parse region/ID from state ID %q: %v", sid, err))

		return "", "", err
	}

	if id == "" {
		tflog.Error(ctx, fmt.Sprintf("empty ID parsed from state ID %q", sid))

		return "", "", fmt.Errorf("empty ID parsed from state ID %q", sid)
	}

	return region, id, nil
}

package regional

import (
	"errors"
	"fmt"
	"strings"

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
		return region, outerID, innerID, err
	}

	region, err = scw.ParseRegion(loc)

	return region, outerID, innerID, err
}

// ParseID parses a regionalID and extracts the resource region and id.
func ParseID(regionalID string) (region scw.Region, id string, err error) {
	loc, id, err := locality.ParseLocalizedID(regionalID)
	if err != nil {
		return region, id, err
	}

	region, err = scw.ParseRegion(loc)

	return region, id, err
}

// ResolveID resolves a regional ID (`region/uuid`) or a bare UUID.
// Precedence: regional ID locality, then regionValue, then the client's default region.
// Prefer a regional ID when present so a missing region attribute still works.
func ResolveID(idValue string, regionValue string, client *scw.Client) (scw.Region, string, error) {
	if idValue == "" {
		return "", "", errors.New("id is empty")
	}

	region, id, err := ParseID(idValue)
	if err == nil {
		return region, id, nil
	}

	if strings.Contains(idValue, "/") {
		return "", "", fmt.Errorf("invalid regional id %q: %w", idValue, err)
	}

	if regionValue != "" {
		parsedRegion, parseErr := scw.ParseRegion(regionValue)
		if parseErr != nil {
			return "", "", parseErr
		}

		return parsedRegion, idValue, nil
	}

	if client == nil {
		return "", "", ErrRegionNotFound
	}

	fallbackRegion, exists := client.GetDefaultRegion()
	if !exists {
		return "", "", ErrRegionNotFound
	}

	return fallbackRegion, idValue, nil
}

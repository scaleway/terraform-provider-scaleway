package meta

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

// ExtractFrameworkRegion resolves the region from a Plugin Framework attribute or the client default.
func ExtractFrameworkRegion(regionAttr types.String, client *scw.Client) (scw.Region, error) {
	if !regionAttr.IsNull() && !regionAttr.IsUnknown() && regionAttr.ValueString() != "" {
		return scw.ParseRegion(regionAttr.ValueString())
	}

	region, exists := client.GetDefaultRegion()
	if exists {
		return region, nil
	}

	return "", regional.ErrRegionNotFound
}

// ExtractFrameworkZone resolves the zone from a Plugin Framework attribute or the client default.
func ExtractFrameworkZone(zoneAttr types.String, client *scw.Client) (scw.Zone, error) {
	if !zoneAttr.IsNull() && !zoneAttr.IsUnknown() && zoneAttr.ValueString() != "" {
		return scw.ParseZone(zoneAttr.ValueString())
	}

	zone, exists := client.GetDefaultZone()
	if exists {
		return zone, nil
	}

	return "", zonal.ErrZoneNotFound
}

// ExtractFrameworkProjectID resolves the project ID from a Plugin Framework attribute or the client default.
func ExtractFrameworkProjectID(projectIDAttr types.String, client *scw.Client) (string, error) {
	if !projectIDAttr.IsNull() && !projectIDAttr.IsUnknown() && projectIDAttr.ValueString() != "" {
		return projectIDAttr.ValueString(), nil
	}

	projectID, exists := client.GetDefaultProjectID()
	if exists {
		return projectID, nil
	}

	return "", ErrProjectIDNotFound
}

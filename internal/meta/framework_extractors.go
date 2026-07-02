package meta

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
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

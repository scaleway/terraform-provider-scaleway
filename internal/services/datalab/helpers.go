package datalab

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resolveRegion(regionAttr types.String, client *scw.Client) (scw.Region, error) {
	if !regionAttr.IsNull() && !regionAttr.IsUnknown() && regionAttr.ValueString() != "" {
		return scw.ParseRegion(regionAttr.ValueString())
	}

	region, exists := client.GetDefaultRegion()
	if exists {
		return region, nil
	}

	return "", errors.New("region is required; set it on the resource or configure a default region on the provider")
}

func resolveProjectID(projectIDAttr types.String, client *scw.Client) (string, error) {
	if !projectIDAttr.IsNull() && !projectIDAttr.IsUnknown() && projectIDAttr.ValueString() != "" {
		return projectIDAttr.ValueString(), nil
	}

	projectID, exists := client.GetDefaultProjectID()
	if exists {
		return projectID, nil
	}

	return "", errors.New("project_id is required; set it on the resource or configure a default project on the provider")
}

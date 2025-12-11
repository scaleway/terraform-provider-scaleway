package regional

import (
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// AllRegions returns all valid Scaleway regions as strings
func AllRegions() []string {
	regions := make([]string, 0, len(scw.AllRegions))
	for _, r := range scw.AllRegions {
		regions = append(regions, r.String())
	}

	return regions
}

// SchemaAttribute returns a Plugin Framework schema attribute for a region field
func SchemaAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Optional:    true,
		Description: "The region you want to attach the resource to",
	}
}

package regional

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

// ComputedSchema returns a standard schema for a region
func ComputedSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: "The region of the resource",
		Computed:    true,
	}
}

func allRegions() []string {
	regions := make([]string, 0, len(scw.AllRegions))
	for _, z := range scw.AllRegions {
		regions = append(regions, z.String())
	}

	return regions
}

// Schema returns a standard schema for a region
func Schema() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeString,
		Description:      "The region you want to attach the resource to",
		Optional:         true,
		ForceNew:         true,
		ValidateDiagFunc: locality.ValidateStringInSliceWithWarning(allRegions(), "region"),
		DiffSuppressFunc: suppressSDKNullAssignment,
	}
}

func suppressSDKNullAssignment(k, old, new string, d *schema.ResourceData) bool {
	return new == "" && old != ""
}

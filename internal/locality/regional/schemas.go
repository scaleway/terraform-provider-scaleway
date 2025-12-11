package regional

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

// ComputedSchema returns a standard schema for a region
func ComputedSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: "The region of the resource",
		Computed:    true,
	}
}

func AllRegions() []string {
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
		ValidateDiagFunc: verify.ValidateStringInSliceWithWarning(AllRegions(), "region"),
		DiffSuppressFunc: locality.SuppressSDKNullAssignment,
	}
}

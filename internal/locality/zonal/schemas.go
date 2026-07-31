package zonal

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

// ComputedSchema returns a standard schema for a zone
func ComputedSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: "The zone of the resource",
		Computed:    true,
	}
}

func AllZones() []string {
	zones := make([]string, 0, len(scw.AllZones))
	for _, z := range scw.AllZones {
		zones = append(zones, z.String())
	}

	return zones
}

// Schema returns a standard schema for a zone
func Schema() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeString,
		Description:      "The zone you want to attach the resource to",
		Optional:         true,
		ForceNew:         true,
		ValidateDiagFunc: verify.ValidateStringInSliceWithWarning(AllZones(), "zone"),
		DiffSuppressFunc: locality.SuppressSDKNullAssignment,
	}
}

// ProviderSchema returns a standard schema for a zone attribute inside the
// provider schema.
// This schema does not define the "Computed" field for compatibility with the
// schema of the provider using the Terraform Framework.
func ProviderSchema() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeString,
		Description:      "The zone you want to attach the resource to",
		Optional:         true,
		ForceNew:         true,
		ValidateDiagFunc: verify.ValidateStringInSliceWithWarning(AllZones(), "zone"),
		DiffSuppressFunc: locality.SuppressSDKNullAssignment,
	}
}

package regional

import (
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
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
func SchemaAttribute(description ...string) schema.StringAttribute {
	desc := "The region you want to attach the resource to"
	if len(description) > 0 {
		desc = description[0]
	}

	return schema.StringAttribute{
		Optional:    true,
		Description: desc,
		Validators: []validator.String{
			verify.IsStringOneOfWithWarning(AllRegions()),
		},
	}
}

package zonal

import (
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

// SchemaAttribute returns a Plugin Framework schema attribute for a zone field
func SchemaAttribute(description string) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:    true,
		Description: description,
		Validators: []validator.String{
			verify.IsStringOneOfWithWarning(AllZones()),
		},
	}
}

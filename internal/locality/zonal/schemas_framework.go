package zonal

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

// SchemaAttributeComputed returns a Plugin Framework resource schema attribute for a
// zone field, with the `Computed` field set to `true`.
func SchemaAttributeComputed(description string) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: description,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
			stringplanmodifier.UseStateForUnknown(),
		},
		Validators: []validator.String{
			verify.IsStringOneOfWithWarning(AllZones()),
		},
	}
}

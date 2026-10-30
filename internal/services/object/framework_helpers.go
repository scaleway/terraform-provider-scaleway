package object

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

// projectIDFrameworkAttribute returns the framework schema attribute for the
// shared project_id field, matching the pattern used across framework resources.
func projectIDFrameworkAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "The project ID the resource belongs to. Defaults to the provider's project ID.",
		Validators: []validator.String{
			verify.IsStringUUID(),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

func normalizePolicyJSON(raw string) (string, error) {
	return structure.NormalizeJsonString(raw)
}

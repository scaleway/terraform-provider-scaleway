package account

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

// ProjectIDSchemaAttribute returns the Framework schema for project_id (Optional + Computed + RequiresReplace).
func ProjectIDSchemaAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "The project_id you want to attach the resource to",
		Description:         "The project_id you want to attach the resource to",
		Validators: []validator.String{
			verify.IsStringUUID(),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

package account

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

// ProjectIDSchemaFw returns the Framework schema for project_id
// (Optional + Computed + RequiresReplace). Description is required.
func ProjectIDSchemaFw(description string) *schema.StringAttribute {
	return &schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: description,
		Description:         description,
		Validators: []validator.String{
			verify.IsStringUUID(),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

// ProjectIDSchemaAttribute is ProjectIDSchemaFw as a value (for Attribute maps).
func ProjectIDSchemaAttribute(description string) schema.StringAttribute {
	return *ProjectIDSchemaFw(description)
}

// ProjectIDSchemaAttributeOptional returns Framework project_id without RequiresReplace
// (Optional + Computed). Use for data sources / filters. Description is required.
func ProjectIDSchemaAttributeOptional(description string) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: description,
		Description:         description,
		Validators: []validator.String{
			verify.IsStringUUID(),
		},
	}
}

// OrganizationIDSchemaAttribute returns Framework organization_id as Computed-only.
// Description is required.
func OrganizationIDSchemaAttribute(description string) schema.StringAttribute {
	return schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: description,
		Description:         description,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

// OrganizationIDSchemaAttributeOptional returns Framework organization_id as Optional + Computed.
// Description is required.
func OrganizationIDSchemaAttributeOptional(description string) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: description,
		Description:         description,
		Validators: []validator.String{
			verify.IsStringUUID(),
		},
	}
}

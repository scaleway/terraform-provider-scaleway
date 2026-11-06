package account

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

const (
	defaultProjectIDDescription      = "The project_id you want to attach the resource to"
	defaultOrganizationIDDescription = "The organization_id you want to attach the resource to"
)

func schemaDescription(defaultDesc string, description ...string) string {
	if len(description) > 0 && description[0] != "" {
		return description[0]
	}

	return defaultDesc
}

// ProjectIDSchemaAttribute returns the Framework schema for project_id
// (Optional + Computed + RequiresReplace). Optional description overrides the default.
func ProjectIDSchemaAttribute(description ...string) schema.StringAttribute {
	desc := schemaDescription(defaultProjectIDDescription, description...)

	return schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: desc,
		Description:         desc,
		Validators: []validator.String{
			verify.IsStringUUID(),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

// ProjectIDSchemaFw is kept for callers that dereference a pointer (block volume).
// Prefer ProjectIDSchemaAttribute.
func ProjectIDSchemaFw(description string) *schema.StringAttribute {
	attr := ProjectIDSchemaAttribute(description)

	return &attr
}

// ProjectIDSchemaAttributeOptional returns Framework project_id without RequiresReplace
// (Optional + Computed). Use for data sources / filters. Optional description overrides the default.
func ProjectIDSchemaAttributeOptional(description ...string) schema.StringAttribute {
	desc := schemaDescription(defaultProjectIDDescription, description...)

	return schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: desc,
		Description:         desc,
		Validators: []validator.String{
			verify.IsStringUUID(),
		},
	}
}

// OrganizationIDSchemaAttribute returns Framework organization_id as Computed-only.
// Optional description overrides the default.
func OrganizationIDSchemaAttribute(description ...string) schema.StringAttribute {
	desc := schemaDescription(defaultOrganizationIDDescription, description...)

	return schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: desc,
		Description:         desc,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

// OrganizationIDSchemaAttributeOptional returns Framework organization_id as Optional + Computed.
// Optional description overrides the default.
func OrganizationIDSchemaAttributeOptional(description ...string) schema.StringAttribute {
	desc := schemaDescription(defaultOrganizationIDDescription, description...)

	return schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: desc,
		Description:         desc,
		Validators: []validator.String{
			verify.IsStringUUID(),
		},
	}
}

package account

import (
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

// OrganizationIDSchema returns a standard schema for a organization_id
func OrganizationIDSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: "The organization_id you want to attach the resource to",
		Computed:    true,
	}
}

func OrganizationIDOptionalSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "ID of organization the resource is associated to.",
	}
}

// ProjectIDSchema returns a standard schema for a project_id
func ProjectIDSchema() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeString,
		Description:      "The project_id you want to attach the resource to",
		Optional:         true,
		ForceNew:         true,
		Computed:         true,
		ValidateDiagFunc: verify.IsUUID(),
	}
}

func ResourceProjectIDSchema(description string) resourceSchema.StringAttribute {
	return resourceSchema.StringAttribute{
		Description: description,
		Optional:    true,
		Validators: []validator.String{
			verify.UUIDValidator{},
		},
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
		MarkdownDescription: description,
	}
}

func DatasourceOrganizationIDSchema(description string) dataSourceSchema.StringAttribute {
	return dataSourceSchema.StringAttribute{
		Description: description,
		Optional:    true,
		Validators: []validator.String{
			verify.UUIDValidator{},
		},
	}
}

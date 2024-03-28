package account

import (
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
		Type:         schema.TypeString,
		Description:  "The project_id you want to attach the resource to",
		Optional:     true,
		ForceNew:     true,
		Computed:     true,
		ValidateFunc: verify.IsUUID(),
	}
}

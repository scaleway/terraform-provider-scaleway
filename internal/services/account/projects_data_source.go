package account

import (
	"context"
	_ "embed"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	accountSDK "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/projects_datasource.md
var projectsDataSourceDescription string

func DataSourceProjects() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceProject().SchemaFunc())
	datasource.AddOptionalFieldsToSchema(dsSchema, "organization_id")

	dsSchema["organization_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Computed:         true,
		Optional:         true,
		Description:      "The ID of the organization",
		ValidateDiagFunc: verify.IsUUID(),
	}
	dsSchema["projects"] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "The list of projects",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "ID of the Project",
				},
				"name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Name of the Project",
				},
				"organization_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Organization ID of the Project",
				},
				"created_at": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Creation date of the Project",
				},
				"updated_at": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Update date of the Project",
				},
				"description": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Description of the Project",
				},
			},
		},
	}

	return &schema.Resource{
		ReadContext: DataSourceAccountProjectsRead,
		Schema:      dsSchema,
		Description: projectsDataSourceDescription,
	}
}

func DataSourceAccountProjectsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	accountAPI := NewProjectAPI(m)

	var orgID *string

	if v, orgIDExists := d.GetOk("organization_id"); orgIDExists {
		orgID = types.ExpandStringPtr(v)
	} else {
		orgID = GetOrganizationID(m, d)
	}

	if orgID == nil {
		return diag.Errorf("organization_id was not specified nor found in the provider configuration")
	}

	res, err := accountAPI.ListProjects(&accountSDK.ProjectAPIListProjectsRequest{
		OrganizationID: *orgID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uuid.New().String())
	_ = d.Set("projects", flattenProjects(res.Projects))
	_ = d.Set("organization_id", orgID)

	return nil
}

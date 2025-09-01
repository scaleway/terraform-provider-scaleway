package account

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	accountSDK "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceProject() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceProject().Schema)
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "organization_id")

	dsSchema["name"].ConflictsWith = []string{"project_id"}
	dsSchema["project_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Computed:         true,
		Optional:         true,
		Description:      "The ID of the project",
		ValidateDiagFunc: verify.IsUUID(),
	}

	return &schema.Resource{
		ReadContext: DataSourceAccountProjectRead,
		Schema:      dsSchema,
	}
}

func DataSourceAccountProjectRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	accountAPI := NewProjectAPI(m)

	var projectID string

	if name, nameExists := d.GetOk("name"); nameExists {
		orgID := GetOrganizationID(m, d)
		if orgID == nil {
			// required not in schema as we could use default
			return diag.Errorf("organization_id is required with name")
		}

		res, err := accountAPI.ListProjects(&accountSDK.ProjectAPIListProjectsRequest{
			OrganizationID: *orgID,
			Name:           types.ExpandStringPtr(name),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundProject, err := datasource.FindExact(
			res.Projects,
			func(s *accountSDK.Project) bool { return s.Name == name.(string) },
			name.(string),
		)
		if err != nil {
			return diag.FromErr(err)
		}

		projectID = foundProject.ID
	} else {
		extractedProjectID, _, err := meta.ExtractProjectID(d, m)
		if err != nil {
			return diag.FromErr(err)
		}

		projectID = extractedProjectID
	}

	d.SetId(projectID)
	_ = d.Set("project_id", projectID)

	diags := resourceAccountProjectRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read account project")...)
	}

	if d.Id() == "" {
		return diag.Errorf("account project (%s) not found", projectID)
	}

	return nil
}

func DataSourceProjects() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceProject().Schema)
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

func flattenProjects(projects []*accountSDK.Project) []map[string]any {
	flattenedProjects := make([]map[string]any, len(projects))
	for i, project := range projects {
		flattenedProjects[i] = map[string]any{
			"id":              project.ID,
			"name":            project.Name,
			"organization_id": project.OrganizationID,
			"created_at":      types.FlattenTime(project.CreatedAt),
			"updated_at":      types.FlattenTime(project.UpdatedAt),
			"description":     project.Description,
		}
	}

	return flattenedProjects
}

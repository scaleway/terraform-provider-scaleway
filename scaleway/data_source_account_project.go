package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	accountV2 "github.com/scaleway/scaleway-sdk-go/api/account/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayAccountProject() *schema.Resource {
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayAccountProject().Schema)
	addOptionalFieldsToSchema(dsSchema, "name", "organization_id")

	dsSchema["name"].ConflictsWith = []string{"project_id"}
	dsSchema["project_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Computed:     true,
		Optional:     true,
		Description:  "The ID of the SSH key",
		ValidateFunc: validationUUID(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayAccountProjectRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayAccountProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	accountAPI := accountV2API(meta)

	var projectID string

	if name, nameExists := d.GetOk("name"); nameExists {
		orgID := getOrganizationID(meta, d)
		if orgID == nil {
			// required not in schema as we could use default
			return diag.Errorf("organization_id is required with name")
		}
		res, err := accountAPI.ListProjects(&accountV2.ListProjectsRequest{
			OrganizationID: *orgID,
			Name:           expandStringPtr(name),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		for _, project := range res.Projects {
			if project.Name == name {
				if projectID != "" {
					return diag.Errorf("more than 1 project found with the same name %s", d.Get("name"))
				}
				projectID = project.ID
			}
		}
		if projectID == "" {
			diag.Errorf("no project found with the name %s", d.Get("name"))
		}
	} else {
		extractedProjectID, _, err := extractProjectID(d, meta.(*Meta))
		if err != nil {
			return diag.FromErr(err)
		}

		projectID = extractedProjectID
	}

	d.SetId(projectID)
	_ = d.Set("project_id", projectID)

	diags := resourceScalewayAccountProjectRead(ctx, d, meta)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read account project")...)
	}

	if d.Id() == "" {
		return diag.Errorf("account project (%s) not found", projectID)
	}

	return nil
}

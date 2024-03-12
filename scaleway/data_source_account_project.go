package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	accountV3 "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
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

func dataSourceScalewayAccountProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	accountAPI := accountV3ProjectAPI(m.(*meta.Meta))

	var projectID string

	if name, nameExists := d.GetOk("name"); nameExists {
		orgID := getOrganizationID(m.(*meta.Meta), d)
		if orgID == nil {
			// required not in schema as we could use default
			return diag.Errorf("organization_id is required with name")
		}
		res, err := accountAPI.ListProjects(&accountV3.ProjectAPIListProjectsRequest{
			OrganizationID: *orgID,
			Name:           expandStringPtr(name),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundProject, err := findExact(
			res.Projects,
			func(s *accountV3.Project) bool { return s.Name == name.(string) },
			name.(string),
		)
		if err != nil {
			return diag.FromErr(err)
		}

		projectID = foundProject.ID
	} else {
		extractedProjectID, _, err := meta.ExtractProjectID(d, m.(*meta.Meta))
		if err != nil {
			return diag.FromErr(err)
		}

		projectID = extractedProjectID
	}

	d.SetId(projectID)
	_ = d.Set("project_id", projectID)

	diags := resourceScalewayAccountProjectRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read account project")...)
	}

	if d.Id() == "" {
		return diag.Errorf("account project (%s) not found", projectID)
	}

	return nil
}

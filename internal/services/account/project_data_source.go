package account

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	accountSDK "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/project_datasource.md
var projectDataSourceDescription string

func DataSourceProject() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceProject().SchemaFunc())
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
		Description: projectDataSourceDescription,
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

	res, err := accountAPI.GetProject(&accountSDK.ProjectAPIGetProjectRequest{
		ProjectID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	setProjectState(d, res)

	if d.Id() == "" {
		return diag.Errorf("account project (%s) not found", projectID)
	}

	return nil
}

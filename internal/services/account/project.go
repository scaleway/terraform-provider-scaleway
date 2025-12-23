package account

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	accountSDK "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/project.md
var projectDescription string

func ResourceProject() *schema.Resource {
	return &schema.Resource{
		Description:   projectDescription,
		CreateContext: resourceAccountProjectCreate,
		ReadContext:   resourceAccountProjectRead,
		UpdateContext: resourceAccountProjectUpdate,
		DeleteContext: resourceAccountProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		SchemaFunc:    projectSchema,
		Identity:      identity.DefaultProjectID(),
	}
}

func projectSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "The name of the project",
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Description of the project",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the creation of the Project (Format ISO 8601)",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the last update of the Project (Format ISO 8601)",
		},
		"organization_id": {
			Type:             schema.TypeString,
			Description:      "The organization_id you want to attach the resource to",
			Optional:         true,
			ForceNew:         true,
			Computed:         true,
			ValidateDiagFunc: verify.IsUUID(),
		},
	}
}

func resourceAccountProjectCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	accountAPI := NewProjectAPI(m)

	request := &accountSDK.ProjectAPICreateProjectRequest{
		Name:        types.ExpandOrGenerateString(d.Get("name"), "project"),
		Description: d.Get("description").(string),
	}

	if organisationIDRaw, exist := d.GetOk("organization_id"); exist {
		request.OrganizationID = organisationIDRaw.(string)
	}

	res, err := accountAPI.CreateProject(request, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.ID)

	return resourceAccountProjectRead(ctx, d, m)
}

func resourceAccountProjectRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	accountAPI := NewProjectAPI(m)

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

	_ = d.Set("name", res.Name)
	_ = d.Set("description", res.Description)
	_ = d.Set("created_at", types.FlattenTime(res.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(res.UpdatedAt))
	_ = d.Set("organization_id", res.OrganizationID)

	return nil
}

func resourceAccountProjectUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	accountAPI := NewProjectAPI(m)

	req := &accountSDK.ProjectAPIUpdateProjectRequest{
		ProjectID: d.Id(),
	}

	hasChanged := false

	if d.HasChange("name") {
		req.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
		hasChanged = true
	}

	if d.HasChange("description") {
		req.Description = types.ExpandUpdatedStringPtr(d.Get("description"))
		hasChanged = true
	}

	if hasChanged {
		_, err := accountAPI.UpdateProject(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceAccountProjectRead(ctx, d, m)
}

func resourceAccountProjectDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	accountAPI := NewProjectAPI(m)

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		err := accountAPI.DeleteProject(&accountSDK.ProjectAPIDeleteProjectRequest{
			ProjectID: d.Id(),
		}, scw.WithContext(ctx))
		if err != nil && !httperrors.Is404(err) {
			if isProjectNotUsableError(err) {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

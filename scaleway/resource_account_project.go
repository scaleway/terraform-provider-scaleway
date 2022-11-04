package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	accountV2 "github.com/scaleway/scaleway-sdk-go/api/account/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayAccountProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayAccountProjectCreate,
		ReadContext:   resourceScalewayAccountProjectRead,
		UpdateContext: resourceScalewayAccountProjectUpdate,
		DeleteContext: resourceScalewayAccountProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
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
			"organization_id": organizationIDSchema(),
		},
	}
}

func resourceScalewayAccountProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	accountAPI := accountV2API(meta)

	res, err := accountAPI.CreateProject(&accountV2.CreateProjectRequest{
		Name:        expandOrGenerateString(d.Get("name"), "project-"),
		Description: expandStringPtr(d.Get("description").(string)),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.ID)

	return resourceScalewayAccountProjectRead(ctx, d, meta)
}

func resourceScalewayAccountProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	accountAPI := accountV2API(meta)
	res, err := accountAPI.GetProject(&accountV2.GetProjectRequest{
		ProjectID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", res.Name)
	_ = d.Set("description", res.Description)
	_ = d.Set("created_at", flattenTime(res.CreatedAt))
	_ = d.Set("updated_at", flattenTime(res.UpdatedAt))
	_ = d.Set("organization_id", res.OrganizationID)

	return nil
}

func resourceScalewayAccountProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	accountAPI := accountV2API(meta)

	req := &accountV2.UpdateProjectRequest{
		ProjectID: d.Id(),
	}

	hasChanged := false

	if d.HasChange("name") {
		req.Name = expandUpdatedStringPtr(d.Get("name"))
		hasChanged = true
	}
	if d.HasChange("description") {
		req.Description = expandUpdatedStringPtr(d.Get("description"))
		hasChanged = true
	}

	if hasChanged {
		_, err := accountAPI.UpdateProject(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayAccountProjectRead(ctx, d, meta)
}

func resourceScalewayAccountProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	accountAPI := accountV2API(meta)

	err := accountAPI.DeleteProject(&accountV2.DeleteProjectRequest{
		ProjectID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}

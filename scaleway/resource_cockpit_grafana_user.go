package scaleway

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayCockpitGrafanaUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayCockpitGrafanaUserCreate,
		ReadContext:   resourceScalewayCockpitGrafanaUserRead,
		DeleteContext: resourceScalewayCockpitGrafanaUserDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultCockpitTimeout),
			Read:    schema.DefaultTimeout(defaultCockpitTimeout),
			Delete:  schema.DefaultTimeout(defaultCockpitTimeout),
			Default: schema.DefaultTimeout(defaultCockpitTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"login": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The login of the Grafana user",
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_]{2,24}$`), "must have between 2 and 24 alphanumeric characters"),
			},
			"password": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The password of the Grafana user",
			},
			"role": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The role of the Grafana user",
				ValidateFunc: validation.StringInSlice([]string{
					cockpit.GrafanaUserRoleEditor.String(),
					cockpit.GrafanaUserRoleViewer.String(),
				}, false),
			},
			"project_id": projectIDSchema(),
		},
	}
}

func resourceScalewayCockpitGrafanaUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, err := cockpitAPI(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	login := d.Get("login").(string)
	role := cockpit.GrafanaUserRole(d.Get("role").(string))

	_, err = api.WaitForCockpit(&cockpit.WaitForCockpitRequest{
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	grafanaUser, err := api.CreateGrafanaUser(&cockpit.CreateGrafanaUserRequest{
		ProjectID: projectID,
		Login:     login,
		Role:      role,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("password", grafanaUser.Password)
	d.SetId(cockpitIDWithProjectID(projectID, fmt.Sprintf("%d", grafanaUser.ID)))
	return resourceScalewayCockpitGrafanaUserRead(ctx, d, meta)
}

func resourceScalewayCockpitGrafanaUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, projectID, grafanaUserID, err := cockpitAPIResourceIDWithProjectID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.WaitForCockpit(&cockpit.WaitForCockpitRequest{
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	res, err := api.ListGrafanaUsers(&cockpit.ListGrafanaUsersRequest{
		ProjectID: projectID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	var grafanaUser *cockpit.GrafanaUser
	for _, user := range res.GrafanaUsers {
		if user.ID == grafanaUserID {
			grafanaUser = user
			break
		}
	}

	if grafanaUser == nil {
		d.SetId("")
		return nil
	}

	_ = d.Set("login", grafanaUser.Login)
	_ = d.Set("role", grafanaUser.Role)
	_ = d.Set("project_id", projectID)

	return nil
}

func resourceScalewayCockpitGrafanaUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, projectID, grafanaUserID, err := cockpitAPIResourceIDWithProjectID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteGrafanaUser(&cockpit.DeleteGrafanaUserRequest{
		ProjectID:     projectID,
		GrafanaUserID: grafanaUserID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

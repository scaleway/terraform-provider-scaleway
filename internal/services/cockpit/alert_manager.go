package cockpit

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceCockpitAlertManager() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceCockpitAlertManagerCreate,
		ReadContext:   ResourceCockpitAlertManagerRead,
		UpdateContext: ResourceCockpitAlertManagerUpdate,
		DeleteContext: ResourceCockpitAlertManagerDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(DefaultCockpitTimeout),
			Read:    schema.DefaultTimeout(DefaultCockpitTimeout),
			Delete:  schema.DefaultTimeout(DefaultCockpitTimeout),
			Default: schema.DefaultTimeout(DefaultCockpitTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"project_id": account.ProjectIDSchema(),
			"enable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func ResourceCockpitAlertManagerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	enable := d.Get("enable").(bool)

	if enable {
		_, err = api.EnableAlertManager(&cockpit.RegionalAPIEnableAlertManagerRequest{
			ProjectID: projectID,
		})

	} else {
		_, err = api.DisableManagedAlerts(&cockpit.RegionalAPIDisableManagedAlertsRequest{
			ProjectID: projectID,
		}, scw.WithContext(ctx))
	}

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(ResourceCockpitAlertManagerID(region, projectID))
	return ResourceCockpitAlertManagerRead(ctx, d, meta)
}

func ResourceCockpitAlertManagerRead(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func ResourceCockpitAlertManagerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, _, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	if d.HasChange("enable") {
		enable := d.Get("enable").(bool)
		if enable {
			_, err = api.EnableAlertManager(&cockpit.RegionalAPIEnableAlertManagerRequest{
				ProjectID: projectID,
			})

		} else {
			_, err = api.DisableManagedAlerts(&cockpit.RegionalAPIDisableManagedAlertsRequest{
				ProjectID: projectID,
			}, scw.WithContext(ctx))
		}

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceCockpitAlertManagerRead(ctx, d, meta)
}

func ResourceCockpitAlertManagerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, _, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	_, err = api.DisableManagedAlerts(&cockpit.RegionalAPIDisableManagedAlertsRequest{
		ProjectID: projectID,
	}, scw.WithContext(ctx))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func ResourceCockpitAlertManagerID(region scw.Region, projectId string) (resourceID string) {
	return fmt.Sprintf("%s/%s/1", region, projectId)
}

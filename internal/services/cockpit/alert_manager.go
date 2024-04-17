package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
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
			"region": regional.Schema(),
		},
	}
}

func ResourceCockpitAlertManagerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, err := NewRegionalAPI(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	enable := d.Get("enable").(bool)
	region := d.Get("region").(string)

	if enable {
		_, err = api.EnableManagedAlerts(&cockpit.RegionalAPIEnableAlertManagerRequest{
			ProjectID: projectID,
			Region:    region,
		}, scw.WithContext(ctx))
	} else {
		_, err = api.DisableManagedAlerts(&cockpit.RegionalAPIDisableManagedAlertsRequest{
			ProjectID: projectID,
		}, scw.WithContext(ctx))
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceCockpitAlertManagerRead(ctx, d, m)
}

func ResourceCockpitAlertManagerRead(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	d.SetId(d.Get("project_id").(string))
	return nil
}

func ResourceCockpitAlertManagerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, err := NewAPI(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	if d.HasChange("enable") {
		enable := d.Get("enable").(bool)
		if enable {
			err = api.EnableManagedAlerts(&cockpit.EnableManagedAlertsRequest{
				ProjectID: projectID,
			}, scw.WithContext(ctx))
		} else {
			err = api.DisableManagedAlerts(&cockpit.DisableManagedAlertsRequest{
				ProjectID: projectID,
			}, scw.WithContext(ctx))
		}

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceCockpitAlertManagerRead(ctx, d, m)
}

func ResourceCockpitAlertManagerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, err := NewAPI(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	err = api.DisableManagedAlerts(&cockpit.DisableManagedAlertsRequest{
		ProjectID: projectID,
	}, scw.WithContext(ctx))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

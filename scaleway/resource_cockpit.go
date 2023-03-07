package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1beta1"
)

func resourceScalewayCockpit() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayCockpitCreate,
		ReadContext:   resourceScalewayCockpitRead,
		DeleteContext: resourceScalewayCockpitDelete,
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
			"project_id": projectIDSchema(),
			"endpoints": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Endpoints",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"metrics_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The metrics URL",
						},
						"logs_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The logs URL",
						},
						"alertmanager_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The alertmanager URL",
						},
						"grafana_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The grafana URL",
						},
					},
				},
			},
		},
	}
}

func resourceScalewayCockpitCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, err := cockpitAPI(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)

	res, err := api.ActivateCockpit(&cockpit.ActivateCockpitRequest{
		ProjectID: projectID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.ProjectID)
	return resourceScalewayCockpitRead(ctx, d, meta)
}

func resourceScalewayCockpitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, err := cockpitAPI(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := waitForCockpit(ctx, api, d.Id(), d.Timeout(schema.TimeoutRead))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("project_id", res.ProjectID)
	_ = d.Set("endpoints", flattenCockpitEndpoints(res.Endpoints))

	return nil
}

func resourceScalewayCockpitDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, err := cockpitAPI(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForCockpit(ctx, api, d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.DeactivateCockpit(&cockpit.DeactivateCockpitRequest{
		ProjectID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForCockpit(ctx, api, d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if !is404Error(err) {
			return diag.FromErr(err)
		}
	}

	return nil
}

package container

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceCron() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceContainerCronCreate,
		ReadContext:   ResourceContainerCronRead,
		UpdateContext: ResourceContainerCronUpdate,
		DeleteContext: ResourceContainerCronDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultContainerCronTimeout),
			Read:    schema.DefaultTimeout(defaultContainerCronTimeout),
			Update:  schema.DefaultTimeout(defaultContainerCronTimeout),
			Delete:  schema.DefaultTimeout(defaultContainerCronTimeout),
			Default: schema.DefaultTimeout(defaultContainerCronTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"container_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Container ID to link with your trigger.",
			},
			"schedule": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: verify.ValidateCronExpression(),
				Description:      "Cron format string, e.g. 0 * * * * or @hourly, as schedule time of its jobs to be created and executed.",
			},
			"args": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Cron arguments as json object to pass through during execution.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Cron job status.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Cron job name",
			},
			"region": regional.Schema(),
		},
	}
}

func ResourceContainerCronCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	jsonObj, err := scw.DecodeJSONObject(d.Get("args").(string), scw.NoEscape)
	if err != nil {
		return diag.FromErr(err)
	}

	containerID := locality.ExpandID(d.Get("container_id").(string))
	schedule := d.Get("schedule").(string)
	req := &container.CreateCronRequest{
		ContainerID: containerID,
		Region:      region,
		Schedule:    schedule,
		Name:        types.ExpandStringPtr(d.Get("name")),
		Args:        &jsonObj,
	}

	res, err := api.CreateCron(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(ctx, fmt.Sprintf("[INFO] Submitted new cron job: %#v", res.Schedule))

	_, err = waitForCron(ctx, api, res.ID, region, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(ctx, "[INFO] cron job ready")

	d.SetId(regional.NewIDString(region, res.ID))

	return ResourceContainerCronRead(ctx, d, m)
}

func ResourceContainerCronRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, containerCronID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cron, err := waitForCron(ctx, api, containerCronID, region, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	args, err := scw.EncodeJSONObject(*cron.Args, scw.NoEscape)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("container_id", regional.NewID(region, cron.ContainerID).String())
	_ = d.Set("schedule", cron.Schedule)
	_ = d.Set("args", args)
	_ = d.Set("status", cron.Status)
	_ = d.Set("name", cron.Name)
	_ = d.Set("region", region)

	return nil
}

func ResourceContainerCronUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, containerCronID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &container.UpdateCronRequest{
		ContainerID: scw.StringPtr(locality.ExpandID(d.Get("container_id"))),
		CronID:      locality.ExpandID(containerCronID),
		Region:      region,
	}

	shouldUpdate := false

	if d.HasChange("schedule") {
		req.Schedule = scw.StringPtr(d.Get("schedule").(string))
		shouldUpdate = true
	}

	if d.HasChange("args") {
		jsonObj, err := scw.DecodeJSONObject(d.Get("args").(string), scw.NoEscape)
		if err != nil {
			return diag.FromErr(err)
		}

		shouldUpdate = true
		req.Args = &jsonObj
	}

	if d.HasChange("name") {
		req.Name = scw.StringPtr(d.Get("name").(string))
		shouldUpdate = true
	}

	if shouldUpdate {
		cron, err := api.UpdateCron(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		tflog.Info(ctx, fmt.Sprintf("[INFO] Updated cron job: %#v", req.Schedule))

		_, err = waitForCron(ctx, api, cron.ID, region, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	tflog.Info(ctx, "[INFO] cron job ready")

	return ResourceContainerCronRead(ctx, d, m)
}

func ResourceContainerCronDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, containerCronID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForCron(ctx, api, containerCronID, region, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.DeleteCron(&container.DeleteCronRequest{
		Region: region,
		CronID: containerCronID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(ctx, "[INFO] cron job deleted")

	return nil
}

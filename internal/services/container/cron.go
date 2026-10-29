package container

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/container/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceCron() *schema.Resource {
	return &schema.Resource{
		CreateContext:      ResourceContainerCronCreate,
		ReadContext:        ResourceContainerCronRead,
		UpdateContext:      ResourceContainerCronUpdate,
		DeleteContext:      ResourceContainerCronDelete,
		DeprecationMessage: "The \"scaleway_container_cron\" resource is deprecated, please use `scaleway_container_trigger` with a cron configuration instead",
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
		SchemaFunc:    cronSchema,
	}
}

func cronSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
	}
}

func ResourceContainerCronCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	containerID := locality.ExpandID(d.Get("container_id").(string))
	schedule := d.Get("schedule").(string)

	timezone, err := time.LoadLocation("Europe/Paris") // Timezone is required for API v1 so we use the same default as the Console
	if err != nil {
		return diag.FromErr(err)
	}

	req := &container.CreateTriggerRequest{
		ContainerID: containerID,
		Region:      region,
		Name:        types.ExpandOrGenerateString(d.Get("name"), "cron"),
		CronConfig: &container.CreateTriggerRequestCronConfig{
			Schedule: schedule,
			Body:     d.Get("args").(string),
			Timezone: timezone.String(),
		},
	}

	res, err := api.CreateTrigger(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(ctx, fmt.Sprintf("[INFO] Submitted new cron job: %#v", res.CronConfig.Schedule))

	_, err = waitForTrigger(ctx, api, res.ID, region, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(ctx, "[INFO] cron job ready")

	d.SetId(regional.NewIDString(region, res.ID))

	return ResourceContainerCronRead(ctx, d, m)
}

func ResourceContainerCronRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, containerCronID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	trigger, err := waitForTrigger(ctx, api, containerCronID, region, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("container_id", regional.NewID(region, trigger.ContainerID).String())
	_ = d.Set("schedule", trigger.CronConfig.Schedule)
	_ = d.Set("args", trigger.CronConfig.Body)
	_ = d.Set("status", trigger.Status)
	_ = d.Set("name", trigger.Name)
	_ = d.Set("region", region)

	return nil
}

func ResourceContainerCronUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, containerCronID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &container.UpdateTriggerRequest{
		TriggerID:  locality.ExpandID(containerCronID),
		Region:     region,
		CronConfig: &container.UpdateTriggerRequestCronConfig{},
	}

	shouldUpdate := false

	if d.HasChange("schedule") {
		req.CronConfig.Schedule = new(d.Get("schedule").(string))
		shouldUpdate = true
	}

	if d.HasChange("args") {
		shouldUpdate = true
		req.CronConfig.Body = new(d.Get("args").(string))
	}

	if d.HasChange("name") {
		req.Name = new(d.Get("name").(string))
		shouldUpdate = true
	}

	if shouldUpdate {
		cron, err := api.UpdateTrigger(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		tflog.Info(ctx, fmt.Sprintf("[INFO] Updated cron job: %#v", req.CronConfig.Schedule))

		_, err = waitForTrigger(ctx, api, cron.ID, region, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	tflog.Info(ctx, "[INFO] cron job ready")

	return ResourceContainerCronRead(ctx, d, m)
}

func ResourceContainerCronDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, containerCronID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForTrigger(ctx, api, containerCronID, region, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.DeleteTrigger(&container.DeleteTriggerRequest{
		Region:    region,
		TriggerID: containerCronID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(ctx, "[INFO] cron job deleted")

	return nil
}

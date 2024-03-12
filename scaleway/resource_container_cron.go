package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/robfig/cron/v3"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func resourceScalewayContainerCron() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayContainerCronCreate,
		ReadContext:   resourceScalewayContainerCronRead,
		UpdateContext: resourceScalewayContainerCronUpdate,
		DeleteContext: resourceScalewayContainerCronDelete,
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
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateCronExpression(),
				Description:  "Cron format string, e.g. 0 * * * * or @hourly, as schedule time of its jobs to be created and executed.",
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

func resourceScalewayContainerCronCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := containerAPIWithRegion(d, m)
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
		Name:        expandStringPtr(d.Get("name")),
		Args:        &jsonObj,
	}

	res, err := api.CreateCron(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(ctx, fmt.Sprintf("[INFO] Submitted new cron job: %#v", res.Schedule))
	_, err = waitForContainerCron(ctx, api, res.ID, region, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}
	tflog.Info(ctx, "[INFO] cron job ready")

	d.SetId(regional.NewIDString(region, res.ID))

	return resourceScalewayContainerCronRead(ctx, d, m)
}

func resourceScalewayContainerCronRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, containerCronID, err := containerAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cron, err := waitForContainerCron(ctx, api, containerCronID, region, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if is404Error(err) {
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

	return nil
}

func resourceScalewayContainerCronUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, containerCronID, err := containerAPIWithRegionAndID(m, d.Id())
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
		_, err = waitForContainerCron(ctx, api, cron.ID, region, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	tflog.Info(ctx, "[INFO] cron job ready")

	return resourceScalewayContainerCronRead(ctx, d, m)
}

func resourceScalewayContainerCronDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, containerCronID, err := containerAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForContainerCron(ctx, api, containerCronID, region, d.Timeout(schema.TimeoutDelete))
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

func validateCronExpression() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of '%s' to be string", k))
			return
		}
		_, err := cron.ParseStandard(v)
		if err != nil {
			es = append(es, fmt.Errorf("'%s' should be an valid Cron expression", k))
		}
		return
	}
}

package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayFunctionCron() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayFunctionCronCreate,
		ReadContext:   resourceScalewayFunctionCronRead,
		UpdateContext: resourceScalewayFunctionCronUpdate,
		DeleteContext: resourceScalewayFunctionCronDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultFunctionCronTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"function_id": {
				Type:        schema.TypeString,
				Description: "The ID of the function to create a cron for.",
				Required:    true,
			},
			"schedule": {
				Type:        schema.TypeString,
				Description: "The schedule of the cron.",
				Required:    true,
			},
			"region": regionSchema(),
		},
	}
}

func resourceScalewayFunctionCronCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := functionAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	cron, err := api.CreateCron(&function.CreateCronRequest{
		FunctionID: d.Get("function_id").(string),
		Schedule:   d.Get("schedule").(string),
		Region:     region,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newRegionalIDString(region, cron.ID))

	return resourceScalewayFunctionCronRead(ctx, d, meta)
}

func resourceScalewayFunctionCronRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := functionAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cron, err := api.WaitForCron(&function.WaitForCronRequest{
		Region: region,
		CronID: id,
	}, scw.WithContext(ctx))

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("function_id", cron.FunctionID)
	_ = d.Set("schedule", cron.Schedule)

	return nil
}

func resourceScalewayFunctionCronUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := functionAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cron, err := api.WaitForCron(&function.WaitForCronRequest{
		Region: region,
		CronID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	req := &function.UpdateCronRequest{
		Region: region,
		CronID: cron.ID,
	}

	if d.HasChange("schedule") {
		req.Schedule = expandStringPtr(d.Get("schedule").(string))
	}

	_, err = api.UpdateCron(req, scw.WithContext(ctx))
	if err != nil {
		return nil
	}

	return resourceScalewayFunctionCronRead(ctx, d, meta)
}

func resourceScalewayFunctionCronDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := functionAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.WaitForCron(&function.WaitForCronRequest{
		Region: region,
		CronID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil
	}

	_, err = api.DeleteCron(&function.DeleteCronRequest{
		Region: region,
		CronID: id,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}

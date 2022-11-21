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
			Read:    schema.DefaultTimeout(defaultFunctionCronTimeout),
			Update:  schema.DefaultTimeout(defaultFunctionCronTimeout),
			Delete:  schema.DefaultTimeout(defaultFunctionCronTimeout),
			Create:  schema.DefaultTimeout(defaultFunctionCronTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"function_id": {
				Type:        schema.TypeString,
				Description: "The ID of the function to create a cron for.",
				Required:    true,
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
				Description: "Functions arguments as json object to pass through during execution.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Cron job status.",
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

	functionID := expandID(d.Get("function_id").(string))
	f, err := waitForFunction(ctx, api, region, functionID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	request := &function.CreateCronRequest{
		FunctionID: f.ID,
		Schedule:   d.Get("schedule").(string),
		Region:     region,
	}

	if args, ok := d.GetOk("args"); ok {
		jsonObj, err := scw.DecodeJSONObject(args.(string), scw.NoEscape)
		if err != nil {
			return diag.FromErr(err)
		}
		request.Args = &jsonObj
	}

	cron, err := api.CreateCron(request, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForFunctionCron(ctx, api, region, cron.ID, d.Timeout(schema.TimeoutCreate))
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

	cron, err := waitForFunctionCron(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("function_id", newRegionalID(region, cron.FunctionID).String())
	_ = d.Set("schedule", cron.Schedule)

	args, err := scw.EncodeJSONObject(*cron.Args, scw.NoEscape)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("args", args)
	_ = d.Set("status", cron.Status)

	return nil
}

func resourceScalewayFunctionCronUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := functionAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cron, err := waitForFunctionCron(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &function.UpdateCronRequest{
		Region: region,
		CronID: cron.ID,
	}

	shouldUpdate := false
	if d.HasChange("schedule") {
		req.Schedule = expandStringPtr(d.Get("schedule").(string))
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

	if shouldUpdate {
		_, err = api.UpdateCron(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayFunctionCronRead(ctx, d, meta)
}

func resourceScalewayFunctionCronDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := functionAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cron, err := waitForFunctionCron(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_, err = api.DeleteCron(&function.DeleteCronRequest{
		Region: region,
		CronID: cron.ID,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}

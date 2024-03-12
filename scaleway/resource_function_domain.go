package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func resourceScalewayFunctionDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayFunctionDomainCreate,
		ReadContext:   resourceScalewayFunctionDomainRead,
		DeleteContext: resourceScalewayFunctionDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultFunctionTimeout),
			Read:    schema.DefaultTimeout(defaultFunctionTimeout),
			Update:  schema.DefaultTimeout(defaultFunctionTimeout),
			Delete:  schema.DefaultTimeout(defaultFunctionTimeout),
			Create:  schema.DefaultTimeout(defaultFunctionTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"function_id": {
				Type:             schema.TypeString,
				Description:      "The ID of the function",
				Required:         true,
				ForceNew:         true,
				ValidateFunc:     validationUUIDorUUIDWithLocality(),
				DiffSuppressFunc: diffSuppressFuncLocality,
			},
			"hostname": {
				Type:        schema.TypeString,
				Description: "The hostname that should be redirected to the function",
				Required:    true,
				ForceNew:    true,
			},
			"url": {
				Type:        schema.TypeString,
				Description: "URL to use to trigger the function",
				Computed:    true,
			},
			"region": regional.Schema(),
		},
		CustomizeDiff: CustomizeDiffLocalityCheck("function_id"),
	}
}

func resourceScalewayFunctionDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := functionAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	functionID := regional.ExpandID(d.Get("function_id").(string)).ID
	_, err = waitForFunction(ctx, api, region, functionID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	hostname := d.Get("hostname").(string)

	req := &function.CreateDomainRequest{
		Region:     region,
		FunctionID: functionID,
		Hostname:   hostname,
	}

	domain, err := retryCreateFunctionDomain(ctx, api, req, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, domain.ID))

	_, err = waitForFunctionDomain(ctx, api, region, domain.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayFunctionDomainRead(ctx, d, meta)
}

func resourceScalewayFunctionDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := functionAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	domain, err := waitForFunctionDomain(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("hostname", domain.Hostname)
	_ = d.Set("function_id", regional.NewIDString(region, domain.FunctionID))
	_ = d.Set("url", domain.URL)
	_ = d.Set("region", region)

	return nil
}

func resourceScalewayFunctionDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := functionAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForFunctionDomain(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return nil
	}

	_, err = api.DeleteDomain(&function.DeleteDomainRequest{
		DomainID: id,
		Region:   region,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForFunctionDomain(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}

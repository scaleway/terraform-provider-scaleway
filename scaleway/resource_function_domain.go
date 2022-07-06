package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
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
				Type:        schema.TypeString,
				Description: "The ID of the function",
				Required:    true,
				ForceNew:    true,
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
			"region": regionSchema(),
		},
	}
}

func resourceScalewayFunctionDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := functionAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &function.CreateDomainRequest{
		Region:     region,
		FunctionID: expandRegionalID(d.Get("function_id").(string)).ID,
		Hostname:   d.Get("hostname").(string),
	}

	domain, err := api.CreateDomain(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newRegionalIDString(region, domain.ID))

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
	_ = d.Set("function_id", domain.FunctionID)
	_ = d.Set("url", domain.URL)

	return nil
}

func resourceScalewayFunctionDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := functionAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForFunctionDomain(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return nil
	}

	_, err = api.DeleteDomain(&function.DeleteDomainRequest{
		DomainID: id,
		Region:   region,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}

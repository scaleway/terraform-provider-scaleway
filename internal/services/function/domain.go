package function

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceFunctionDomainCreate,
		ReadContext:   ResourceFunctionDomainRead,
		DeleteContext: ResourceFunctionDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(DefaultFunctionTimeout),
			Read:    schema.DefaultTimeout(DefaultFunctionTimeout),
			Delete:  schema.DefaultTimeout(DefaultFunctionTimeout),
			Create:  schema.DefaultTimeout(DefaultFunctionTimeout),
		},
		SchemaVersion: 0,
		SchemaFunc:    domainSchema,
		CustomizeDiff: cdf.LocalityCheck("function_id"),
		Identity:      identity.DefaultRegional(),
	}
}

func domainSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"function_id": {
			Type:             schema.TypeString,
			Description:      "The ID of the function",
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
			DiffSuppressFunc: dsf.Locality,
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
	}
}

func ResourceFunctionDomainCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := functionAPIWithRegion(d, m)
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

	err = identity.SetRegionalIdentity(d, region, domain.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDomain(ctx, api, region, domain.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceFunctionDomainRead(ctx, d, m)
}

func ResourceFunctionDomainRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	domain, err := waitForDomain(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("hostname", domain.Hostname)
	_ = d.Set("function_id", regional.NewIDString(region, domain.FunctionID))
	_ = d.Set("url", domain.URL)
	_ = d.Set("region", region)

	err = identity.SetRegionalIdentity(d, region, domain.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ResourceFunctionDomainDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDomain(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if httperrors.Is404(err) {
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

	_, err = waitForDomain(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}

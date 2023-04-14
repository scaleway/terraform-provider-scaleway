package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayTemDomainCheck() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayTemDomainCheckCreate,
		ReadContext:   resourceScalewayTemDomainCheckRead,
		Delete:        schema.RemoveFromState,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Delete:  schema.DefaultTimeout(defaultTemDomainTimeout),
			Default: schema.DefaultTimeout(defaultTemDomainTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"domain_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The domain used when sending emails",
			},
			"is_ready": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the domain check is successful or not",
			},
			"triggers": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				ForceNew:    true,
				Description: "Triggers to check",
			},
			"region": regionSchema(),
		},
	}
}

func resourceScalewayTemDomainCheckCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := temAPIWithRegionAndID(meta, d.Get("domain_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	domain, err := waitForTemDomain(ctx, api, region, id, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	domain, err = api.CheckDomain(&tem.CheckDomainRequest{
		Region:   region,
		DomainID: domain.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newRegionalIDString(region, domain.ID))
	return resourceScalewayTemDomainCheckRead(ctx, d, meta)
}

func resourceScalewayTemDomainCheckRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := temAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	domain, err := waitForTemDomain(ctx, api, region, id, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("domain_id", newRegionalIDString(domain.Region, domain.ID))
	_ = d.Set("is_ready", domain.Status == tem.DomainStatusChecked)
	_ = d.Set("triggers", d.Get("triggers"))
	_ = d.Set("region", string(region))

	return nil
}

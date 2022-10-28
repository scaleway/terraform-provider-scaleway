package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayTemDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayTemDomainCreate,
		ReadContext:   resourceScalewayTemDomainRead,
		DeleteContext: resourceScalewayTemDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultTemDomainTimeout),
			Read:    schema.DefaultTimeout(defaultTemDomainTimeout),
			Delete:  schema.DefaultTimeout(defaultTemDomainTimeout),
			Default: schema.DefaultTimeout(defaultTemDomainTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The domain name used when sending emails",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the domain",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the domain",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date and time of domain's creation (RFC 3339 format)",
			},
			"next_check_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date and time of the next scheduled check (RFC 3339 format)",
			},
			"last_valid_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date and time the domain was last found to be valid (RFC 3339 format)",
			},
			"revoked_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date and time of the revocation of the domain (RFC 3339 format)",
			},
			"last_error": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Error message if the last check failed",
			},
			"spf_config": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Snippet of the SPF record that should be registered in the DNS zone",
			},
			"dkim_config": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DKIM public key, as should be recorded in the DNS zone",
			},
			"statistics": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Description: "Domain's statistics",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"total_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Total number of emails matching the request criteria",
						},
						"new_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of emails still in the `new` transient state (received from the API, not yet processed)",
						},
						"sending_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of emails still in the `sending` transient state (received from the API, not yet in their final status)",
						},
						"sent_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of emails in the final `sent` state (have been delivered to the target mail system)",
						},
						"failed_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of emails in the final `failed` state (refused by the target mail system with a final error status)",
						},
						"canceled_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of emails in the final `canceled` state (canceled by customer's request)",
						},
					},
				},
			},
			"region":     regionSchema(),
			"project_id": projectIDSchema(),
		},
	}
}

func resourceScalewayTemDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := temAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	domain, err := api.CreateDomain(&tem.CreateDomainRequest{
		Region:     region,
		ProjectID:  d.Get("project_id").(string),
		DomainName: d.Get("name").(string),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newRegionalIDString(region, domain.ID))

	return resourceScalewayTemDomainRead(ctx, d, meta)
}

func resourceScalewayTemDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := temAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	domain, err := api.GetDomain(&tem.GetDomainRequest{
		Region:   region,
		DomainID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", domain.Name)
	_ = d.Set("status", domain.Status)
	_ = d.Set("created_at", flattenTime(domain.CreatedAt))
	_ = d.Set("next_check_at", flattenTime(domain.NextCheckAt))
	_ = d.Set("last_valid_at", flattenTime(domain.LastValidAt))
	_ = d.Set("revoked_at", flattenTime(domain.RevokedAt))
	_ = d.Set("last_error", domain.LastError)
	_ = d.Set("spf_config", domain.SpfConfig)
	_ = d.Set("dkim_config", domain.DkimConfig)

	statistics := map[string]interface{}{}
	statistics["total_count"] = domain.Statistics.TotalCount
	statistics["new_count"] = domain.Statistics.NewCount
	statistics["sending_count"] = domain.Statistics.SendingCount
	statistics["sent_count"] = domain.Statistics.SentCount
	statistics["failed_count"] = domain.Statistics.FailedCount
	statistics["canceled_count"] = domain.Statistics.CanceledCount
	_ = d.Set("statistics", []map[string]interface{}{statistics})

	_ = d.Set("region", string(region))
	_ = d.Set("project_id", domain.ProjectID)

	return nil
}

func resourceScalewayTemDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := temAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForTemDomain(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	_, err = api.RevokeDomain(&tem.RevokeDomainRequest{
		Region:   region,
		DomainID: id,
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	_, err = waitForTemDomain(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}

package scaleway

import (
	"context"
	"fmt"

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
			"accept_tos": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
				Description: "Accept the Scaleway Terms of Service",
				ValidateFunc: func(i interface{}, k string) (warnings []string, errors []error) {
					v := i.(bool)
					if !v {
						errors = append(errors, fmt.Errorf("you must accept the Scaleway Terms of Service to use this service"))
						return warnings, errors
					}

					return warnings, errors
				},
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
			"smtp_host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "SMTP host to use to send emails",
			},
			// Port 25
			"smtp_port_unsecure": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "SMTP port to use to send emails",
			},
			// Port 587
			"smtp_port": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "SMTP port to use to send emails over TLS",
			},
			// Port 2587
			"smtp_port_alternative": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "SMTP port to use to send emails over TLS",
			},
			// Port 465
			"smtps_port": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "SMTPS port to use to send emails over TLS Wrapper",
			},
			// Port 2465
			"smtps_port_alternative": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "SMTPS port to use to send emails over TLS Wrapper",
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
		AcceptTos:  d.Get("accept_tos").(bool),
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
	_ = d.Set("accept_tos", true)
	_ = d.Set("status", domain.Status)
	_ = d.Set("created_at", flattenTime(domain.CreatedAt))
	_ = d.Set("next_check_at", flattenTime(domain.NextCheckAt))
	_ = d.Set("last_valid_at", flattenTime(domain.LastValidAt))
	_ = d.Set("revoked_at", flattenTime(domain.RevokedAt))
	_ = d.Set("last_error", domain.LastError)
	_ = d.Set("spf_config", domain.SpfConfig)
	_ = d.Set("dkim_config", domain.DkimConfig)
	_ = d.Set("smtp_host", "smtp.tem.scw.cloud")
	_ = d.Set("smtp_port_unsecure", 25)
	_ = d.Set("smtp_port", 587)
	_ = d.Set("smtp_port_alternative", 2587)
	_ = d.Set("smtps_port", 465)
	_ = d.Set("smtps_port_alternative", 2465)
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

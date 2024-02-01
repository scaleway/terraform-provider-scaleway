package scaleway

import (
	"context"
	"errors"
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
				ValidateFunc: func(i interface{}, k string) (warnings []string, errs []error) {
					v := i.(bool)
					if !v {
						errs = append(errs, errors.New("you must accept the Scaleway Terms of Service to use this service"))
						return warnings, errs
					}

					return warnings, errs
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
			"smtp_port_unsecure": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("SMTP port to use to send emails. (Port %d)", tem.SMTPPortUnsecure),
			},
			"smtp_port": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("SMTP port to use to send emails over TLS. (Port %d)", tem.SMTPPort),
			},
			"smtp_port_alternative": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("SMTP port to use to send emails over TLS. (Port %d)", tem.SMTPPortAlternative),
			},
			"smtps_port": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("SMTPS port to use to send emails over TLS Wrapper. (Port %d)", tem.SMTPSPort),
			},
			"smtps_port_alternative": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("SMTPS port to use to send emails over TLS Wrapper. (Port %d)", tem.SMTPSPortAlternative),
			},
			"mx_blackhole": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Scaleway's blackhole MX server to use",
			},
			"reputation": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The domain's reputation",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of the domain's reputation",
						},
						"score": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "A range from 0 to 100 that determines your domain's reputation score",
						},
						"scored_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Time and date the score was calculated",
						},
						"previous_score": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The previously-calculated domain's reputation score",
						},
						"previous_scored_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Time and date the previous reputation score was calculated",
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
	_ = d.Set("smtp_host", tem.SMTPHost)
	_ = d.Set("smtp_port_unsecure", tem.SMTPPortUnsecure)
	_ = d.Set("smtp_port", tem.SMTPPort)
	_ = d.Set("smtp_port_alternative", tem.SMTPPortAlternative)
	_ = d.Set("smtps_port", tem.SMTPSPort)
	_ = d.Set("smtps_port_alternative", tem.SMTPSPortAlternative)
	_ = d.Set("mx_blackhole", tem.MXBlackhole)
	_ = d.Set("reputation", flattenDomainReputation(domain.Reputation))
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

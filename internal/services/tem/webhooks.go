package tem

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceWebhook() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceWebhookCreate,
		ReadContext:   ResourceWebhookRead,
		DeleteContext: ResourceWebhookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Delete:  schema.DefaultTimeout(DefaultWebhookTimeout),
			Default: schema.DefaultTimeout(DefaultWebhookTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"domain_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The domain id",
				ValidateFunc: validation.IsUUID,
			},
			"region":     regional.Schema(),
			"project_id": account.ProjectIDSchema(),
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(3, 127),
			},
			"event_types": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringNotInSlice([]string{"unknown_type"}, false),
				},
				Description: "List of event types",
				MinItems:    1,
			},
			"sns_arn": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "SNS ARN",
				ValidateFunc: validation.StringLenBetween(3, 127),
			},
			"organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization id",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The webhook id",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update timestamp",
			},
		},
	}
}

//func ResourceWebhookCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
//	api, region, err := temAPIWithRegion(d, m)
//	if err != nil {
//		return diag.FromErr(err)
//	}
//
//	domain, err := api.CreateDomain(&tem.CreateDomainRequest{
//		Region:     region,
//		ProjectID:  d.Get("project_id").(string),
//		DomainName: d.Get("name").(string),
//		AcceptTos:  d.Get("accept_tos").(bool),
//	}, scw.WithContext(ctx))
//	if err != nil {
//		return diag.FromErr(err)
//	}
//
//	d.SetId(regional.NewIDString(region, domain.ID))
//
//	return ResourceWebhookRead(ctx, d, m)
//}
//
//func ResourceWebhookRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
//	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
//	if err != nil {
//		return diag.FromErr(err)
//	}
//
//	domain, err := api.GetDomain(&tem.GetDomainRequest{
//		Region:   region,
//		DomainID: id,
//	}, scw.WithContext(ctx))
//	if err != nil {
//		if httperrors.Is404(err) {
//			d.SetId("")
//			return nil
//		}
//		return diag.FromErr(err)
//	}
//
//	_ = d.Set("name", domain.Name)
//	_ = d.Set("accept_tos", true)
//	_ = d.Set("status", domain.Status)
//	_ = d.Set("created_at", types.FlattenTime(domain.CreatedAt))
//	_ = d.Set("next_check_at", types.FlattenTime(domain.NextCheckAt))
//	_ = d.Set("last_valid_at", types.FlattenTime(domain.LastValidAt))
//	_ = d.Set("revoked_at", types.FlattenTime(domain.RevokedAt))
//	_ = d.Set("last_error", domain.LastError)
//	_ = d.Set("spf_config", domain.SpfConfig)
//	_ = d.Set("dkim_config", domain.DkimConfig)
//	_ = d.Set("dmarc_name", domain.Records.Dmarc.Name)
//	_ = d.Set("dmarc_config", domain.Records.Dmarc.Value)
//	_ = d.Set("smtp_host", tem.SMTPHost)
//	_ = d.Set("smtp_port_unsecure", tem.SMTPPortUnsecure)
//	_ = d.Set("smtp_port", tem.SMTPPort)
//	_ = d.Set("smtp_port_alternative", tem.SMTPPortAlternative)
//	_ = d.Set("smtps_port", tem.SMTPSPort)
//	_ = d.Set("smtps_port_alternative", tem.SMTPSPortAlternative)
//	_ = d.Set("mx_blackhole", tem.MXBlackhole)
//	_ = d.Set("reputation", flattenDomainReputation(domain.Reputation))
//	_ = d.Set("region", string(region))
//	_ = d.Set("project_id", domain.ProjectID)
//	_ = d.Set("smtps_auth_user", domain.ProjectID)
//	return nil
//}
//
//func ResourceWebhookDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
//	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
//	if err != nil {
//		return diag.FromErr(err)
//	}
//
//	_, err = WaitForDomain(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
//	if err != nil {
//		if httperrors.Is404(err) {
//			d.SetId("")
//			return nil
//		}
//
//		return diag.FromErr(err)
//	}
//
//	_, err = api.RevokeDomain(&tem.RevokeDomainRequest{
//		Region:   region,
//		DomainID: id,
//	}, scw.WithContext(ctx))
//	if err != nil && !httperrors.Is404(err) {
//		return diag.FromErr(err)
//	}
//
//	_, err = WaitForDomain(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
//	if err != nil && !httperrors.Is404(err) {
//		return diag.FromErr(err)
//	}
//
//	return nil
//}
